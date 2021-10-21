package server

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sync"

	"github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/labstack/echo"
	"github.com/labstack/echo/middleware"
	"github.com/pkg/errors"
	"github.com/qtumproject/janus/pkg/eth"
	"github.com/qtumproject/janus/pkg/qtum"
	"github.com/qtumproject/janus/pkg/transformer"
)

type Server struct {
	address       string
	transformer   *transformer.Transformer
	qtumRPCClient *qtum.Qtum
	logWriter     io.Writer
	logger        log.Logger
	httpsKey      string
	httpsCert     string
	debug         bool
	mutex         *sync.Mutex
	echo          *echo.Echo
}

func New(
	qtumRPCClient *qtum.Qtum,
	transformer *transformer.Transformer,
	addr string,
	opts ...Option,
) (*Server, error) {
	p := &Server{
		logger:        log.NewNopLogger(),
		echo:          echo.New(),
		address:       addr,
		qtumRPCClient: qtumRPCClient,
		transformer:   transformer,
	}

	var err error
	for _, opt := range opts {
		if err = opt(p); err != nil {
			return nil, err
		}
	}

	return p, nil
}

func (s *Server) Start() error {
	logWriter := s.logWriter
	e := s.echo
	e.Use(middleware.CORS())
	e.Use(middleware.BodyDump(func(c echo.Context, req []byte, res []byte) {
		myctx := c.Get("myctx")
		cc, ok := myctx.(*myCtx)
		if !ok {
			return
		}

		if s.debug {
			reqBody, err := qtum.ReformatJSON(req)
			resBody, err := qtum.ReformatJSON(res)
			if err == nil {
				cc.GetDebugLogger().Log("msg", "ETH RPC")
				fmt.Fprintf(logWriter, "=> ETH request\n%s\n", reqBody)
				fmt.Fprintf(logWriter, "<= ETH response\n%s\n", resBody)
			}
		}
	}))

	e.Use(func(h echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			cc := &myCtx{
				Context:     c,
				logWriter:   logWriter,
				logger:      s.logger,
				transformer: s.transformer,
			}

			c.Set("myctx", cc)

			return h(c)
		}
	})

	// support batch requests
	e.Use(batchRequestsMiddleware)

	e.HTTPErrorHandler = errorHandler
	e.HideBanner = true
	if s.mutex == nil {
		e.POST("/*", httpHandler)
		e.GET("/*", websocketHandler)
	} else {
		level.Info(s.logger).Log("msg", "Processing RPC requests single threaded")
		e.POST("/*", func(c echo.Context) error {
			s.mutex.Lock()
			defer s.mutex.Unlock()
			return httpHandler(c)
		})
		e.GET("/*", websocketHandler)
	}

	https := (s.httpsKey != "" && s.httpsCert != "")
	level.Warn(s.logger).Log("listen", s.address, "qtum_rpc", s.qtumRPCClient.URL, "msg", "proxy started", "https", https)

	if https {
		level.Info(s.logger).Log("msg", "SSL enabled")
		return e.StartTLS(s.address, s.httpsCert, s.httpsKey)
	} else {
		return e.Start(s.address)
	}
}

type Option func(*Server) error

func SetLogWriter(logWriter io.Writer) Option {
	return func(p *Server) error {
		p.logWriter = logWriter
		return nil
	}
}

func SetLogger(l log.Logger) Option {
	return func(p *Server) error {
		p.logger = l
		return nil
	}
}

func SetDebug(debug bool) Option {
	return func(p *Server) error {
		p.debug = debug
		return nil
	}
}

func SetSingleThreaded(singleThreaded bool) Option {
	return func(p *Server) error {
		if singleThreaded {
			p.mutex = &sync.Mutex{}
		} else {
			p.mutex = nil
		}
		return nil
	}
}

func SetHttps(key string, cert string) Option {
	return func(p *Server) error {
		p.httpsKey = key
		p.httpsCert = cert
		return nil
	}
}

func batchRequestsMiddleware(h echo.HandlerFunc) echo.HandlerFunc {
	return func(c echo.Context) error {
		myctx := c.Get("myctx")
		cc, ok := myctx.(*myCtx)
		if !ok {
			return errors.New("Could not find myctx")
		}

		// Request
		reqBody := []byte{}
		if c.Request().Body != nil { // Read
			var err error
			reqBody, err = ioutil.ReadAll(c.Request().Body)
			if err != nil {
				panic(fmt.Sprintf("%v", err))
			}
		}
		isBatchRequests := func(msg json.RawMessage) bool {
			return len(msg) != 0 && msg[0] == '['
		}
		c.Request().Body = ioutil.NopCloser(bytes.NewBuffer(reqBody)) // Reset

		if !isBatchRequests(reqBody) {
			return h(c)
		}

		var rpcReqs []*eth.JSONRPCRequest
		if err := c.Bind(&rpcReqs); err != nil {

			return err
		}

		results := make([]*eth.JSONRPCResult, 0, len(rpcReqs))

		for _, req := range rpcReqs {
			result, err := callHttpHandler(cc, req)
			if err != nil {
				return err
			}

			results = append(results, result)
		}

		return c.JSON(http.StatusOK, results)
	}
}

func callHttpHandler(cc *myCtx, req *eth.JSONRPCRequest) (*eth.JSONRPCResult, error) {
	reqBytes, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	httpreq := httptest.NewRequest(echo.POST, "/", ioutil.NopCloser(bytes.NewReader(reqBytes)))
	httpreq.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()

	newCtx := cc.Echo().NewContext(httpreq, rec)
	myCtx := &myCtx{
		Context:     newCtx,
		logWriter:   cc.GetLogWriter(),
		logger:      cc.logger,
		transformer: cc.transformer,
	}
	newCtx.Set("myctx", myCtx)
	if err = httpHandler(myCtx); err != nil {
		errorHandler(err, myCtx)
	}

	var result *eth.JSONRPCResult
	if err = json.Unmarshal(rec.Body.Bytes(), &result); err != nil {
		return nil, err
	}

	return result, nil
}
