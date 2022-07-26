package transformer

import (
	"context"
	"fmt"
	"strings"

	"github.com/labstack/echo"
	"github.com/qtumproject/janus/pkg/eth"
	"github.com/qtumproject/janus/pkg/qtum"
	"github.com/qtumproject/janus/pkg/utils"
	"github.com/shopspring/decimal"
)

// ProxyETHSendTransaction implements ETHProxy
type ProxyETHSignTransaction struct {
	*qtum.Qtum
}

func (p *ProxyETHSignTransaction) Method() string {
	return "eth_signTransaction"
}

func (p *ProxyETHSignTransaction) Request(rawreq *eth.JSONRPCRequest, c echo.Context) (interface{}, eth.JSONRPCError) {
	var req eth.SendTransactionRequest
	if err := unmarshalRequest(rawreq.Params, &req); err != nil {
		// TODO: Correct error code?
		return nil, eth.NewInvalidParamsError(err.Error())
	}

	ctx := c.Request().Context()

	if req.IsCreateContract() {
		p.GetDebugLogger().Log("method", p.Method(), "msg", "transaction is a create contract request")
		return p.requestCreateContract(ctx, &req)
	} else if req.IsSendEther() {
		p.GetDebugLogger().Log("method", p.Method(), "msg", "transaction is a send ether request")
		return p.requestSendToAddress(ctx, &req)
	} else if req.IsCallContract() {
		p.GetDebugLogger().Log("method", p.Method(), "msg", "transaction is a call contract request")
		return p.requestSendToContract(ctx, &req)
	} else {
		p.GetDebugLogger().Log("method", p.Method(), "msg", "transaction is an unknown request")
	}

	return nil, eth.NewInvalidParamsError("Unknown operation")
}

func (p *ProxyETHSignTransaction) getRequiredUtxos(ctx context.Context, from string, neededAmount decimal.Decimal) ([]qtum.RawTxInputs, decimal.Decimal, error) {
	//convert address to qtum address
	addr := utils.RemoveHexPrefix(from)
	base58Addr, err := p.FromHexAddress(addr)
	if err != nil {
		return nil, decimal.Decimal{}, err
	}
	// need to get utxos with txid and vouts. In order to do this we get a list of unspent transactions and begin summing them up
	var getaddressutxos *qtum.GetAddressUTXOsRequest = &qtum.GetAddressUTXOsRequest{Addresses: []string{base58Addr}}
	qtumresp, err := p.GetAddressUTXOs(ctx, getaddressutxos)
	if err != nil {
		return nil, decimal.Decimal{}, err
	}

	//Convert minSumAmount to Satoshis
	minimumSum := convertFromQtumToSatoshis(neededAmount)
	var utxos []qtum.RawTxInputs
	var minUTXOsSum decimal.Decimal
	for _, utxo := range *qtumresp {
		minUTXOsSum = minUTXOsSum.Add(utxo.Satoshis)
		utxos = append(utxos, qtum.RawTxInputs{TxID: utxo.TXID, Vout: utxo.OutputIndex})
		if minUTXOsSum.GreaterThanOrEqual(minimumSum) {
			return utxos, minUTXOsSum, nil
		}
	}

	return nil, decimal.Decimal{}, fmt.Errorf("Insufficient UTXO value attempted to be sent")
}

func calculateChange(balance, neededAmount decimal.Decimal) (decimal.Decimal, error) {
	if balance.LessThan(neededAmount) {
		return decimal.Decimal{}, fmt.Errorf("insufficient funds to create fee to chain")
	}
	return balance.Sub(neededAmount), nil
}

func calculateNeededAmount(value, gasLimit, gasPrice decimal.Decimal) decimal.Decimal {
	return value.Add(gasLimit.Mul(gasPrice))
}

func (p *ProxyETHSignTransaction) requestSendToContract(ctx context.Context, ethtx *eth.SendTransactionRequest) (string, eth.JSONRPCError) {
	gasLimit, gasPrice, err := EthGasToQtum(ethtx)
	if err != nil {
		return "", eth.NewInvalidParamsError(err.Error())
	}

	amount := decimal.NewFromFloat(0.0)
	if ethtx.Value != "" {
		var err error
		amount, err = EthValueToQtumAmount(ethtx.Value, ZeroSatoshi)
		if err != nil {
			return "", eth.NewInvalidParamsError(err.Error())
		}
	}

	newGasPrice, err := decimal.NewFromString(gasPrice)
	if err != nil {
		return "", eth.NewInvalidParamsError(err.Error())
	}
	neededAmount := calculateNeededAmount(amount, decimal.NewFromBigInt(gasLimit, 0), newGasPrice)

	inputs, balance, err := p.getRequiredUtxos(ctx, ethtx.From, neededAmount)
	if err != nil {
		return "", eth.NewCallbackError(err.Error())
	}

	change, err := calculateChange(balance, neededAmount)
	if err != nil {
		return "", eth.NewCallbackError(err.Error())
	}

	contractInteractTx := &qtum.SendToContractRawRequest{
		ContractAddress: utils.RemoveHexPrefix(ethtx.To),
		Datahex:         utils.RemoveHexPrefix(ethtx.Data),
		Amount:          amount,
		GasLimit:        gasLimit,
		GasPrice:        gasPrice,
	}

	if from := ethtx.From; from != "" && utils.IsEthHexAddress(from) {
		from, err = p.FromHexAddress(from)
		if err != nil {
			return "", eth.NewInvalidParamsError(err.Error())
		}
		contractInteractTx.SenderAddress = from
	}

	fromAddr := utils.RemoveHexPrefix(ethtx.From)

	acc := p.Qtum.Accounts.FindByHexAddress(strings.ToLower(fromAddr))
	if acc == nil {
		return "", eth.NewInvalidParamsError(fmt.Sprintf("No such account: %s", fromAddr))
	}

	rawtxreq := []interface{}{inputs, []interface{}{map[string]*qtum.SendToContractRawRequest{"contract": contractInteractTx}, map[string]decimal.Decimal{contractInteractTx.SenderAddress: change}}}
	var rawTx string
	if err := p.Qtum.Request(qtum.MethodCreateRawTx, rawtxreq, &rawTx); err != nil {
		return "", eth.NewCallbackError(err.Error())
	}

	var resp *qtum.SignRawTxResponse
	if err := p.Qtum.Request(qtum.MethodSignRawTx, []interface{}{rawTx}, &resp); err != nil {
		return "", eth.NewCallbackError(err.Error())
	}
	if !resp.Complete {
		return "", eth.NewCallbackError("something went wrong with signing the transaction; transaction incomplete")
	}
	return utils.AddHexPrefix(resp.Hex), nil
}

func (p *ProxyETHSignTransaction) requestSendToAddress(ctx context.Context, req *eth.SendTransactionRequest) (string, eth.JSONRPCError) {
	getQtumWalletAddress := func(addr string) (string, error) {
		if utils.IsEthHexAddress(addr) {
			return p.FromHexAddress(utils.RemoveHexPrefix(addr))
		}
		return addr, nil
	}

	to, err := getQtumWalletAddress(req.To)
	if err != nil {
		return "", eth.NewCallbackError(err.Error())
	}

	from, err := getQtumWalletAddress(req.From)
	if err != nil {
		return "", eth.NewCallbackError(err.Error())
	}

	amount, err := EthValueToQtumAmount(req.Value, ZeroSatoshi)
	if err != nil {
		return "", eth.NewInvalidParamsError(err.Error())
	}

	inputs, balance, err := p.getRequiredUtxos(ctx, req.From, amount)
	if err != nil {
		return "", eth.NewCallbackError(err.Error())
	}

	change, err := calculateChange(balance, amount)
	if err != nil {
		return "", eth.NewCallbackError(err.Error())
	}

	var addressValMap = map[string]decimal.Decimal{to: amount, from: change}
	rawtxreq := []interface{}{inputs, addressValMap}
	var rawTx string
	if err := p.Qtum.Request(qtum.MethodCreateRawTx, rawtxreq, &rawTx); err != nil {
		return "", eth.NewCallbackError(err.Error())
	}

	var resp *qtum.SignRawTxResponse
	signrawtxreq := []interface{}{rawTx}
	if err := p.Qtum.Request(qtum.MethodSignRawTx, signrawtxreq, &resp); err != nil {
		return "", eth.NewCallbackError(err.Error())
	}
	if !resp.Complete {
		return "", eth.NewCallbackError("something went wrong with signing the transaction; transaction incomplete")
	}
	return utils.AddHexPrefix(resp.Hex), nil
}

func (p *ProxyETHSignTransaction) requestCreateContract(ctx context.Context, req *eth.SendTransactionRequest) (string, eth.JSONRPCError) {
	gasLimit, gasPrice, err := EthGasToQtum(req)
	if err != nil {
		return "", eth.NewInvalidParamsError(err.Error())
	}

	from := req.From
	if utils.IsEthHexAddress(from) {
		from, err = p.FromHexAddress(from)
		if err != nil {
			return "", eth.NewInvalidParamsError(err.Error())
		}
	}

	contractDeploymentTx := &qtum.CreateContractRawRequest{
		ByteCode:      utils.RemoveHexPrefix(req.Data),
		GasLimit:      gasLimit,
		GasPrice:      gasPrice,
		SenderAddress: from,
	}

	newGasPrice, err := decimal.NewFromString(gasPrice)
	if err != nil {
		return "", eth.NewInvalidParamsError(err.Error())
	}
	neededAmount := calculateNeededAmount(decimal.NewFromFloat(0.0), decimal.NewFromBigInt(gasLimit, 0), newGasPrice)

	inputs, balance, err := p.getRequiredUtxos(ctx, req.From, neededAmount)
	if err != nil {
		return "", eth.NewCallbackError(err.Error())
	}

	change, err := calculateChange(balance, neededAmount)
	if err != nil {
		return "", eth.NewCallbackError(err.Error())
	}

	rawtxreq := []interface{}{inputs, []interface{}{map[string]*qtum.CreateContractRawRequest{"contract": contractDeploymentTx}, map[string]decimal.Decimal{from: change}}}
	var rawTx string
	if err := p.Qtum.Request(qtum.MethodCreateRawTx, rawtxreq, &rawTx); err != nil {
		return "", eth.NewCallbackError(err.Error())
	}

	var resp *qtum.SignRawTxResponse
	signrawtxreq := []interface{}{rawTx}
	if err := p.Qtum.Request(qtum.MethodSignRawTx, signrawtxreq, &resp); err != nil {
		return "", eth.NewCallbackError(err.Error())
	}
	if !resp.Complete {
		return "", eth.NewCallbackError("something went wrong with signing the transaction; transaction incomplete")
	}
	return utils.AddHexPrefix(resp.Hex), nil
}
