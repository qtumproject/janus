package transformer

import (
	"github.com/labstack/echo"
	"github.com/pkg/errors"
	"github.com/qtumproject/janus/pkg/eth"
	"github.com/qtumproject/janus/pkg/qtum"
	"github.com/qtumproject/janus/pkg/utils"
	"github.com/shopspring/decimal"
)

// ProxyETHSendTransaction implements ETHProxy
type ProxyETHSendTransaction struct {
	*qtum.Qtum
}

func (p *ProxyETHSendTransaction) Method() string {
	return "eth_sendTransaction"
}

func (p *ProxyETHSendTransaction) Request(rawreq *eth.JSONRPCRequest, c echo.Context) (interface{}, error) {
	var req eth.SendTransactionRequest
	err := unmarshalRequest(rawreq.Params, &req)
	if err != nil {
		return nil, err
	}

	var result interface{}

	if req.IsCreateContract() {
		result, err = p.requestCreateContract(&req)
	} else if req.IsSendEther() {
		result, err = p.requestSendToAddress(&req)
	} else if req.IsCallContract() {
		result, err = p.requestSendToContract(&req)
	} else {
		return nil, errors.New("Unknown operation")
	}

	if p.CanGenerate() {
		p.GenerateIfPossible()
	}

	return result, err
}

func (p *ProxyETHSendTransaction) requestSendToContract(ethtx *eth.SendTransactionRequest) (*eth.SendTransactionResponse, error) {
	gasLimit, gasPrice, err := EthGasToQtum(ethtx)
	if err != nil {
		return nil, err
	}

	amount := decimal.NewFromFloat(0.0)
	if ethtx.Value != "" {
		var err error
		amount, err = EthValueToQtumAmount(ethtx.Value, ZeroSatoshi)
		if err != nil {
			return nil, errors.Wrap(err, "EthValueToQtumAmount:")
		}
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
			return nil, err
		}
		qtumreq.SenderAddress = from
	}

	var resp *qtum.SendToContractResponse
	if err := p.Qtum.Request(qtum.MethodSendToContract, &qtumreq, &resp); err != nil {
		return nil, err
	}

	ethresp := eth.SendTransactionResponse(utils.AddHexPrefix(resp.Txid))
	return &ethresp, nil
}

func (p *ProxyETHSendTransaction) requestSendToAddress(req *eth.SendTransactionRequest) (*eth.SendTransactionResponse, error) {
	getQtumWalletAddress := func(addr string) (string, error) {
		if utils.IsEthHexAddress(addr) {
			return p.FromHexAddress(utils.RemoveHexPrefix(addr))
		}
		return addr, nil
	}

	from, err := getQtumWalletAddress(req.From)
	if err != nil {
		return nil, err
	}

	to, err := getQtumWalletAddress(req.To)
	if err != nil {
		return nil, err
	}

	amount, err := EthValueToQtumAmount(req.Value, ZeroSatoshi)
	if err != nil {
		return nil, errors.Wrap(err, "EthValueToQtumAmount:")
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
		return nil, err
	}

	ethresp := eth.SendTransactionResponse(utils.AddHexPrefix(string(qtumresp)))

	return &ethresp, nil
}

func (p *ProxyETHSendTransaction) requestCreateContract(req *eth.SendTransactionRequest) (*eth.SendTransactionResponse, error) {
	gasLimit, gasPrice, err := EthGasToQtum(req)
	if err != nil {
		return nil, err
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
				return nil, err
			}
		}

		qtumreq.SenderAddress = from
	}

	var resp *qtum.CreateContractResponse
	if err := p.Qtum.Request(qtum.MethodCreateContract, qtumreq, &resp); err != nil {
		return nil, err
	}

	ethresp := eth.SendTransactionResponse(utils.AddHexPrefix(string(resp.Txid)))

	return &ethresp, nil
}
