package transformer

import (
	"encoding/json"

	"github.com/qtumproject/janus/pkg/eth"
	"github.com/qtumproject/janus/pkg/qtum"
	"github.com/qtumproject/janus/pkg/utils"
)

// ProxyETHGetLogs implements ETHProxy
type ProxyETHGetLogs struct {
	*qtum.Qtum
}

func (p *ProxyETHGetLogs) Method() string {
	return "eth_getLogs"
}

func (p *ProxyETHGetLogs) Request(rawreq *eth.JSONRPCRequest) (interface{}, error) {
	var req eth.GetLogsRequest
	if err := unmarshalRequest(rawreq.Params, &req); err != nil {
		return nil, err
	}

	// TODO: Graph Node is sending the topic
	// if len(req.Topics) != 0 {
	// 	return nil, errors.New("topics is not supported yet")
	// }

	qtumreq, err := p.ToRequest(&req)
	if err != nil {
		return nil, err
	}

	return p.request(qtumreq)
}

func (p *ProxyETHGetLogs) request(req *qtum.SearchLogsRequest) (*eth.GetLogsResponse, error) {
	receipts, err := p.SearchLogs(req)
	if err != nil {
		return nil, err
	}

	logs := make([]eth.Log, 0)
	for _, receipt := range receipts {
		r := qtum.TransactionReceipt(receipt)
		logs = append(logs, extractETHLogsFromTransactionReceipt(&r)...)
	}

	resp := eth.GetLogsResponse(logs)
	return &resp, nil
}

func (p *ProxyETHGetLogs) ToRequest(ethreq *eth.GetLogsRequest) (*qtum.SearchLogsRequest, error) {
	from, err := getBlockNumberByParam(p.Qtum, ethreq.FromBlock, 0)
	if err != nil {
		return nil, err
	}

	to, err := getBlockNumberByParam(p.Qtum, ethreq.ToBlock, -1)
	if err != nil {
		return nil, err
	}

	var addresses []string
	if ethreq.Address != nil {
		if isBytesOfString(ethreq.Address) {
			var addr string
			if err = json.Unmarshal(ethreq.Address, &addr); err != nil {
				return nil, err
			}
			addresses = append(addresses, addr)
		} else {
			if err = json.Unmarshal(ethreq.Address, &addresses); err != nil {
				return nil, err
			}
		}
		for i, _ := range addresses {
			addresses[i] = utils.RemoveHexPrefix(addresses[i])
		}
	}

	return &qtum.SearchLogsRequest{
		Addresses: addresses,
		FromBlock: from,
		ToBlock:   to,
	}, nil
}
