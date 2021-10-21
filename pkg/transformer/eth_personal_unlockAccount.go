package transformer

import (
	"github.com/labstack/echo"
	"github.com/qtumproject/janus/pkg/eth"
)

// ProxyETHPersonalUnlockAccount implements ETHProxy
type ProxyETHPersonalUnlockAccount struct{}

func (p *ProxyETHPersonalUnlockAccount) Method() string {
	return "personal_unlockAccount"
}

func (p *ProxyETHPersonalUnlockAccount) Request(req *eth.JSONRPCRequest, c echo.Context) (interface{}, error) {
	return eth.PersonalUnlockAccountResponse(true), nil
}
