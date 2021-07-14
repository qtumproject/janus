package transformer

import (
	"github.com/pkg/errors"
	"github.com/qtumproject/janus/pkg/eth"
	"github.com/qtumproject/janus/pkg/qtum"
)

// ProxyETHSyncStatus implements ETHProxy
type ProxyETHSyncing struct {
	*qtum.Qtum
}

func (p *ProxyETHSyncing) Method() string {
	return "eth_syncing"
}

func (p *ProxyETHSyncing) Request(rpcReq *eth.JSONRPCRequest) (interface{}, error) {
	req := new(eth.SyncingRequest)
	if err := unmarshalRequest(rpcReq.Params, req); err != nil {
		return nil, errors.WithMessage(err, "couldn't unmarhsal rpc request")
	}
	return p.request(req)
}

func (p *ProxyETHSyncing) request(req *eth.SyncingRequest) (*eth.SyncingResponse, error) {

	var (
		getSyncStatusReq = &eth.SyncingRequest{
			Message: req.Message,
		}
		proxy = &ProxyETHSyncing{Qtum: p.Qtum}
	)
	status, err := proxy.request(getSyncStatusReq)
	if err != nil {
		return nil, errors.WithMessage(err, "couldn't get sync status")
	}

	return status, nil
}
