package transformer

import (
	"context"
	"math/big"

	"github.com/labstack/echo"
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

func (p *ProxyETHGetBlockByNumber) Request(rpcReq *eth.JSONRPCRequest, c echo.Context) (interface{}, eth.JSONRPCError) {
	req := new(eth.GetBlockByNumberRequest)
	if err := unmarshalRequest(rpcReq.Params, req); err != nil {
		// TODO: Correct error code?
		return nil, eth.NewInvalidParamsError(err.Error())
	}
	return p.request(c.Request().Context(), req)
}

func (p *ProxyETHGetBlockByNumber) request(ctx context.Context, req *eth.GetBlockByNumberRequest) (*eth.GetBlockByNumberResponse, eth.JSONRPCError) {
	blockNum, err := getBlockNumberByRawParam(ctx, p.Qtum, req.BlockNumber, false)
	if err != nil {
		return nil, eth.NewCallbackError("couldn't get block number by parameter")
	}

	blockHash, jsonErr := proxyETHGetBlockByHash(ctx, p, p.Qtum, blockNum)
	if jsonErr != nil {
		return nil, jsonErr
	}
	if blockHash == nil {
		return nil, nil
	}

	var (
		getBlockByHashReq = &eth.GetBlockByHashRequest{
			BlockHash:       string(*blockHash),
			FullTransaction: req.FullTransaction,
		}
		proxy = &ProxyETHGetBlockByHash{Qtum: p.Qtum}
	)
	block, jsonErr := proxy.request(ctx, getBlockByHashReq)
	if jsonErr != nil {
		p.GetDebugLogger().Log("function", p.Method(), "msg", "couldn't get block by hash", "err", err)
		return nil, eth.NewCallbackError("couldn't get block by hash")
	}
	if blockNum != nil {
		p.GetDebugLogger().Log("function", p.Method(), "request", string(req.BlockNumber), "msg", "Successfully got block by number", "result", blockNum.String())
	}
	return block, nil
}

// Properly handle unknown blocks
func proxyETHGetBlockByHash(ctx context.Context, p ETHProxy, q *qtum.Qtum, blockNum *big.Int) (*qtum.GetBlockHashResponse, eth.JSONRPCError) {
	resp, err := q.GetBlockHash(ctx, blockNum)
	if err != nil {
		if err == qtum.ErrInvalidParameter {
			// block doesn't exist, ETH rpc returns null
			/**
			{
				"jsonrpc": "2.0",
				"id": 1234,
				"result": null
			}
			**/
			q.GetDebugLogger().Log("function", p.Method(), "request", blockNum.String(), "msg", "Unknown block")
			return nil, nil
		}
		return nil, eth.NewCallbackError("couldn't get block hash")
	}
	return &resp, nil
}
