package transformer

import (
	"github.com/labstack/echo"
	"github.com/qtumproject/janus/pkg/eth"
)

// Web3ClientVersion implements web3_clientVersion
type Web3ClientVersion struct {
	// *qtum.Qtum
}

func (p *Web3ClientVersion) Method() string {
	return "web3_clientVersion"
}

func (p *Web3ClientVersion) Request(_ *eth.JSONRPCRequest, c echo.Context) (interface{}, error) {
	return "QTUM ETHTestRPC/ethereum-js", nil
}

// func (p *Web3ClientVersion) ToResponse(ethresp *qtum.CallContractResponse) *eth.CallResponse {
// 	data := utils.AddHexPrefix(ethresp.ExecutionResult.Output)
// 	qtumresp := eth.CallResponse(data)
// 	return &qtumresp
// }
