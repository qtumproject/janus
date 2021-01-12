package transformer

import (
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/pkg/errors"
	"github.com/qtumproject/janus/pkg/eth"
	"github.com/qtumproject/janus/pkg/qtum"
	"github.com/qtumproject/janus/pkg/utils"
)

// ProxyETHGetBlockByHash implements ETHProxy
type ProxyETHGetBlockByHash struct {
	*qtum.Qtum
}

func (p *ProxyETHGetBlockByHash) Method() string {
	return "eth_getBlockByHash"
}

func (p *ProxyETHGetBlockByHash) Request(rawreq *eth.JSONRPCRequest) (interface{}, error) {
	var req eth.GetBlockByHashRequest
	if err := unmarshalRequest(rawreq.Params, &req); err != nil {
		return nil, err
	}

	qtumreq := p.ToRequest(&req)

	return p.request(qtumreq)
}
func (p *ProxyETHGetBlockByHash) request(req *eth.GetBlockByHashRequest) (*eth.GetBlockByHashResponse, error) {
	blockHeaderResp, err := p.GetBlockHeader(req.BlockHash)
	if err != nil {
		return nil, errors.WithMessage(err, "couldn't get block header")
	}

	// TODO: Correct to normal values
	// ? What is the correct value
	if blockHeaderResp.Previousblockhash == "" {
		blockHeaderResp.Previousblockhash = "0000000000000000000000000000000000000000000000000000000000000000"
	}

	// TODO: Correct translation into hex
	// ? What is the correct value
	nonce := hexutil.EncodeUint64(uint64(blockHeaderResp.Nonce))
	if len(strings.TrimLeft(nonce, "0x")) < 16 {
		res := strings.TrimLeft(nonce, "0x")
		for i := 0; i < 16-len(res); {
			res = "0" + res
		}
		nonce = res
	}

	blockResp, err := p.GetBlock(string(req.BlockHash))
	if err != nil {
		return nil, errors.WithMessage(err, "couldn't get block")
	}

	result := &eth.GetBlockByHashResponse{
		// TODO: If ETH block has pending status, then number must be null
		// ? Is it possible case for Qtum
		Number: hexutil.EncodeUint64(uint64(blockHeaderResp.Height)),

		// TODO: If ETH block has pending status, then hash must be null
		// ? Is it possible case for Qtum
		Hash: utils.AddHexPrefix(req.BlockHash),

		// TODO: see related TODO above
		ParentHash: utils.AddHexPrefix(blockHeaderResp.Previousblockhash),

		// TODO: If ETH block has pending status, then nonce must be null
		// ? Is it possible case for Qtum
		Nonce: utils.AddHexPrefix(nonce),

		// TODO: discuss/research
		// ? What is it
		// ! Not found
		Sha3Uncles: "0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347",

		// TODO: discuss/research
		// ? What is it
		// ! Not found, but https://docs.qtum.site/en/Qtum-RPC-API/#gettransactionreceipt
		// ? May we use the method above to calculate value. If so, what is the process of calculation
		//
		// TODO: If ETH block has pending status, then nonce must be null
		// ? Is it possible case for Qtum
		LogsBloom: "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",

		TransactionsRoot: utils.AddHexPrefix(blockResp.Merkleroot),
		StateRoot:        utils.AddHexPrefix(blockHeaderResp.HashStateRoot),

		// TODO: discuss
		// ! Not found, but https://docs.qtum.site/en/Qtum-RPC-API/#gettransactionreceipt
		// ? May we use the method above to get value. Request data for 1-st tx
		ReceiptsRoot: utils.AddHexPrefix(blockResp.Merkleroot),

		// TODO: research
		// ! Not found, buy https://docs.qtum.site/en/Qtum-RPC-API/#getblockstats
		Miner: "0x0000000000000000000000000000000000000000",

		// TODO: discuss
		// ? ETH value is amboguous
		Difficulty: hexutil.EncodeUint64(uint64(blockHeaderResp.Difficulty)),

		// TODO: discuss
		// ? May we request each block in the chain to calculate value
		TotalDifficulty: "0x00",

		// TODO: discuss / research
		// ? What is it
		// ? Possibly, should be always empty
		ExtraData: "0x00",

		Size: hexutil.EncodeUint64(uint64(blockResp.Size)),

		// TODO: discuss
		// ! Found only for contracts
		GasLimit: "0x00",
		GasUsed:  "0x00",
		//
		// lastTxHash := blockResp.Tx[len(blockResp.Tx)-1]
		// receipt, err := p.GetTransactionReceipt(lastTxHash)
		// if err != nil {
		// 	return nil, errors.WithMessage(err, "couldn't get receipt of the last transaction")
		// }
		// result.GasUsed = hexutil.EncodeUint64(receipt.CumulativeGasUsed)

		Timestamp:    hexutil.EncodeUint64(blockHeaderResp.Time),
		Transactions: make([]interface{}, 0, len(blockResp.Tx)),

		// TODO: discuss
		// ! Not found
		Uncles: []string{},
	}

	if req.FullTransaction {
		for i, txHash := range blockResp.Tx {
			// TODO: discuss
			// ? Is it true for genezis block
			// ? Use block header
			if blockHeaderResp.Height == 0 {
				break
			}

			tx, err := GetTransactionByHash(p.Qtum, txHash, blockHeaderResp.Height, i)
			if err != nil {
				return nil, errors.WithMessage(err, "couldn't get transaction by hash")
			}
			result.Transactions = append(result.Transactions, *tx)
		}
	} else {
		for _, txHash := range blockResp.Tx {
			// TODO: discuss
			// ? Tx string string of [32]byte in []string, would like 0x + hash30(string)
			// 	* Qtum txHash length is [64]byte
			result.Transactions = append(result.Transactions, utils.AddHexPrefix(txHash))
		}
	}

	return result, nil
}

func (p *ProxyETHGetBlockByHash) ToRequest(ethreq *eth.GetBlockByHashRequest) *eth.GetBlockByHashRequest {
	return &eth.GetBlockByHashRequest{
		BlockHash:       utils.RemoveHexPrefix(strings.Trim(ethreq.BlockHash, "\"")),
		FullTransaction: ethreq.FullTransaction,
	}
}
