package transformer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"math/big"
	"strconv"

	"github.com/qtumproject/janus/pkg/eth"
	"github.com/qtumproject/janus/pkg/qtum"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/pkg/errors"
	"github.com/qtumproject/janus/pkg/utils"
	"github.com/shopspring/decimal"
)

type EthGas interface {
	GasHex() string
	GasPriceHex() string
}

func EthGasToQtum(g EthGas) (gasLimit *big.Int, gasPrice string, err error) {
	gasLimit = g.(*eth.SendTransactionRequest).Gas.Int

	gasPriceDecimal, err := EthValueToQtumAmount(g.GasPriceHex())
	if err != nil {
		return nil, "0.0", err
	}
	gasPrice = fmt.Sprintf("%v", gasPriceDecimal)

	return
}

func EthValueToQtumAmount(val string) (decimal.Decimal, error) {
	if val == "" {
		return decimal.NewFromFloat(0.0000004), nil
	}

	ethVal, err := utils.DecodeBig(val)
	if err != nil {
		return decimal.NewFromFloat(0.0), err
	}

	ethValDecimal, err := decimal.NewFromString(ethVal.String())
	if err != nil {
		return decimal.NewFromFloat(0.0), errors.New("decimal.NewFromString was not a success")
	}

	amount := ethValDecimal.Mul(decimal.NewFromFloat(float64(1e-8)))

	return amount, nil
}

func formatQtumAmount(amount decimal.Decimal) (string, error) {
	decimalAmount := amount.Mul(decimal.NewFromFloat(float64(1e8)))

	//convert decimal to Integer
	result := decimalAmount.BigInt()

	if !decimalAmount.Equals(decimal.NewFromBigInt(result, 0)) {
		return "0x0", errors.New("decimal.BigInt() was not a success")
	}

	return hexutil.EncodeBig(result), nil
}

func unmarshalRequest(data []byte, v interface{}) error {
	if err := json.Unmarshal(data, v); err != nil {
		return errors.Wrap(err, "Invalid RPC input")
	}
	return nil
}

// NOTE:
// 	- is not for reward transactions
// 	- Vin[i].N (vout number) -> get Transaction(txID).Vout[N].Address
// 	- returning address already has 0x prefix
func getNonContractTxSenderAddress(p *qtum.Qtum, vins []*qtum.DecodedRawTransactionInV) (string, error) {
	for _, vin := range vins {
		prevQtumTx, err := p.GetRawTransaction(vin.TxID, false)
		if err != nil {
			return "", errors.WithMessage(err, "couldn't get vin's previous transaction")
		}
		for _, out := range prevQtumTx.Vouts {
			for _, address := range out.Details.Addresses {
				return utils.AddHexPrefix(address), nil
			}
		}
	}
	return "", errors.New("not found")
}

// NOTE:
// 	- is not for reward transactions
// 	- returning address already has 0x prefix
//
// 	TODO: researching
// 	- Vout[0].Addresses[0] - temporary solution
func findNonContractTxReceiverAddress(vouts []*qtum.DecodedRawTransactionOutV) (string, error) {
	for _, vout := range vouts {
		for _, address := range vout.ScriptPubKey.Addresses {
			return utils.AddHexPrefix(address), nil
		}
	}
	return "", errors.New("not found")
}

func getBlockNumberByHash(p *qtum.Qtum, hash string) (uint64, error) {
	block, err := p.GetBlock(hash)
	if err != nil {
		return 0, errors.WithMessage(err, "couldn't get block")
	}
	return uint64(block.Height), nil
}

func getTransactionIndexInBlock(p *qtum.Qtum, txHash string, blockHash string) (int64, error) {
	block, err := p.GetBlock(blockHash)
	if err != nil {
		return -1, errors.WithMessage(err, "couldn't get block")
	}
	for i, blockTx := range block.Txs {
		if txHash == blockTx {
			return int64(i), nil
		}
	}
	return -1, errors.New("not found")
}

func formatQtumNonce(nonce int) string {
	var (
		hexedNonce     = strconv.FormatInt(int64(nonce), 16)
		missedCharsNum = 16 - len(hexedNonce)
	)
	for i := 0; i < missedCharsNum; i++ {
		hexedNonce = "0" + hexedNonce
	}
	return "0x" + hexedNonce
}

// Returns Qtum block number. Result depends on a passed raw param. Raw param's slice of bytes should
// has one of the following values:
// 	- hex string representation of a number of a specific block
// 	- string "latest" - for the latest mined block
// 	- string "earliest" for the genesis block
// 	- string "pending" - for the pending state/transactions
func getBlockNumberByParam(p *qtum.Qtum, rawParam json.RawMessage, defaultVal int64) (*big.Int, error) {
	if len(rawParam) < 1 {
		return nil, errors.Errorf("empty parameter value")
	}
	if !isBytesOfString(rawParam) {
		return nil, errors.Errorf("invalid parameter format - string is expected")
	}

	param := string(rawParam[1 : len(rawParam)-1]) // trim \" runes
	switch param {
	case "latest":
		res, err := p.GetBlockChainInfo()
		if err != nil {
			return nil, err
		}
		return big.NewInt(res.Blocks), nil

	case "earliest":
		// TODO: discuss
		// ! Genesis block cannot be retreived
		return big.NewInt(0), nil

	case "pending":
		// TODO: discuss
		// 	! Researching
		return nil, errors.New("TODO: tag is in implementation")

	default: // hex number
		n, err := utils.DecodeBig(param)
		if err != nil {
			return nil, errors.Wrap(err, "couldn't decode hex number to big int")
		}
		return n, nil
	}
}

func isBytesOfString(v json.RawMessage) bool {
	println(string(v))
	dQuote := []byte{'"'}
	if !bytes.HasPrefix(v, dQuote) && !bytes.HasSuffix(v, dQuote) {
		return false
	}
	if bytes.Count(v, dQuote) != 2 {
		return false
	}
	// TODO: decide
	// ? Should we iterate over v to check if v[1:len(v)-2] is in a range of a-A, z-Z, 0-9
	return true
}

func extractETHLogsFromTransactionReceipt(receipt *qtum.TransactionReceipt) []eth.Log {
	logs := make([]eth.Log, 0, len(receipt.Log))
	for i, log := range receipt.Log {
		topics := make([]string, 0, len(log.Topics))
		for _, topic := range log.Topics {
			topics = append(topics, utils.AddHexPrefix(topic))
		}
		logs = append(logs, eth.Log{
			TransactionHash:  utils.AddHexPrefix(receipt.TransactionHash),
			TransactionIndex: hexutil.EncodeUint64(receipt.TransactionIndex),
			BlockHash:        utils.AddHexPrefix(receipt.BlockHash),
			BlockNumber:      hexutil.EncodeUint64(receipt.BlockNumber),
			Data:             utils.AddHexPrefix(log.Data),
			Address:          utils.AddHexPrefix(log.Address),
			Topics:           topics,
			LogIndex:         hexutil.EncodeUint64(uint64(i)),
		})
	}
	return logs
}
