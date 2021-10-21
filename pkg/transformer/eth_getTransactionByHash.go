package transformer

import (
	"encoding/json"

	"github.com/ethereum/go-ethereum/common/hexutil"
	"github.com/labstack/echo"
	"github.com/pkg/errors"
	"github.com/qtumproject/janus/pkg/eth"
	"github.com/qtumproject/janus/pkg/qtum"
	"github.com/qtumproject/janus/pkg/utils"
)

// ProxyETHGetTransactionByHash implements ETHProxy
type ProxyETHGetTransactionByHash struct {
	*qtum.Qtum
}

func (p *ProxyETHGetTransactionByHash) Method() string {
	return "eth_getTransactionByHash"
}

func (p *ProxyETHGetTransactionByHash) Request(req *eth.JSONRPCRequest, c echo.Context) (interface{}, error) {
	var txHash eth.GetTransactionByHashRequest
	if err := json.Unmarshal(req.Params, &txHash); err != nil {
		return nil, errors.Wrap(err, "couldn't unmarshal request")
	}
	if txHash == "" {
		return nil, errors.New("transaction hash is empty")
	}

	qtumReq := &qtum.GetTransactionRequest{
		TxID: utils.RemoveHexPrefix(string(txHash)),
	}
	return p.request(qtumReq)
}

func (p *ProxyETHGetTransactionByHash) request(req *qtum.GetTransactionRequest) (*eth.GetTransactionByHashResponse, error) {
	ethTx, err := getTransactionByHash(p.Qtum, req.TxID)
	if err != nil {
		return nil, err
	}
	return ethTx, nil
}

// TODO: think of returning flag if it's a reward transaction for miner
func getTransactionByHash(p *qtum.Qtum, hash string) (*eth.GetTransactionByHashResponse, error) {
	qtumTx, err := p.GetTransaction(hash)
	if err != nil {
		if errors.Cause(err) != qtum.ErrInvalidAddress {
			return nil, err
		}
		ethTx, err := getRewardTransactionByHash(p, hash)
		if err != nil {
			if errors.Cause(err) == qtum.ErrInvalidAddress {
				return nil, nil
			}
			rawTx, err := p.GetRawTransaction(hash, false)
			if err != nil {
				if errors.Cause(err) == qtum.ErrInvalidAddress {
					return nil, nil
				}
				return nil, err
			} else {
				qtumTx = &qtum.GetTransactionResponse{
					BlockHash:  rawTx.BlockHash,
					BlockIndex: 1, // TODO: Possible to get this somewhere?
					Hex:        rawTx.Hex,
				}
			}
		}
		return ethTx, nil
	}
	qtumDecodedRawTx, err := p.DecodeRawTransaction(qtumTx.Hex)
	if err != nil {
		return nil, errors.WithMessage(err, "couldn't get raw transaction")
	}

	ethTx := &eth.GetTransactionByHashResponse{
		Hash:  utils.AddHexPrefix(qtumDecodedRawTx.ID),
		Nonce: "0x0",

		// TODO: researching
		// ? Do we need those values
		V: "",
		R: "",
		S: "",
	}

	if !qtumTx.IsPending() { // otherwise, the following values must be nulls
		blockNumber, err := getBlockNumberByHash(p, qtumTx.BlockHash)
		if err != nil {
			return nil, errors.WithMessage(err, "couldn't get block number by hash")
		}
		ethTx.BlockNumber = hexutil.EncodeUint64(blockNumber)
		ethTx.BlockHash = utils.AddHexPrefix(qtumTx.BlockHash)
		ethTx.TransactionIndex = hexutil.EncodeUint64(uint64(qtumTx.BlockIndex))
	}

	ethAmount, err := formatQtumAmount(qtumDecodedRawTx.CalcAmount())
	if err != nil {
		return nil, errors.WithMessage(err, "couldn't format amount")
	}
	ethTx.Value = ethAmount

	qtumTxContractInfo, isContractTx, err := qtumDecodedRawTx.ExtractContractInfo()
	if err != nil {
		return nil, errors.WithMessage(err, "couldn't extract contract info")
	}
	if isContractTx {
		// TODO: research is this allowed? ethTx.Input = utils.AddHexPrefix(qtumTxContractInfo.UserInput)
		if qtumTxContractInfo.UserInput == "" {
			ethTx.Input = "0x0"
		} else {
			ethTx.Input = utils.AddHexPrefix(qtumTxContractInfo.UserInput)
		}
		ethTx.From = utils.AddHexPrefix(qtumTxContractInfo.From)
		//TODO: research if 'To' adress could be other than zero address when 'isContractTx == TRUE'
		if len(qtumTxContractInfo.To) == 0 {
			ethTx.To = utils.AddHexPrefix(qtum.ZeroAddress)

		} else {
			ethTx.To = utils.AddHexPrefix(qtumTxContractInfo.To)

		}
		ethTx.Gas = hexutil.Encode([]byte(qtumTxContractInfo.GasLimit))
		ethTx.GasPrice = hexutil.Encode([]byte(qtumTxContractInfo.GasPrice))

		return ethTx, nil
	}

	if qtumTx.Generated {
		ethTx.From = utils.AddHexPrefix(qtum.ZeroAddress)
	} else {
		// TODO: Figure out proper way to do this
		// There is a problem with this function, sometimes it returns errors on regtest, empty block?
		// commenting it out as its being overwritten below anyway
		/*
			ethTx.From, err = getNonContractTxSenderAddress(p, qtumDecodedRawTx.Vins)
			if err != nil {
				return nil, errors.WithMessage(err, "couldn't get non contract transaction sender address")
			}
		*/
		// TODO: discuss
		// ? Does func above return incorrect address for graph-node (len is < 40)
		// ! Temporary solution
		ethTx.From = utils.AddHexPrefix(qtum.ZeroAddress)
	}
	ethTx.To, err = findNonContractTxReceiverAddress(qtumDecodedRawTx.Vouts)
	if err != nil {
		// TODO: discuss, research
		// ? Some vouts doesn't have `receive` category at all
		ethTx.To = utils.AddHexPrefix(qtum.ZeroAddress)

		// TODO: uncomment, after todo above will be resolved
		// return nil, errors.WithMessage(err, "couldn't get non contract transaction receiver address")
	}
	// TODO: discuss
	// ? Does func above return incorrect address for graph-node (len is < 40)
	// ! Temporary solution
	ethTx.To = utils.AddHexPrefix(qtum.ZeroAddress)

	// TODO: researching
	// ! Temporary solution
	//	if len(qtumTx.Hex) == 0 {
	//		ethTx.Input = "0x0"
	//	} else {
	//		ethTx.Input = utils.AddHexPrefix(qtumTx.Hex)
	//	}
	ethTx.Input = utils.AddHexPrefix(qtumTx.Hex)

	// TODO: researching
	// ? Is it correct for non contract transaction
	ethTx.Gas = "0x0"
	ethTx.GasPrice = "0x0"

	return ethTx, nil
}

// TODO: discuss
// ? There are `witness` transactions, that is not acquireable nither via `gettransaction`, nor `getrawtransaction`
func getRewardTransactionByHash(p *qtum.Qtum, hash string) (*eth.GetTransactionByHashResponse, error) {
	rawQtumTx, err := p.GetRawTransaction(hash, false)
	if err != nil {
		return nil, errors.WithMessage(err, "couldn't get raw reward transaction")
	}

	ethTx := &eth.GetTransactionByHashResponse{
		Hash:  utils.AddHexPrefix(hash),
		Nonce: "0x0",

		// TODO: discuss
		// ? Expect this value to be always zero
		// Geth returns 0x if there is no input data for a transaction
		Input: "0x",

		// TODO: discuss
		// ? Are zero values applicable
		Gas:      "0x0",
		GasPrice: "0x0",

		// TODO: researching
		// ? Do we need those values
		V: "",
		R: "",
		S: "",
	}

	if !rawQtumTx.IsPending() {
		blockIndex, err := getTransactionIndexInBlock(p, hash, rawQtumTx.BlockHash)
		if err != nil {
			return nil, errors.WithMessage(err, "couldn't get transaction index in block")
		}
		ethTx.TransactionIndex = hexutil.EncodeUint64(uint64(blockIndex))

		blockNumber, err := getBlockNumberByHash(p, rawQtumTx.BlockHash)
		if err != nil {
			return nil, errors.WithMessage(err, "couldn't get block number by hash")
		}
		ethTx.BlockNumber = hexutil.EncodeUint64(blockNumber)

		ethTx.BlockHash = utils.AddHexPrefix(rawQtumTx.BlockHash)
	}

	for i := range rawQtumTx.Vouts {
		// TODO: discuss
		// ! The response may be null, even if txout is presented
		_, err := p.GetTransactionOut(hash, i, rawQtumTx.IsPending())
		if err != nil {
			return nil, errors.WithMessage(err, "couldn't get transaction out")
		}
		// TODO: discuss, researching
		// ? Where is a reward amount
		ethTx.Value = "0x0"
	}

	// TODO: discuss
	// ? Do we have to set `from` == `0x00..00`
	ethTx.From = utils.AddHexPrefix(qtum.ZeroAddress)

	// I used Base58AddressToHex at the moment
	// because convertQtumAddress functions causes error for
	// P2Sh address(such as MUrenj2sPqEVTiNbHQ2RARiZYyTAAeKiDX) and BECH32 address (such as qc1qkt33x6hkrrlwlr6v59wptwy6zskyrjfe40y0lx)
	if rawQtumTx.OP_SENDER != "" {
		// addr, err := convertQtumAddress(rawQtumTx.OP_SENDER)
		addr, err := p.Base58AddressToHex(rawQtumTx.OP_SENDER)
		if err == nil {
			ethTx.From = utils.AddHexPrefix(addr)
		}
	} else if len(rawQtumTx.Vins) > 0 && rawQtumTx.Vins[0].Address != "" {
		// addr, err := convertQtumAddress(rawQtumTx.Vins[0].Address)
		addr, err := p.Base58AddressToHex(rawQtumTx.Vins[0].Address)
		if err == nil {
			ethTx.From = utils.AddHexPrefix(addr)
		}
	}
	// TODO: discuss
	// ? Where is a `to`
	ethTx.To = utils.AddHexPrefix(qtum.ZeroAddress)

	return ethTx, nil
}
