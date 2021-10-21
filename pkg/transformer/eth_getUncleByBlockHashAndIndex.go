package transformer

import (
	"github.com/labstack/echo"
	"github.com/qtumproject/janus/pkg/eth"
)

type ETHGetUncleByBlockHashAndIndex struct {
}

func (p *ETHGetUncleByBlockHashAndIndex) Method() string {
	return "eth_getUncleByBlockHashAndIndex"
}

func (p *ETHGetUncleByBlockHashAndIndex) Request(rawreq *eth.JSONRPCRequest, c echo.Context) (interface{}, error) {
	// hardcoded to nil
	return nil, nil
}
