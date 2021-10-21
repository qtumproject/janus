package transformer

import (
	"encoding/json"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
	"github.com/qtumproject/janus/pkg/eth"
	"github.com/qtumproject/janus/pkg/qtum"
)

// ProxyETHGetTransactionByBlockHashAndIndex implements ETHProxy
type ProxyETHGetTransactionByBlockHashAndIndex struct {
	*qtum.Qtum
}

func (p *ProxyETHGetTransactionByBlockHashAndIndex) Method() string {
	return "eth_getTransactionByBlockHashAndIndex"
}

func (p *ProxyETHGetTransactionByBlockHashAndIndex) Request(rawreq *eth.JSONRPCRequest, c echo.Context) (interface{}, error) {
	var req eth.GetTransactionByBlockHashAndIndex
	if err := json.Unmarshal(rawreq.Params, &req); err != nil {
		return nil, errors.Wrap(err, "couldn't unmarshal request")
	}
	if req.BlockHash == "" {
		return nil, errors.New("invalid argument 0: empty hex string")
	}

	return p.request(&req)
}

func (p *ProxyETHGetTransactionByBlockHashAndIndex) request(req *eth.GetTransactionByBlockHashAndIndex) (interface{}, error) {
	transactionIndex, err := hexutil.DecodeUint64(req.TransactionIndex)
	if err != nil {
		return nil, errors.Wrap(err, "invalid argument 1")
	}

	// Proxy eth_getBlockByHash and return the transaction at requested index
	getBlockByNumber := ProxyETHGetBlockByHash{p.Qtum}
	blockByNumber, err := getBlockByNumber.request(&eth.GetBlockByHashRequest{BlockHash: req.BlockHash, FullTransaction: true})

	if err != nil {
		return nil, err
	}

	if blockByNumber == nil {
		return nil, nil
	}

	if len(blockByNumber.Transactions) <= int(transactionIndex) {
		return nil, nil
	}

	return blockByNumber.Transactions[int(transactionIndex)], nil
}
