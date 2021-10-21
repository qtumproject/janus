package transformer

import (
	"github.com/labstack/echo"
	"github.com/qtumproject/janus/pkg/eth"
	"github.com/qtumproject/janus/pkg/qtum"
)

// ProxyETHNetVersion implements ETHProxy
type ProxyETHNetVersion struct {
	*qtum.Qtum
}

func (p *ProxyETHNetVersion) Method() string {
	return "net_version"
}

func (p *ProxyETHNetVersion) Request(_ *eth.JSONRPCRequest, c echo.Context) (interface{}, error) {
	return p.request()
}

func (p *ProxyETHNetVersion) request() (*eth.NetVersionResponse, error) {
	var qtumresp *qtum.GetBlockChainInfoResponse
	if err := p.Qtum.Request(qtum.MethodGetBlockChainInfo, nil, &qtumresp); err != nil {
		return nil, err
	}

	var networkID string
	switch qtumresp.Chain {
	case "regtest":
		// See: https://github.com/trufflesuite/ganache/issues/112 for an idea on how to generate an ID.
		// https://github.com/ethereum/wiki/wiki/JSON-RPC#net_version
		networkID = "113"
	default:
		// TODO: discuss policy? NetworkID has to be an integer, can't just return qtumresp.Chain.
		networkID = "81"
		p.GetDebugLogger().Log("method", p.Method(), "msg", "Unknown chain "+qtumresp.Chain)
	}

	resp := eth.NetVersionResponse(networkID)
	p.GetDebugLogger().Log("method", p.Method(), "result", networkID)
	return &resp, nil
}
