package transformer

import (
	"github.com/pkg/errors"
	"github.com/qtumproject/janus/pkg/eth"
	"github.com/qtumproject/janus/pkg/qtum"
)

// ProxyETHGetBlockByNumber implements ETHProxy
type ProxyETHGetBlockByNumber struct {
	*qtum.Qtum
}

func (p *ProxyETHGetBlockByNumber) Method() string {
	return "eth_getBlockByNumber"
}

func (p *ProxyETHGetBlockByNumber) Request(rawreq *eth.JSONRPCRequest) (interface{}, error) {
	var req eth.GetBlockByNumberRequest
	if err := unmarshalRequest(rawreq.Params, &req); err != nil {
		return nil, errors.WithMessage(err, "couldn't unmarhsal request")
	}
	return p.request(&req)
}

func (p *ProxyETHGetBlockByNumber) request(req *eth.GetBlockByNumberRequest) (*eth.GetBlockByNumberResponse, error) {
	blockNum, err := getBlockNumberByParam(p.Qtum, req.BlockNumber, 0)
	if err != nil {
		return nil, errors.WithMessage(err, "couldn't get block number")
	}
	blockHash, err := p.GetBlockHash(blockNum)
	if err != nil {
		return nil, errors.WithMessage(err, "couldn't get block hash")
	}

	var (
		getBlockByHashReq = &eth.GetBlockByHashRequest{
			BlockHash:       string(blockHash),
			FullTransaction: req.FullTransaction,
		}
		proxy = &ProxyETHGetBlockByHash{Qtum: p.Qtum}
	)
	block, err := proxy.request(getBlockByHashReq)
	if err != nil {
		return nil, errors.WithMessage(err, "couldn't get block by hash")
	}
	return block, nil
}
