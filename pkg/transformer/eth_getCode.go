package transformer

import (
	"context"

	"github.com/labstack/echo"
	"github.com/qtumproject/janus/pkg/eth"
	"github.com/qtumproject/janus/pkg/qtum"
	"github.com/qtumproject/janus/pkg/utils"
)

// ProxyETHGetCode implements ETHProxy
type ProxyETHGetCode struct {
	*qtum.Qtum
}

func (p *ProxyETHGetCode) Method() string {
	return "eth_getCode"
}

func (p *ProxyETHGetCode) Request(rawreq *eth.JSONRPCRequest, c echo.Context) (interface{}, eth.JSONRPCError) {
	var req eth.GetCodeRequest
	if err := unmarshalRequest(rawreq.Params, &req); err != nil {
		// TODO: Correct error code?
		return nil, eth.NewInvalidParamsError(err.Error())
	}

	return p.request(c.Request().Context(), &req)
}

func (p *ProxyETHGetCode) request(ctx context.Context, ethreq *eth.GetCodeRequest) (eth.GetCodeResponse, eth.JSONRPCError) {
	qtumreq := qtum.GetAccountInfoRequest(utils.RemoveHexPrefix(ethreq.Address))

	qtumresp, err := p.GetAccountInfo(ctx, &qtumreq)
	if err != nil {
		if err == qtum.ErrInvalidAddress {
			/**
			// correct response for an invalid address
			{
				"jsonrpc": "2.0",
				"id": 123,
				"result": "0x"
			}
			**/
			return "0x", nil
		} else {
			return "", eth.NewCallbackError(err.Error())
		}
	}

	// qtum res -> eth res
	return eth.GetCodeResponse(utils.AddHexPrefix(qtumresp.Code)), nil
}
