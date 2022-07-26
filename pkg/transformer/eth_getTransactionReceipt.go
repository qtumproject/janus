package transformer

import (
	"context"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
	"github.com/qtumproject/janus/pkg/conversion"
	"github.com/qtumproject/janus/pkg/eth"
	"github.com/qtumproject/janus/pkg/qtum"
	"github.com/qtumproject/janus/pkg/utils"
)

var STATUS_SUCCESS = "0x1"
var STATUS_FAILURE = "0x0"

// ProxyETHGetTransactionReceipt implements ETHProxy
type ProxyETHGetTransactionReceipt struct {
	*qtum.Qtum
}

func (p *ProxyETHGetTransactionReceipt) Method() string {
	return "eth_getTransactionReceipt"
}

func (p *ProxyETHGetTransactionReceipt) Request(rawreq *eth.JSONRPCRequest, c echo.Context) (interface{}, eth.JSONRPCError) {
	var req eth.GetTransactionReceiptRequest
	if err := unmarshalRequest(rawreq.Params, &req); err != nil {
		// TODO: Correct error code?
		return nil, eth.NewInvalidParamsError(err.Error())
	}
	if req == "" {
		// TODO: Correct error code?
		return nil, eth.NewInvalidParamsError("empty transaction hash")
	}
	var (
		txHash  = utils.RemoveHexPrefix(string(req))
		qtumReq = qtum.GetTransactionReceiptRequest(txHash)
	)
	return p.request(c.Request().Context(), &qtumReq)
}

func (p *ProxyETHGetTransactionReceipt) request(ctx context.Context, req *qtum.GetTransactionReceiptRequest) (*eth.GetTransactionReceiptResponse, eth.JSONRPCError) {
	qtumReceipt, err := p.Qtum.GetTransactionReceipt(ctx, string(*req))
	if err != nil {
		ethTx, _, getRewardTransactionErr := getRewardTransactionByHash(ctx, p.Qtum, string(*req))
		if getRewardTransactionErr != nil {
			errCause := errors.Cause(err)
			if errCause == qtum.EmptyResponseErr {
				return nil, nil
			}
			p.Qtum.GetDebugLogger().Log("msg", "Transaction does not exist", "txid", string(*req))
			return nil, eth.NewCallbackError(err.Error())
		}
		if ethTx == nil {
			// unconfirmed tx, return nil
			// https://github.com/openethereum/parity-ethereum/issues/3482
			return nil, nil
		}
		return &eth.GetTransactionReceiptResponse{
			TransactionHash:  ethTx.Hash,
			TransactionIndex: ethTx.TransactionIndex,
			BlockHash:        ethTx.BlockHash,
			BlockNumber:      ethTx.BlockNumber,
			// TODO: This is higher than GasUsed in geth but does it matter?
			CumulativeGasUsed: NonContractVMGasLimit,
			EffectiveGasPrice: "0x0",
			GasUsed:           NonContractVMGasLimit,
			From:              ethTx.From,
			To:                ethTx.To,
			Logs:              []eth.Log{},
			LogsBloom:         eth.EmptyLogsBloom,
			Status:            STATUS_SUCCESS,
		}, nil
	}

	ethReceipt := &eth.GetTransactionReceiptResponse{
		TransactionHash:   utils.AddHexPrefix(qtumReceipt.TransactionHash),
		TransactionIndex:  hexutil.EncodeUint64(qtumReceipt.TransactionIndex),
		BlockHash:         utils.AddHexPrefix(qtumReceipt.BlockHash),
		BlockNumber:       hexutil.EncodeUint64(qtumReceipt.BlockNumber),
		ContractAddress:   utils.AddHexPrefixIfNotEmpty(qtumReceipt.ContractAddress),
		CumulativeGasUsed: hexutil.EncodeUint64(qtumReceipt.CumulativeGasUsed),
		EffectiveGasPrice: "0x0",
		GasUsed:           hexutil.EncodeUint64(qtumReceipt.GasUsed),
		From:              utils.AddHexPrefixIfNotEmpty(qtumReceipt.From),
		To:                utils.AddHexPrefixIfNotEmpty(qtumReceipt.To),

		// TODO: researching
		// ! Temporary accept this value to be always zero, as it is at eth logs
		LogsBloom: eth.EmptyLogsBloom,
	}

	status := STATUS_FAILURE
	if qtumReceipt.Excepted == "None" {
		status = STATUS_SUCCESS
	} else {
		p.Qtum.GetDebugLogger().Log("transaction", ethReceipt.TransactionHash, "msg", "transaction excepted", "message", qtumReceipt.Excepted)
	}
	ethReceipt.Status = status

	r := qtum.TransactionReceipt(*qtumReceipt)
	ethReceipt.Logs = conversion.ExtractETHLogsFromTransactionReceipt(&r, r.Log)

	qtumTx, err := p.Qtum.GetRawTransaction(ctx, qtumReceipt.TransactionHash, false)
	if err != nil {
		p.GetDebugLogger().Log("msg", "couldn't get transaction", "err", err)
		return nil, eth.NewCallbackError("couldn't get transaction")
	}
	decodedRawQtumTx, err := p.Qtum.DecodeRawTransaction(ctx, qtumTx.Hex)
	if err != nil {
		p.GetDebugLogger().Log("msg", "couldn't decode raw transaction", "err", err)
		return nil, eth.NewCallbackError("couldn't decode raw transaction")
	}
	if decodedRawQtumTx.IsContractCreation() {
		ethReceipt.To = ""
	} else {
		ethReceipt.ContractAddress = ""
	}

	// TODO: researching
	// - The following code reason is unknown (see original comment)
	// - Code temporary commented, until an error occures
	// ! Do not remove
	// // contractAddress : DATA, 20 Bytes - The contract address created, if the transaction was a contract creation, otherwise null.
	// if status != "0x1" {
	// 	// if failure, should return null for contractAddress, instead of the zero address.
	// 	ethTxReceipt.ContractAddress = ""
	// }

	return ethReceipt, nil
}
