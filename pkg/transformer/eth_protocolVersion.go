package transformer

import (
	"github.com/labstack/echo"
	"github.com/qtumproject/janus/pkg/eth"
)

type ETHProtocolVersion struct {
}

func (p *ETHProtocolVersion) Method() string {
	return "eth_protocolVersion"
}

func (p *ETHProtocolVersion) Request(rawreq *eth.JSONRPCRequest, c echo.Context) (interface{}, error) {
	return "0x41", nil
}
