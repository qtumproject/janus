package transformer

import (
	"github.com/labstack/echo"
	"github.com/qtumproject/janus/pkg/eth"
)

type ETHGetUncleCountByBlockHash struct {
}

func (p *ETHGetUncleCountByBlockHash) Method() string {
	return "eth_getUncleCountByBlockHash"
}

func (p *ETHGetUncleCountByBlockHash) Request(rawreq *eth.JSONRPCRequest, c echo.Context) (interface{}, error) {
	// hardcoded to 0
	return 0, nil
}
