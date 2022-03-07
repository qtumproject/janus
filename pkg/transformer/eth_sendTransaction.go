package transformer

import (
	"fmt"

	"github.com/labstack/echo"
	"github.com/qtumproject/janus/pkg/eth"
	"github.com/qtumproject/janus/pkg/qtum"
	"github.com/qtumproject/janus/pkg/utils"
	"github.com/shopspring/decimal"
)

var MinimumGasLimit = int64(22000)

// ProxyETHSendTransaction implements ETHProxy
type ProxyETHSendTransaction struct {
	*qtum.Qtum
}

func (p *ProxyETHSendTransaction) Method() string {
	return "eth_sendTransaction"
}

func (p *ProxyETHSendTransaction) Request(rawreq *eth.JSONRPCRequest, c echo.Context) (interface{}, eth.JSONRPCError) {
	var req eth.SendTransactionRequest
	err := unmarshalRequest(rawreq.Params, &req)
	if err != nil {
		// TODO: Correct error code?
		return nil, eth.NewInvalidParamsError(err.Error())
	}

	if req.Gas != nil && req.Gas.Int64() < MinimumGasLimit {
		p.GetLogger().Log("msg", "Gas limit is too low", "gasLimit", req.Gas.String())
	}

	var result interface{}
	var jsonErr eth.JSONRPCError

	if req.IsCreateContract() {
		result, jsonErr = p.requestCreateContract(&req)
	} else if req.IsSendEther() {
		result, jsonErr = p.requestSendToAddress(&req)
	} else if req.IsCallContract() {
		result, jsonErr = p.requestSendToContract(&req)
	} else {
		return nil, eth.NewInvalidParamsError("Unknown operation")
	}

	if err == nil {
		p.GenerateIfPossible()
	}

	return result, jsonErr
}

func (p *ProxyETHSendTransaction) requestSendToContract(ethtx *eth.SendTransactionRequest) (*eth.SendTransactionResponse, eth.JSONRPCError) {
	gasLimit, gasPrice, err := EthGasToQtum(ethtx)
	if err != nil {
		return nil, eth.NewInvalidParamsError(err.Error())
	}

	amount := decimal.NewFromFloat(0.0)
	if ethtx.Value != "" {
		var err error
		amount, err = ParseValue(p.Qtum, ethtx.Value, ZeroSatoshi)
		if err != nil {
			return nil, eth.NewInvalidParamsError(err.Error())
		}

		fmt.Println("****************************")
		fmt.Println(amount.String())
	}

	qtumreq := qtum.SendToContractRequest{
		ContractAddress: utils.RemoveHexPrefix(ethtx.To),
		Datahex:         utils.RemoveHexPrefix(ethtx.Data),
		Amount:          amount,
		GasLimit:        gasLimit,
		GasPrice:        gasPrice,
	}

	if from := ethtx.From; from != "" && utils.IsEthHexAddress(from) {
		from, err = p.FromHexAddress(from)
		if err != nil {
			return nil, eth.NewCallbackError(err.Error())
		}
		qtumreq.SenderAddress = from
	}

	var resp *qtum.SendToContractResponse
	if err := p.Qtum.Request(qtum.MethodSendToContract, &qtumreq, &resp); err != nil {
		return nil, eth.NewCallbackError(err.Error())
	}

	ethresp := eth.SendTransactionResponse(utils.AddHexPrefix(resp.Txid))
	return &ethresp, nil
}

func (p *ProxyETHSendTransaction) requestSendToAddress(req *eth.SendTransactionRequest) (*eth.SendTransactionResponse, eth.JSONRPCError) {
	getQtumWalletAddress := func(addr string) (string, error) {
		if utils.IsEthHexAddress(addr) {
			return p.FromHexAddress(utils.RemoveHexPrefix(addr))
		}
		return addr, nil
	}

	from, err := getQtumWalletAddress(req.From)
	if err != nil {
		return nil, eth.NewInvalidParamsError(err.Error())
	}

	to, err := getQtumWalletAddress(req.To)
	if err != nil {
		return nil, eth.NewInvalidParamsError(err.Error())
	}

	amount, err := ParseValue(p.Qtum, req.Value, ZeroSatoshi)
	if err != nil {
		return nil, eth.NewInvalidParamsError(err.Error())
	}

	p.GetDebugLogger().Log("msg", "successfully converted from wei to QTUM", "wei", req.Value, "qtum", amount)

	qtumreq := qtum.SendToAddressRequest{
		Address:       to,
		Amount:        amount,
		SenderAddress: from,
	}

	var qtumresp qtum.SendToAddressResponse
	if err := p.Qtum.Request(qtum.MethodSendToAddress, &qtumreq, &qtumresp); err != nil {
		// this can fail with:
		// "error": {
		//   "code": -3,
		//   "message": "Sender address does not have any unspent outputs"
		// }
		// this can happen if there are enough coins but some required are untrusted
		// you can get the trusted coin balance via getbalances rpc call
		return nil, eth.NewCallbackError(err.Error())
	}

	ethresp := eth.SendTransactionResponse(utils.AddHexPrefix(string(qtumresp)))

	return &ethresp, nil
}

func (p *ProxyETHSendTransaction) requestCreateContract(req *eth.SendTransactionRequest) (*eth.SendTransactionResponse, eth.JSONRPCError) {
	gasLimit, gasPrice, err := EthGasToQtum(req)
	if err != nil {
		return nil, eth.NewInvalidParamsError(err.Error())
	}

	qtumreq := &qtum.CreateContractRequest{
		ByteCode: utils.RemoveHexPrefix(req.Data),
		GasLimit: gasLimit,
		GasPrice: gasPrice,
	}

	if req.From != "" {
		from := req.From
		if utils.IsEthHexAddress(from) {
			from, err = p.FromHexAddress(from)
			if err != nil {
				return nil, eth.NewCallbackError(err.Error())
			}
		}

		qtumreq.SenderAddress = from
	}

	var resp *qtum.CreateContractResponse
	if err := p.Qtum.Request(qtum.MethodCreateContract, qtumreq, &resp); err != nil {
		return nil, eth.NewCallbackError(err.Error())
	}

	ethresp := eth.SendTransactionResponse(utils.AddHexPrefix(string(resp.Txid)))

	return &ethresp, nil
}
