package transformer

import (
	"strings"

	"github.com/ethereum/go-ethereum/common/hexutil"
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

	// TODO: Remove repetition
	if !req.FullTransaction {
		txs := make([]string, 0, len(blockResp.Tx))
		for _, tx := range blockResp.Tx {
			txs = append(txs, utils.AddHexPrefix(tx))
		}

		/// TODO: Correct to normal values
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

			/// TODO: Correct to normal values
			ethTx, err := p.GetTransactionByHash(tx, blockHeaderResp.Height, i)
			if err != nil {
				return nil, err
			}

			txs = append(txs, *ethTx)
		}

		/// TODO: Correct to normal values
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
