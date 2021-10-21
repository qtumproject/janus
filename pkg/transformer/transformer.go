package transformer

import (
	"github.com/go-kit/kit/log"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
	"github.com/qtumproject/janus/pkg/eth"
	"github.com/qtumproject/janus/pkg/notifier"
	"github.com/qtumproject/janus/pkg/qtum"
)

type Transformer struct {
	qtumClient   *qtum.Qtum
	debugMode    bool
	logger       log.Logger
	transformers map[string]ETHProxy
}

// New creates a new Transformer
func New(qtumClient *qtum.Qtum, proxies []ETHProxy, opts ...Option) (*Transformer, error) {
	if qtumClient == nil {
		return nil, errors.New("qtumClient cannot be nil")
	}

	t := &Transformer{
		qtumClient: qtumClient,
		logger:     log.NewNopLogger(),
	}

	var err error
	for _, p := range proxies {
		if err = t.Register(p); err != nil {
			return nil, err
		}
	}

	for _, opt := range opts {
		if err := opt(t); err != nil {
			return nil, err
		}
	}

	return t, nil
}

// Register registers an ETHProxy to a Transformer
func (t *Transformer) Register(p ETHProxy) error {
	if t.transformers == nil {
		t.transformers = make(map[string]ETHProxy)
	}

	m := p.Method()
	if _, ok := t.transformers[m]; ok {
		return errors.Errorf("method already exist: %s ", m)
	}

	t.transformers[m] = p

	return nil
}

// Transform takes a Transformer and transforms the request from ETH request and returns the proxy request
func (t *Transformer) Transform(req *eth.JSONRPCRequest, c echo.Context) (interface{}, error) {
	proxy, err := t.getProxy(req.Method)
	if err != nil {
		return nil, errors.WithMessage(err, "couldn't get proxy")
	}
	resp, err := proxy.Request(req, c)
	if err != nil {
		return nil, errors.WithMessagef(err, "couldn't proxy %s request", req.Method)
	}
	return resp, nil
}

func (t *Transformer) getProxy(method string) (ETHProxy, error) {
	proxy, ok := t.transformers[method]
	if !ok {
		return nil, errors.Errorf("The method %s does not exist/is not available", method)
	}
	return proxy, nil
}

func (t *Transformer) IsDebugEnabled() bool {
	return t.debugMode
}

// DefaultProxies are the default proxy methods made available
func DefaultProxies(qtumRPCClient *qtum.Qtum, agent *notifier.Agent) []ETHProxy {
	filter := eth.NewFilterSimulator()
	getFilterChanges := &ProxyETHGetFilterChanges{Qtum: qtumRPCClient, filter: filter}
	ethCall := &ProxyETHCall{Qtum: qtumRPCClient}

	return []ETHProxy{
		ethCall,
		&ProxyNetListening{Qtum: qtumRPCClient},
		&ProxyETHPersonalUnlockAccount{},
		&ProxyETHChainId{Qtum: qtumRPCClient},
		&ProxyETHBlockNumber{Qtum: qtumRPCClient},
		&ProxyETHHashrate{Qtum: qtumRPCClient},
		&ProxyETHMining{Qtum: qtumRPCClient},
		&ProxyETHNetVersion{Qtum: qtumRPCClient},
		&ProxyETHGetTransactionByHash{Qtum: qtumRPCClient},
		&ProxyETHGetTransactionByBlockNumberAndIndex{Qtum: qtumRPCClient},
		&ProxyETHGetLogs{Qtum: qtumRPCClient},
		&ProxyETHGetTransactionReceipt{Qtum: qtumRPCClient},
		&ProxyETHSendTransaction{Qtum: qtumRPCClient},
		&ProxyETHAccounts{Qtum: qtumRPCClient},
		&ProxyETHGetCode{Qtum: qtumRPCClient},

		&ProxyETHNewFilter{Qtum: qtumRPCClient, filter: filter},
		&ProxyETHNewBlockFilter{Qtum: qtumRPCClient, filter: filter},
		getFilterChanges,
		&ProxyETHGetFilterLogs{ProxyETHGetFilterChanges: getFilterChanges},
		&ProxyETHUninstallFilter{Qtum: qtumRPCClient, filter: filter},

		&ProxyETHEstimateGas{ProxyETHCall: ethCall},
		&ProxyETHGetBlockByNumber{Qtum: qtumRPCClient},
		&ProxyETHGetBlockByHash{Qtum: qtumRPCClient},
		&ProxyETHGetBalance{Qtum: qtumRPCClient},
		&ProxyETHGetStorageAt{Qtum: qtumRPCClient},
		&ETHGetCompilers{},
		&ETHProtocolVersion{},
		&ETHGetUncleByBlockHashAndIndex{},
		&ETHGetUncleCountByBlockHash{},
		&ETHGetUncleCountByBlockNumber{},
		&Web3ClientVersion{},
		&Web3Sha3{},
		&ProxyETHSign{Qtum: qtumRPCClient},
		&ProxyETHGasPrice{Qtum: qtumRPCClient},
		&ProxyETHTxCount{Qtum: qtumRPCClient},
		&ProxyETHSignTransaction{Qtum: qtumRPCClient},
		&ProxyETHSendRawTransaction{Qtum: qtumRPCClient},

		&ETHSubscribe{Qtum: qtumRPCClient, Agent: agent},
		&ETHUnsubscribe{Qtum: qtumRPCClient, Agent: agent},

		&ProxyQTUMGetUTXOs{Qtum: qtumRPCClient},

		&ProxyNetPeerCount{Qtum: qtumRPCClient},
	}
}

func SetDebug(debug bool) func(*Transformer) error {
	return func(t *Transformer) error {
		t.debugMode = debug
		return nil
	}
}

func SetLogger(l log.Logger) func(*Transformer) error {
	return func(t *Transformer) error {
		t.logger = log.WithPrefix(l, "component", "transformer")
		return nil
	}
}
