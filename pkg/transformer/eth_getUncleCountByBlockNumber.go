package transformer

import (
	"github.com/labstack/echo"
	"github.com/qtumproject/janus/pkg/eth"
)

type ETHGetUncleCountByBlockNumber struct {
}

func (p *ETHGetUncleCountByBlockNumber) Method() string {
	return "eth_getUncleCountByBlockNumber"
}

func (p *ETHGetUncleCountByBlockNumber) Request(rawreq *eth.JSONRPCRequest, c echo.Context) (interface{}, error) {
	// hardcoded to 0
	return "0x0", nil
}
