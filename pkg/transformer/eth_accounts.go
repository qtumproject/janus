package transformer

import (
	"github.com/labstack/echo"
	"github.com/qtumproject/janus/pkg/eth"
	"github.com/qtumproject/janus/pkg/qtum"
	"github.com/qtumproject/janus/pkg/utils"
)

// ProxyETHAccounts implements ETHProxy
type ProxyETHAccounts struct {
	*qtum.Qtum
}

func (p *ProxyETHAccounts) Method() string {
	return "eth_accounts"
}

func (p *ProxyETHAccounts) Request(_ *eth.JSONRPCRequest, c echo.Context) (interface{}, error) {
	return p.request()
}

func (p *ProxyETHAccounts) request() (eth.AccountsResponse, error) {
	var accounts eth.AccountsResponse

	for _, acc := range p.Accounts {
		acc := qtum.Account{acc}
		addr := acc.ToHexAddress()

		accounts = append(accounts, utils.AddHexPrefix(addr))
	}

	return accounts, nil
}

func (p *ProxyETHAccounts) ToResponse(ethresp *qtum.CallContractResponse) *eth.CallResponse {
	data := utils.AddHexPrefix(ethresp.ExecutionResult.Output)
	qtumresp := eth.CallResponse(data)
	return &qtumresp
}
