package transformer

import (
	"context"
	"fmt"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
	"github.com/qtumproject/janus/pkg/blockhash"
	"github.com/qtumproject/janus/pkg/eth"
	"github.com/qtumproject/janus/pkg/qtum"
	"github.com/qtumproject/janus/pkg/utils"
)

var ErrBlockHashNotConfigured = errors.New("BlockHash database not configured")
var ErrBlockHashUnknown = errors.New("BlockHash unknown")

// ProxyETHGetBlockByHash implements ETHProxy
type ProxyETHGetBlockByHash struct {
	*qtum.Qtum
}

func (p *ProxyETHGetBlockByHash) Method() string {
	return "eth_getBlockByHash"
}

func (p *ProxyETHGetBlockByHash) Request(rawreq *eth.JSONRPCRequest, c echo.Context) (interface{}, eth.JSONRPCError) {
	req := new(eth.GetBlockByHashRequest)
	if err := unmarshalRequest(rawreq.Params, req); err != nil {
		// TODO: Correct error code?
		return nil, eth.NewInvalidParamsError(err.Error())
	}

	blockHash := c.Get("blockHash")
	bh, ok := blockHash.(*blockhash.BlockHash)
	if !ok {
		// ok, do nothing
	}

	req.BlockHash = utils.RemoveHexPrefix(req.BlockHash)

	resultChan := make(chan *eth.GetBlockByHashResponse, 2)
	errorChan := make(chan eth.JSONRPCError, 1)
	qtumBlockErrorChan := make(chan error, 1)
	ctx, cancelFunc := context.WithCancel(context.Background())
	defer cancelFunc()

	go func() {
		result, err := p.request(ctx, req)
		if err != nil {
			errorChan <- err
			return
		}

		resultChan <- result
	}()

	if bh == nil {
		qtumBlockErrorChan <- ErrBlockHashNotConfigured
	} else {
		go func() {
			qtumBlockHash, err := bh.GetQtumBlockHashContext(ctx, req.BlockHash)
			if err != nil {
				qtumBlockErrorChan <- err
				return
			}

			if qtumBlockHash == nil {
				qtumBlockErrorChan <- ErrBlockHashUnknown
				return
			}

			request := &eth.GetBlockByHashRequest{
				BlockHash:       utils.RemoveHexPrefix(*qtumBlockHash),
				FullTransaction: req.FullTransaction,
			}

			result, jsonErr := p.request(ctx, request)
			if jsonErr != nil {
				qtumBlockErrorChan <- jsonErr.Error()
				return
			}

			resultChan <- result
		}()
	}

	select {
	case result := <-resultChan:
		// TODO: Stop remaining request
		if result == nil {
			select {
			case result := <-resultChan:
				// backup succeeded
				return result, nil
			case <-qtumBlockErrorChan:
				// backup failed, return original request
				return nil, nil
			}
		} else {
			return result, nil
		}
	case err := <-errorChan:
		// the main request failed, wait for backup to finish
		select {
		case result := <-resultChan:
			// backup succeeded
			return result, nil
		case <-qtumBlockErrorChan:
			// backup failed, return original request
			return nil, err
		}
	}
}

func (p *ProxyETHGetBlockByHash) request(ctx context.Context, req *eth.GetBlockByHashRequest) (*eth.GetBlockByHashResponse, eth.JSONRPCError) {
	blockHeader, err := p.GetBlockHeader(ctx, req.BlockHash)
	if err != nil {
		if err == qtum.ErrInvalidAddress {
			// unknown block hash should return {result: null}
			p.GetDebugLogger().Log("msg", "Unknown block hash", "blockHash", req.BlockHash)
			return nil, nil
		}
		p.GetDebugLogger().Log("msg", "couldn't get block header", "blockHash", req.BlockHash)
		return nil, eth.NewCallbackError("couldn't get block header")
	}
	block, err := p.GetBlock(ctx, req.BlockHash)
	if err != nil {
		p.GetDebugLogger().Log("msg", "couldn't get block", "blockHash", req.BlockHash)
		return nil, eth.NewCallbackError("couldn't get block")
	}
	nonce := hexutil.EncodeUint64(uint64(block.Nonce))
	// left pad nonce with 0 to length 16, eg: 0x0000000000000042
	nonce = utils.AddHexPrefix(fmt.Sprintf("%016v", utils.RemoveHexPrefix(nonce)))
	resp := &eth.GetBlockByHashResponse{
		// TODO: researching
		// * If ETH block has pending status, then the following values must be null
		// ? Is it possible case for Qtum
		Hash:   utils.AddHexPrefix(req.BlockHash),
		Number: hexutil.EncodeUint64(uint64(block.Height)),

		// TODO: researching
		// ! Not found
		// ! Has incorrect value for compatability
		ReceiptsRoot: utils.AddHexPrefix(block.Merkleroot),

		// TODO: researching
		// ! Not found
		// ! Probably, may be calculated by huge amount of requests
		TotalDifficulty: hexutil.EncodeUint64(uint64(blockHeader.Difficulty)),

		// TODO: researching
		// ! Not found
		// ? Expect it always to be null
		Uncles: []string{},

		// TODO: check value correctness
		Sha3Uncles: eth.DefaultSha3Uncles,

		// TODO: backlog
		// ! Not found
		// - Temporary expect this value to be always zero, as Etherium logs are usually zeros
		LogsBloom: eth.EmptyLogsBloom,

		// TODO: researching
		// ? What value to put
		// - Temporary set this value to be always zero
		// - the graph requires this to be of length 64
		ExtraData: "0x0000000000000000000000000000000000000000000000000000000000000000",

		Nonce:            nonce,
		Size:             hexutil.EncodeUint64(uint64(block.Size)),
		Difficulty:       hexutil.EncodeUint64(uint64(blockHeader.Difficulty)),
		StateRoot:        utils.AddHexPrefix(blockHeader.HashStateRoot),
		TransactionsRoot: utils.AddHexPrefix(block.Merkleroot),
		Transactions:     make([]interface{}, 0, len(block.Txs)),
		Timestamp:        hexutil.EncodeUint64(blockHeader.Time),
	}

	if blockHeader.IsGenesisBlock() {
		resp.ParentHash = "0x0000000000000000000000000000000000000000000000000000000000000000"
		resp.Miner = utils.AddHexPrefix(qtum.ZeroAddress)
	} else {
		resp.ParentHash = utils.AddHexPrefix(blockHeader.Previousblockhash)
		// ! Not found
		//
		// NOTE:
		// 	In order to find a miner it seems, that we have to check
		// 	address field of the txout method response. Current
		// 	suggestion is to fill this field with zeros, not to
		// 	spend much time on requests execution
		//
		// TODO: check if it's value is acquirable via logs
		resp.Miner = "0x0000000000000000000000000000000000000000"
	}

	// TODO: rethink later
	// ! Found only for contracts transactions
	// As there is no gas values presented at common block info, we set
	// gas limit value equalling to default gas limit of a block
	resp.GasLimit = utils.AddHexPrefix(qtum.DefaultBlockGasLimit)
	resp.GasUsed = "0x0"

	// TODO: Future improvement: If getBlock is called with verbosity 2 it also returns full tx info as if getRawTransaction was called for each,
	// so using that from the start instead of requesting each tx individually as done here would save a lot of back-and-forth

	if req.FullTransaction {
		for _, txHash := range block.Txs {
			tx, err := getTransactionByHash(ctx, p.Qtum, txHash)
			if err != nil {
				p.GetDebugLogger().Log("msg", "Couldn't get transaction by hash", "hash", txHash, "err", err)
				return nil, eth.NewCallbackError("couldn't get transaction by hash")
			}
			if tx == nil {
				if block.Height == 0 {
					// Error Invalid address - The genesis block coinbase is not considered an ordinary transaction and cannot be retrieved
					// the coinbase we can ignore since its not a real transaction, mainnet ethereum also doesn't return any data about the genesis coinbase
					p.GetDebugLogger().Log("msg", "Failed to get transaction in genesis block, probably the coinbase which we can't get")
				} else {
					p.GetDebugLogger().Log("msg", "Failed to get transaction by hash included in a block", "hash", txHash)
					if !p.GetFlagBool(qtum.FLAG_IGNORE_UNKNOWN_TX) {
						return nil, eth.NewCallbackError("couldn't get transaction by hash included in a block")
					}
				}
			} else {
				resp.Transactions = append(resp.Transactions, *tx)
			}
			// TODO: fill gas used
			// TODO: fill gas limit?
		}
	} else {
		for _, txHash := range block.Txs {
			// NOTE:
			// 	Etherium RPC API doc says, that tx hashes must be of [32]byte,
			// 	however it doesn't seem to be correct, 'cause Etherium tx hash
			// 	has [64]byte just like Qtum tx hash has. In this case we do no
			// 	additional convertations now, while everything works fine
			resp.Transactions = append(resp.Transactions, utils.AddHexPrefix(txHash))
		}
	}

	return resp, nil
}
