package transformer

import (
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

func (p *ProxyETHGetCode) Request(rawreq *eth.JSONRPCRequest, c echo.Context) (interface{}, error) {
	var req eth.GetCodeRequest
	if err := unmarshalRequest(rawreq.Params, &req); err != nil {
		return nil, err
	}

	return p.request(&req)
}

func (p *ProxyETHGetCode) request(ethreq *eth.GetCodeRequest) (eth.GetCodeResponse, error) {
	qtumreq := qtum.GetAccountInfoRequest(utils.RemoveHexPrefix(ethreq.Address))

	qtumresp, err := p.GetAccountInfo(&qtumreq)
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
			return "", err
		}
	}

	// qtum res -> eth res
	return eth.GetCodeResponse(utils.AddHexPrefix(qtumresp.Code)), nil
}
