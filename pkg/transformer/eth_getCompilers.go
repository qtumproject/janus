package transformer

import (
	"github.com/labstack/echo"
	"github.com/qtumproject/janus/pkg/eth"
)

type ETHGetCompilers struct {
}

func (p *ETHGetCompilers) Method() string {
	return "eth_getCompilers"
}

func (p *ETHGetCompilers) Request(rawreq *eth.JSONRPCRequest, c echo.Context) (interface{}, error) {
	// hardcoded to empty
	return []string{}, nil
}
