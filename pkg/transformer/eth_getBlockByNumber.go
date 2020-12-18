package transformer

import (
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/pkg/errors"
	"github.com/qtumproject/janus/pkg/eth"
	"github.com/qtumproject/janus/pkg/qtum"
	"github.com/qtumproject/janus/pkg/utils"
)

// ProxyETHGetBlockByNumber implements ETHProxy
type ProxyETHGetBlockByNumber struct {
	*qtum.Qtum
}

func (p *ProxyETHGetBlockByNumber) Method() string {
	return "eth_getBlockByNumber"
}

func (p *ProxyETHGetBlockByNumber) Request(rpcReq *eth.JSONRPCRequest) (interface{}, error) {
	req := new(eth.GetBlockByNumberRequest)
	if err := unmarshalRequest(rpcReq.Params, req); err != nil {
		return nil, errors.WithMessage(err, "couldn't unmarhsal rpc request")
	}
	return p.request(req)
}

func (p *ProxyETHGetBlockByNumber) request(req *eth.GetBlockByNumberRequest) (*eth.GetBlockByNumberResponse, error) {
	blockNum, err := getBlockNumberByParam(p.Qtum, req.BlockNumber, 0)
	if err != nil {
		return nil, errors.WithMessage(err, "couldn't get block number by parameter")
	}
	blockHash, err := p.GetBlockHash(blockNum)
	if err != nil {
		return nil, errors.WithMessage(err, "couldn't get block hash")
	}

	if blockHeaderResp.Previousblockhash == "" {
		blockHeaderResp.Previousblockhash = "0000000000000000000000000000000000000000000000000000000000000000"
	}

	nonce := hexutil.EncodeUint64(uint64(blockHeaderResp.Nonce))

	if len(strings.TrimLeft(nonce, "0x")) < 16 {
		res := strings.TrimLeft(nonce, "0x")
		for i := 0; i < 16-len(res); {
			res = "0" + res
		}
		nonce = res
	}

	blockResp, err := p.GetBlock(string(blockHash))
	if err != nil {
		return nil, err
	}

	if !req.FullTransaction {
		txs := make([]string, 0, len(blockResp.Tx))
		for _, tx := range blockResp.Tx {
			txs = append(txs, utils.AddHexPrefix(tx))
		}

		return &eth.GetBlockByNumberResponse{
			Hash:             utils.AddHexPrefix(blockHeaderResp.Hash),
			Nonce:            utils.AddHexPrefix(nonce),
			Number:           hexutil.EncodeUint64(uint64(blockHeaderResp.Height)),
			ParentHash:       utils.AddHexPrefix(blockHeaderResp.Previousblockhash),
			Difficulty:       hexutil.EncodeUint64(uint64(blockHeaderResp.Difficulty)),
			Timestamp:        hexutil.EncodeUint64(blockHeaderResp.Time),
			StateRoot:        utils.AddHexPrefix(blockHeaderResp.HashStateRoot),
			Size:             hexutil.EncodeUint64(uint64(blockResp.Size)),
			Transactions:     txs,
			TransactionsRoot: utils.AddHexPrefix(blockResp.Merkleroot),
			ReceiptsRoot:     utils.AddHexPrefix(blockResp.Merkleroot),

			ExtraData:       "0x00",
			Miner:           "0x0000000000000000000000000000000000000000",
			TotalDifficulty: "0x00",
			GasLimit:        "0x00",
			GasUsed:         "0x00",
			LogsBloom:       "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",

			Sha3Uncles: "0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347",
			Uncles:     []string{},
		}, nil
	} else {
		txs := make([]eth.GetTransactionByHashResponse, 0, len(blockResp.Tx))
		for i, tx := range blockResp.Tx {
			if blockHeaderResp.Height == 0 {
				break
			}

			txData, err := p.GetTransaction(tx)
			if err != nil {
				return nil, err
			}

			ethVal, err := formatQtumAmount(txData.Amount)
			if err != nil {
				return nil, err
			}

			decodedRawTx, err := p.Qtum.DecodeRawTransaction(txData.Hex)
			if err != nil {
				return nil, errors.Wrap(err, "Qtum#DecodeRawTransaction")
			}

			ethTx := eth.GetTransactionByHashResponse{
				Hash:             utils.AddHexPrefix(txData.Txid),
				Nonce:            "0x01",
				BlockHash:        utils.AddHexPrefix(txData.Blockhash),
				BlockNumber:      hexutil.EncodeUint64(uint64(blockHeaderResp.Height)),
				TransactionIndex: hexutil.EncodeUint64(uint64(i)),
				From:             "0x0000000000000000000000000000000000000000",
				To:               "0x0000000000000000000000000000000000000000",
				Value:            ethVal,
				GasPrice:         hexutil.EncodeUint64(txData.Fee.BigInt().Uint64()),
				Gas:              "0x01",
				Input:            "0x00",
			}

			var invokeInfo *qtum.ContractInvokeInfo

			// We assume that this tx is a contract invokation (create or call), if we can
			// find a create or call script.
			for _, out := range decodedRawTx.Vout {
				script := strings.Split(out.ScriptPubKey.Asm, " ")
				finalOp := script[len(script)-1]

				// switch out.ScriptPubKey.Type
				switch finalOp {
				case "OP_CALL":
					info, err := qtum.ParseCallSenderASM(script)
					// OP_CALL with OP_SENDER has the script type "nonstandard"
					if err != nil {
						return nil, err
					}

					invokeInfo = info

					break
				case "OP_CREATE":
					// OP_CALL with OP_SENDER has the script type "create_sender"
					info, err := qtum.ParseCreateSenderASM(script)
					if err != nil {
						return nil, err
					}

					invokeInfo = info

					break
				}
			}

			if invokeInfo != nil {
				ethTx.From = utils.AddHexPrefix(invokeInfo.From)
				ethTx.Gas = utils.AddHexPrefix(invokeInfo.GasLimit) // not really "gas sent by user", but ¯\_(ツ)_/¯
				ethTx.GasPrice = utils.AddHexPrefix(invokeInfo.GasPrice)
				ethTx.Input = utils.AddHexPrefix(invokeInfo.CallData)

				// receipt, err := p.Qtum.GetTransactionReceipt(txData.Txid)
				// if err != nil && err != qtum.EmptyResponseErr {
				// 	return nil, err
				// }

				// if receipt != nil {
				// 	ethTx.BlockNumber = hexutil.EncodeUint64(receipt.BlockNumber)
				// 	ethTx.TransactionIndex = hexutil.EncodeUint64(receipt.TransactionIndex)

				// 	if receipt.ContractAddress != "0000000000000000000000000000000000000000" {
				// 		ethTx.To = utils.AddHexPrefix(receipt.ContractAddress)
				// 	}
				// }
			}
			txs = append(txs, ethTx)
		}

		return &eth.GetBlockByNumberResponse{
			Hash:             utils.AddHexPrefix(blockHeaderResp.Hash),
			Nonce:            utils.AddHexPrefix(nonce),
			Number:           hexutil.EncodeUint64(uint64(blockHeaderResp.Height)),
			ParentHash:       utils.AddHexPrefix(blockHeaderResp.Previousblockhash),
			Difficulty:       hexutil.EncodeUint64(uint64(blockHeaderResp.Difficulty)),
			Timestamp:        hexutil.EncodeUint64(blockHeaderResp.Time),
			StateRoot:        utils.AddHexPrefix(blockHeaderResp.HashStateRoot),
			Size:             hexutil.EncodeUint64(uint64(blockResp.Size)),
			Transactions:     txs,
			TransactionsRoot: utils.AddHexPrefix(blockResp.Merkleroot),
			ReceiptsRoot:     utils.AddHexPrefix(blockResp.Merkleroot),

			ExtraData:       "0x00",
			Miner:           "0x0000000000000000000000000000000000000000",
			TotalDifficulty: "0x00",
			GasLimit:        "0x00",
			GasUsed:         "0x00",
			LogsBloom:       "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",

			Sha3Uncles: "0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347",
			Uncles:     []string{},
		}, nil
	}
}
