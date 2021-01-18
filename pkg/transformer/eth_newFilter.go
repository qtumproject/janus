package transformer

import (
	"encoding/json"

	"github.com/qtumproject/janus/pkg/utils"

	"math/big"

	"github.com/dcb9/go-ethereum/common/hexutil"
	"github.com/qtumproject/janus/pkg/eth"
	"github.com/qtumproject/janus/pkg/qtum"
)

// ProxyETHNewFilter implements ETHProxy
type ProxyETHNewFilter struct {
	*qtum.Qtum
	filter *eth.FilterSimulator
}

func (p *ProxyETHNewFilter) Method() string {
	return "eth_newFilter"
}

func (p *ProxyETHNewFilter) Request(rawreq *eth.JSONRPCRequest) (interface{}, error) {
	var req eth.NewFilterRequest
	if err := json.Unmarshal(rawreq.Params, &req); err != nil {
		return nil, err
	}

	return p.request(&req)
}

func (p *ProxyETHNewFilter) request(ethreq *eth.NewFilterRequest) (eth.NewFilterResponse, error) {
	var from *big.Int
	var err error
	if ethreq.FromBlock == nil {
		from, err = getBlockNumberByParam(p.Qtum, []byte("latest"), 0)
	} else {
		from, err = getBlockNumberByParam(p.Qtum, ethreq.FromBlock, 0)
	}
	if err != nil {
		return "", err
	}

	filter := p.filter.New(eth.NewFilterTy, ethreq)
	filter.Data.Store("lastBlockNumber", from.Uint64())

	if len(ethreq.Topics) > 0 {
		filter.Data.Store("topics", convertTopics(ethreq.Topics))
	}

	return eth.NewFilterResponse(hexutil.EncodeUint64(filter.ID)), nil
}

func convertTopics(ethtopics []interface{}) []interface{} {
	var topics []interface{}
	for _, topic := range ethtopics {
		switch topic.(type) {
		case []interface{}:
			topics = append(topics, convertTopics(topic.([]interface{})))
		case string:
			topics = append(topics, utils.RemoveHexPrefix(topic.(string)))
		case nil:
			topics = append(topics, "null")
		}
	}

	return topics
}
