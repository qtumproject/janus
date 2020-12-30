package transformer

import (
	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/pkg/errors"
	"github.com/qtumproject/janus/pkg/eth"
	"github.com/qtumproject/janus/pkg/qtum"
	"github.com/qtumproject/janus/pkg/utils"
)

// ProxyETHGetTransactionReceipt implements ETHProxy
type ProxyETHGetTransactionReceipt struct {
	*qtum.Qtum
}

func (p *ProxyETHGetTransactionReceipt) Method() string {
	return "eth_getTransactionReceipt"
}

func (p *ProxyETHGetTransactionReceipt) Request(rawreq *eth.JSONRPCRequest) (interface{}, error) {
	var req eth.GetTransactionReceiptRequest
	if err := unmarshalRequest(rawreq.Params, &req); err != nil {
		return nil, err
	}
	if req == "" {
		return nil, errors.New("empty transaction hash")
	}
	var (
		txHash  = utils.RemoveHexPrefix(string(req))
		qtumReq = qtum.GetTransactionReceiptRequest(txHash)
	)
	return p.request(&qtumReq)
}

func (p *ProxyETHGetTransactionReceipt) request(req *qtum.GetTransactionReceiptRequest) (*eth.GetTransactionReceiptResponse, error) {
	receipt, err := p.GetTransactionReceipt(string(*req))
	if err != nil {
		if err == qtum.EmptyResponseErr {
			ethTx, err := GetTransactionByHash(p.Qtum, string(*req), 0, 0)
			if err != nil {
				return nil, err
			}

			// TODO: Correct to normal values
			return &eth.GetTransactionReceiptResponse{
				TransactionHash:   ethTx.Hash,
				TransactionIndex:  "0x0",
				BlockHash:         ethTx.BlockHash,
				BlockNumber:       ethTx.BlockNumber,
				From:              ethTx.From,
				To:                ethTx.To,
				CumulativeGasUsed: ethTx.Gas,
				GasUsed:           ethTx.Gas,
				ContractAddress:   ethTx.To,
				Logs:              []eth.Log{},
				LogsBloom:         "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
				Status:            "0x1",
			}, nil
		}
		return nil, err
	}

	ethReceipt := &eth.GetTransactionReceiptResponse{
		TransactionHash:   utils.AddHexPrefix(qtumReceipt.TransactionHash),
		TransactionIndex:  hexutil.EncodeUint64(qtumReceipt.TransactionIndex),
		BlockHash:         utils.AddHexPrefix(qtumReceipt.BlockHash),
		BlockNumber:       hexutil.EncodeUint64(qtumReceipt.BlockNumber),
		ContractAddress:   utils.AddHexPrefix(qtumReceipt.ContractAddress),
		CumulativeGasUsed: hexutil.EncodeUint64(qtumReceipt.CumulativeGasUsed),
		GasUsed:           hexutil.EncodeUint64(qtumReceipt.GasUsed),
		From:              utils.AddHexPrefix(qtumReceipt.From),
		To:                utils.AddHexPrefix(qtumReceipt.To),

		// TODO: researching
		// ! Temporary accept this value to be always zero, as it is at eth logs
		LogsBloom: "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
	}

	status := "0x0"
	if qtumReceipt.Excepted == "None" {
		status = "0x1"
	}
	ethReceipt.Status = status

	r := qtum.TransactionReceiptStruct(*receipt)
	logs := getEthLogs(&r)

	// TODO: Correct to normal values
	ethTxReceipt := eth.GetTransactionReceiptResponse{
		TransactionHash:   utils.AddHexPrefix(receipt.TransactionHash),
		TransactionIndex:  hexutil.EncodeUint64(receipt.TransactionIndex),
		BlockHash:         utils.AddHexPrefix(receipt.BlockHash),
		BlockNumber:       hexutil.EncodeUint64(receipt.BlockNumber),
		ContractAddress:   utils.AddHexPrefix(receipt.ContractAddress),
		CumulativeGasUsed: hexutil.EncodeUint64(receipt.CumulativeGasUsed),
		GasUsed:           hexutil.EncodeUint64(receipt.GasUsed),
		Logs:              logs,
		Status:            status,

		// see Known issues
		LogsBloom: "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
	}
	decodedRawQtumTx, err := p.Qtum.DecodeRawTransaction(qtumTx.Hex)
	if err != nil {
		return nil, errors.WithMessage(err, "couldn't decode raw transaction")
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
