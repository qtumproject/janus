package transformer

import (
	"github.com/qtumproject/janus/pkg/eth"
	"github.com/qtumproject/janus/pkg/qtum"
)

//ProxyETHGetHashrate implements ETHProxy
type ProxyETHMining struct {
	*qtum.Qtum
}

func (p *ProxyETHMining) Method() string {
	return "eth_mining"
}

func (p *ProxyETHMining) Request(_ *eth.JSONRPCRequest) (interface{}, error){
	return p.request()
}

func (p *ProxyETHMining) request() (*eth.MiningResponse, error) {
	qtumresp, err := p.Qtum.GetMining()
	if err != nil {
		return nil, err
	}

	// qtum res -> eth res
	return p.ToResponse(qtumresp), nil
}

func (p *ProxyETHMining) ToResponse(qtumresp *qtum.GetMiningResponse) *eth.MiningResponse {
	ethresp := eth.MiningResponse(qtumresp.Staking)
	return &ethresp
}
