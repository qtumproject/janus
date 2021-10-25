package qtum

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math"
	"math/big"
	"math/rand"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/pkg/errors"
)

var FLAG_GENERATE_ADDRESS_TO = "REGTEST_GENERATE_ADDRESS_TO"
var FLAG_IGNORE_UNKNOWN_TX = "IGNORE_UNKNOWN_TX"
var FLAG_DISABLE_SNIPPING_LOGS = "DISABLE_SNIPPING_LOGS"
var FLAG_HIDE_QTUMD_LOGS = "HIDE_QTUMD_LOGS"

var maximumRequestTime = 10000
var maximumBackoff = (2 * time.Second).Milliseconds()

type Client struct {
	URL  string
	doer doer

	// hex addresses to return for eth_accounts
	Accounts Accounts

	logWriter io.Writer
	logger    log.Logger
	debug     bool

	// is this client using the main network?
	isMain bool

	id      *big.Int
	idStep  *big.Int
	idMutex sync.Mutex

	mutex *sync.RWMutex
	flags map[string]interface{}
}

func ReformatJSON(input []byte) ([]byte, error) {
	var v interface{}
	err := json.Unmarshal([]byte(input), &v)
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(v, "", "  ")
}

func NewClient(isMain bool, rpcURL string, opts ...func(*Client) error) (*Client, error) {
	err := checkRPCURL(rpcURL)
	if err != nil {
		return nil, err
	}

	c := &Client{
		isMain: isMain,
		doer:   http.DefaultClient,
		URL:    rpcURL,
		logger: log.NewNopLogger(),
		debug:  false,
		id:     big.NewInt(0),
		idStep: big.NewInt(1),
		mutex:  &sync.RWMutex{},
		flags:  make(map[string]interface{}),
	}

	for _, opt := range opts {
		if err := opt(c); err != nil {
			return nil, err
		}
	}

	return c, nil
}

func (c *Client) IsMain() bool {
	return c.isMain
}

func (c *Client) Request(method string, params interface{}, result interface{}) error {
	return c.RequestWithContext(nil, method, params, result)
}

func (c *Client) RequestWithContext(ctx context.Context, method string, params interface{}, result interface{}) error {
	req, err := c.NewRPCRequest(method, params)
	if err != nil {
		return errors.WithMessage(err, "couldn't make new rpc request")
	}

	var resp *SuccessJSONRPCResult
	max := int(math.Floor(math.Max(float64(maximumRequestTime/int(maximumBackoff)), 1)))
	for i := 0; i < max; i++ {
		resp, err = c.Do(ctx, req)
		if err != nil {
			if strings.Contains(err.Error(), ErrQtumWorkQueueDepth.Error()) && i != max-1 {
				requestString := marshalToString(req)
				backoffTime := computeBackoff(i, true)
				c.GetLogger().Log("msg", fmt.Sprintf("QTUM process busy, backing off for %f seconds", backoffTime.Seconds()), "request", requestString)
				time.Sleep(backoffTime)
				c.GetLogger().Log("msg", "Retrying QTUM command")
			} else {
				if i != 0 {
					c.GetLogger().Log("msg", fmt.Sprintf("Giving up on QTUM RPC call after %d tries since its busy", i+1))
				}
				return err
			}
		} else {
			break
		}
	}

	err = json.Unmarshal(resp.RawResult, result)
	if err != nil {
		c.GetDebugLogger().Log("method", method, "params", params, "result", result, "error", err)
		return errors.Wrap(err, "couldn't unmarshal response result field")
	}
	return nil
}

func (c *Client) Do(ctx context.Context, req *JSONRPCRequest) (*SuccessJSONRPCResult, error) {
	reqBody, err := json.MarshalIndent(req, "", "  ")
	if err != nil {
		return nil, err
	}

	debugLogger := c.GetDebugLogger()

	debugLogger.Log("method", req.Method)

	if c.IsDebugEnabled() && !c.GetFlagBool(FLAG_HIDE_QTUMD_LOGS) {
		fmt.Fprintf(c.logWriter, "=> qtum RPC request\n%s\n", reqBody)
	}

	respBody, err := c.do(ctx, bytes.NewReader(reqBody))
	if err != nil {
		return nil, errors.Wrap(err, "Client#do")
	}

	if c.IsDebugEnabled() && !c.GetFlagBool(FLAG_HIDE_QTUMD_LOGS) {
		formattedBody, err := ReformatJSON(respBody)
		formattedBodyStr := string(formattedBody)
		if !c.GetFlagBool(FLAG_DISABLE_SNIPPING_LOGS) {
			maxBodySize := 1024 * 8
			if len(formattedBodyStr) > maxBodySize {
				formattedBodyStr = formattedBodyStr[0:maxBodySize/2] + "\n...snip...\n" + formattedBodyStr[len(formattedBody)-maxBodySize/2:]
			}
		}

		if err == nil {
			fmt.Fprintf(c.logWriter, "<= qtum RPC response\n%s\n", formattedBodyStr)
		}
	}

	res, err := c.responseBodyToResult(respBody)
	if err != nil {
		if respBody == nil || len(respBody) == 0 {
			debugLogger.Log("Empty response")
			return nil, errors.Wrap(err, "responseBodyToResult empty response")
		}
		if IsKnownError(err) {
			return nil, err
		}
		if string(respBody) == ErrQtumWorkQueueDepth.Error() {
			// QTUM http server queue depth reached, need to retry
			return nil, ErrQtumWorkQueueDepth
		}
		debugLogger.Log("msg", "Failed to parse response body", "body", string(respBody), "error", err)
		return nil, errors.Wrap(err, "responseBodyToResult")
	}

	return res, nil
}

func (c *Client) NewRPCRequest(method string, params interface{}) (*JSONRPCRequest, error) {
	paramsJSON, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	c.idMutex.Lock()
	c.id = c.id.Add(c.id, c.idStep)
	c.idMutex.Unlock()

	return &JSONRPCRequest{
		JSONRPC: RPCVersion,
		ID:      json.RawMessage(`"` + c.id.String() + `"`),
		Method:  method,
		Params:  paramsJSON,
	}, nil
}

func (c *Client) do(ctx context.Context, body io.Reader) ([]byte, error) {
	var req *http.Request
	var err error
	if ctx != nil {
		req, err = http.NewRequestWithContext(ctx, http.MethodPost, c.URL, body)
	} else {
		req, err = http.NewRequest(http.MethodPost, c.URL, body)
	}
	if err != nil {
		return nil, err
	}

	req.Close = true

	resp, err := c.doer.Do(req)
	if err != nil {
		return nil, err
	}
	defer func() {
		if resp != nil {
			io.Copy(ioutil.Discard, resp.Body)
			resp.Body.Close()
		}
	}()

	reader, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(err, "ioutil error in qtum client package")
	}
	return reader, nil
}

func (c *Client) SetFlag(key string, value interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.setFlagImpl(key, value)
}

func (c *Client) setFlagImpl(key string, value interface{}) {
	c.flags[key] = value
}

func (c *Client) GetFlag(key string) interface{} {
	c.mutex.RLock()
	defer c.mutex.RUnlock()
	return c.getFlagImpl(key)
}

func (c *Client) getFlagImpl(key string) interface{} {
	return c.flags[key]
}

func (c *Client) GetFlagString(key string) *string {
	value := c.GetFlag(key)
	if value == nil {
		return nil
	}
	result := fmt.Sprintf("%v", value)
	return &result
}

func (c *Client) GetFlagBool(key string) bool {
	value := c.GetFlag(key)
	if value == nil {
		return false
	}
	result, ok := value.(bool)
	if !ok {
		return false
	}
	return result
}

type doer interface {
	Do(*http.Request) (*http.Response, error)
}

func SetDoer(d doer) func(*Client) error {
	return func(c *Client) error {
		c.doer = d
		return nil
	}
}

func SetDebug(debug bool) func(*Client) error {
	return func(c *Client) error {
		c.debug = debug
		return nil
	}
}

func SetLogWriter(logWriter io.Writer) func(*Client) error {
	return func(c *Client) error {
		c.logWriter = logWriter
		return nil
	}
}

func SetLogger(l log.Logger) func(*Client) error {
	return func(c *Client) error {
		c.logger = log.WithPrefix(l, "component", "qtum.Client")
		return nil
	}
}

func SetAccounts(accounts Accounts) func(*Client) error {
	return func(c *Client) error {
		c.Accounts = accounts
		return nil
	}
}

func SetGenerateToAddress(address string) func(*Client) error {
	return func(c *Client) error {
		if address != "" {
			c.SetFlag(FLAG_GENERATE_ADDRESS_TO, address)
		}
		return nil
	}
}

func SetIgnoreUnknownTransactions(ignore bool) func(*Client) error {
	return func(c *Client) error {
		c.SetFlag(FLAG_IGNORE_UNKNOWN_TX, ignore)
		return nil
	}
}

func SetDisableSnippingQtumRpcOutput(disable bool) func(*Client) error {
	return func(c *Client) error {
		c.SetFlag(FLAG_DISABLE_SNIPPING_LOGS, !disable)
		return nil
	}
}

func SetHideQtumdLogs(hide bool) func(*Client) error {
	return func(c *Client) error {
		c.SetFlag(FLAG_HIDE_QTUMD_LOGS, hide)
		return nil
	}
}

func (c *Client) GetLogWriter() io.Writer {
	return c.logWriter
}

func (c *Client) GetLogger() log.Logger {
	return c.logger
}

func (c *Client) GetDebugLogger() log.Logger {
	if !c.IsDebugEnabled() {
		return log.NewNopLogger()
	}
	return log.With(level.Debug(c.logger))
}

func (c *Client) GetErrorLogger() log.Logger {
	return log.With(level.Error(c.logger))
}

func (c *Client) IsDebugEnabled() bool {
	return c.debug
}

func (c *Client) responseBodyToResult(body []byte) (*SuccessJSONRPCResult, error) {
	var res *JSONRPCResult
	if err := json.Unmarshal(body, &res); err != nil {
		return nil, err
	}
	if res.Error != nil {
		knownError := res.Error.TryGetKnownError()
		if knownError != res.Error {
			c.GetDebugLogger().Log("msg", fmt.Sprintf("Got error code %d with message '%s' mapped to %s", res.Error.Code, res.Error.Message, knownError.Error()))
		}
		return nil, knownError
	}

	return &SuccessJSONRPCResult{
		ID:        res.ID,
		RawResult: res.RawResult,
		JSONRPC:   res.JSONRPC,
	}, nil
}

func computeBackoff(i int, random bool) time.Duration {
	i = int(math.Min(float64(i), 10))
	randomNumberMilliseconds := 0
	if random {
		randomNumberMilliseconds = rand.Intn(500) - 250
	}
	exponentialBase := math.Pow(2, float64(i)) * 0.25
	exponentialBaseInSeconds := time.Duration(exponentialBase*float64(time.Second)) + time.Duration(randomNumberMilliseconds)*time.Millisecond
	backoffTimeInMilliseconds := math.Min(float64(exponentialBaseInSeconds.Milliseconds()), float64(maximumBackoff))
	return time.Duration(backoffTimeInMilliseconds * float64(time.Millisecond))
}

func checkRPCURL(u string) error {
	if u == "" {
		return errors.New("URL must be set")
	}

	qtumRPC, err := url.Parse(u)
	if err != nil {
		return errors.Errorf("QTUM_RPC URL: %s", u)
	}

	if qtumRPC.User == nil {
		return errors.Errorf("QTUM_RPC URL (must specify user & password): %s", u)
	}

	return nil
}
