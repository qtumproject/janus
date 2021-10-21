package eth

import (
	"bytes"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/pkg/errors"
	"github.com/qtumproject/janus/pkg/utils"
	"github.com/shopspring/decimal"
)

var DefaultGasAmountForQtum = big.NewInt(250000)

// QTUM default gas value (also the minimum gas) in wei
var DefaultGasPriceInWei = big.NewInt(40000000000)

type (
	SendTransactionResponse string

	// SendTransactionRequest eth_sendTransaction
	SendTransactionRequest struct {
		From     string  `json:"from"`
		To       string  `json:"to"`
		Gas      *ETHInt `json:"gas"`      // optional
		GasPrice *ETHInt `json:"gasPrice"` // optional
		Value    string  `json:"value"`    // optional
		Data     string  `json:"data"`     // optional
		Nonce    string  `json:"nonce"`    // optional
	}
)

func (r *SendTransactionRequest) UnmarshalJSON(data []byte) error {
	type Request SendTransactionRequest

	var params []Request
	if err := json.Unmarshal(data, &params); err != nil {
		return err
	}

	*r = SendTransactionRequest(params[0])

	if r.Gas == nil {
		// ETH: (optional, default: 90000) Integer of the gas provided for the transaction execution. It will return unused gas.
		// QTUM: (numeric or string, optional) gasLimit, default: 250000, max: 40000000
		r.Gas = &ETHInt{DefaultGasAmountForQtum}
	}

	if r.GasPrice == nil {
		// ETH: (optional, default: To-Be-Determined) Integer of the gasPrice used for each paid gas
		// QTUM: (numeric or string, optional) gasPrice Qtum price per gas unit, default: 0.0000004, min:0.0000004
		r.GasPrice = &ETHInt{DefaultGasPriceInWei}
	}

	return nil
}

// see: https://ethereum.stackexchange.com/questions/8384/transfer-an-amount-between-two-ethereum-accounts-using-json-rpc
func (t *SendTransactionRequest) IsSendEther() bool {
	// data must be empty
	return t.Value != "" && t.To != "" && t.From != "" && t.Data == ""
}

func (t *SendTransactionRequest) IsCreateContract() bool {
	return t.To == "" && t.Data != ""
}

func (t *SendTransactionRequest) IsCallContract() bool {
	return t.To != "" && t.Data != ""
}

func (t *SendTransactionRequest) GasHex() string {
	if t.Gas == nil {
		return ""
	}

	return t.Gas.Hex()
}

func (t *SendTransactionRequest) GasPriceHex() string {
	if t.GasPrice == nil {
		return ""
	}
	return t.GasPrice.Hex()
}

// ========== eth_sendRawTransaction ============= //

type (
	// Presents hexed string of a raw transaction
	SendRawTransactionRequest [1]string
	// Presents hexed string of a transaction hash
	SendRawTransactionResponse string
)

// CallResponse
type CallResponse string

// CallRequest eth_call
type CallRequest struct {
	From     string  `json:"from"`
	To       string  `json:"to"`
	Gas      *ETHInt `json:"gas"`      // optional
	GasPrice *ETHInt `json:"gasPrice"` // optional
	Value    string  `json:"value"`    // optional
	Data     string  `json:"data"`     // optional
}

func (t *CallRequest) GasHex() string {
	if t.Gas == nil {
		return ""
	}
	return t.Gas.Hex()
}

func (t *CallRequest) GasPriceHex() string {
	if t.GasPrice == nil {
		return ""
	}
	return t.GasPrice.Hex()
}

func (t *CallRequest) UnmarshalJSON(data []byte) error {
	var err error
	var params []json.RawMessage
	if err = json.Unmarshal(data, &params); err != nil {
		return err
	}

	if len(params) == 0 {
		return errors.New("params must be set")
	}

	type txCallObject CallRequest
	var obj txCallObject
	if err = json.Unmarshal(params[0], &obj); err != nil {
		return err
	}

	cr := CallRequest(obj)
	*t = cr
	return nil
}

type (
	PersonalUnlockAccountResponse bool
	BlockNumberResponse           string
	NetVersionResponse            string
	HashrateResponse              string
	MiningResponse                bool
)

// ========== eth_sign ============= //

type (
	SignRequest struct {
		Account string
		Message []byte
	}
	SignResponse string
)

func (t *SignRequest) UnmarshalJSON(data []byte) (err error) {
	var params []interface{}

	err = json.Unmarshal(data, &params)
	if err != nil {
		return errors.Wrap(err, "json unmarshalling")
	}

	if len(params) != 2 {
		return errors.New("expects 2 arguments")
	}

	if account, ok := params[0].(string); ok {
		t.Account = account
	} else {
		return errors.New("account address should be a hex string")
	}

	if data, ok := params[1].(string); ok {
		var msg []byte
		if !strings.HasPrefix(data, "0x") {
			msg = []byte(data)
		} else {
			msg, err = hex.DecodeString(utils.RemoveHexPrefix(data))
			if err != nil {
				return errors.Wrap(err, "invalid data format")
			}
		}

		t.Message = msg
	} else {
		return errors.New("data should be a hex string")
	}

	return nil
}

// ========== GetLogs ============= //

type (
	GetLogsRequest struct {
		FromBlock json.RawMessage `json:"fromBlock"`
		ToBlock   json.RawMessage `json:"toBlock"`
		Address   json.RawMessage `json:"address"` // string or []string
		Topics    []interface{}   `json:"topics"`
		Blockhash string          `json:"blockhash"`
	}
	GetLogsResponse []Log
)

func (r *GetLogsRequest) UnmarshalJSON(data []byte) error {
	type Request GetLogsRequest
	var params []Request
	if err := json.Unmarshal(data, &params); err != nil {
		return errors.Wrap(err, "json unmarshalling")
	}

	if len(params) == 0 {
		return errors.New("params must be set")
	}

	*r = GetLogsRequest(params[0])

	return nil
}

// ========== GetTransactionByHash ============= //
type (
	// Presents transaction hash value
	GetTransactionByHashRequest  string
	GetTransactionByHashResponse struct {
		// NOTE: must be null when its pending
		BlockHash string `json:"blockHash"`
		// NOTE: must be null when its pending
		BlockNumber string `json:"blockNumber"`

		// Hex representation of an integer - position in the block
		//
		// NOTE: must be null when its pending
		TransactionIndex string `json:"transactionIndex"`

		Hash string `json:"hash"`

		// The number of transactions made by the sender prior to this one
		// NOTE:
		// 	Unnecessary value, but keep it to be always 0x0, to be
		// 	graph-node compatible
		Nonce string `json:"nonce"`

		// Value transferred in Wei
		Value string `json:"value"`
		// The data send along with the transaction
		Input string `json:"input"`

		From string `json:"from"`
		// NOTE: must be null, if it's a contract creation transaction
		To string `json:"to"`

		// Gas provided by the sender
		Gas string `json:"gas"`
		// Gas price provided by the sender in Wei
		GasPrice string `json:"gasPrice"`

		// ECDSA recovery id
		V string `json:"v,omitempty"`
		// ECDSA signature r
		R string `json:"r,omitempty"`
		// ECDSA signature s
		S string `json:"s,omitempty"`
	}
)

func (r *GetTransactionByHashRequest) UnmarshalJSON(data []byte) error {
	var params []interface{}
	if err := json.Unmarshal(data, &params); err != nil {
		return err
	}
	if paramsNum := len(params); paramsNum != 1 {
		return fmt.Errorf("invalid parameters number - %d/1", paramsNum)
	}

	switch t := params[0].(type) {
	case string:
		*r = GetTransactionByHashRequest(t)
		return nil
	default:
		return fmt.Errorf("invalid parameter type %T, but %T is expected", t, "")
	}
}

// ========== GetTransactionByBlockHashAndIndex ========== //

type GetTransactionByBlockHashAndIndex struct {
	BlockHash        string
	TransactionIndex string
}

func (r *GetTransactionByBlockHashAndIndex) UnmarshalJSON(data []byte) error {
	var params []interface{}
	if err := json.Unmarshal(data, &params); err != nil {
		return errors.Wrap(err, "couldn't unmarhsal parameters")
	}
	paramsNum := len(params)
	if paramsNum == 0 {
		return errors.Errorf("missing value for required argument 0")
	} else if paramsNum == 1 {
		return errors.Errorf("missing value for required argument 1")
	} else if paramsNum > 2 {
		return errors.Errorf("too many arguments, want at most 2")
	}

	blockHash, ok := params[0].(string)
	if !ok {
		return newErrInvalidParameterType(1, params[0], "")
	}
	r.BlockHash = blockHash

	transactionIndex, ok := params[1].(string)
	if !ok {
		return newErrInvalidParameterType(2, params[1], "")
	}
	r.TransactionIndex = transactionIndex

	return nil
}

// ========== GetTransactionByBlockNumberAndIndex ========== //

type GetTransactionByBlockNumberAndIndex struct {
	BlockNumber      string
	TransactionIndex string
}

func (r *GetTransactionByBlockNumberAndIndex) UnmarshalJSON(data []byte) error {
	var params []interface{}
	if err := json.Unmarshal(data, &params); err != nil {
		return errors.Wrap(err, "couldn't unmarhsal parameters")
	}
	paramsNum := len(params)
	if paramsNum == 0 {
		return errors.Errorf("missing value for required argument 0")
	} else if paramsNum == 1 {
		return errors.Errorf("missing value for required argument 1")
	} else if paramsNum > 2 {
		return errors.Errorf("too many arguments, want at most 2")
	}

	blockNumber, ok := params[0].(string)
	if !ok {
		return newErrInvalidParameterType(1, params[0], "")
	}
	r.BlockNumber = blockNumber

	transactionIndex, ok := params[1].(string)
	if !ok {
		return newErrInvalidParameterType(2, params[1], "")
	}
	r.TransactionIndex = transactionIndex

	return nil
}

// ========== GetTransactionReceipt ============= //

type (
	// Presents transaction hash of a contract
	GetTransactionReceiptRequest  string
	GetTransactionReceiptResponse struct {
		TransactionHash  string `json:"transactionHash"`  // DATA, 32 Bytes - hash of the transaction.
		TransactionIndex string `json:"transactionIndex"` // QUANTITY - integer of the transactions index position in the block.
		BlockHash        string `json:"blockHash"`        // DATA, 32 Bytes - hash of the block where this transaction was in.
		BlockNumber      string `json:"blockNumber"`      // QUANTITY - block number where this transaction was in.
		From             string `json:"from,omitempty"`   // DATA, 20 Bytes - address of the sender.
		// NOTE: must be null if it's a contract creation transaction
		To                string `json:"to,omitempty"` // DATA, 20 Bytes - address of the receiver. null when its a contract creation transaction.
		EffectiveGasPrice string `json:"effectiveGasPrice"`
		CumulativeGasUsed string `json:"cumulativeGasUsed"` // QUANTITY - The total amount of gas used when this transaction was executed in the block.
		GasUsed           string `json:"gasUsed"`           // QUANTITY - The amount of gas used by this specific transaction alone.
		// NOTE: must be null if it's NOT a contract creation transaction
		ContractAddress string `json:"contractAddress,omitempty"` // DATA, 20 Bytes - The contract address created, if the transaction was a contract creation, otherwise null.
		Logs            []Log  `json:"logs"`                      // Array - Array of log objects, which this transaction generated.
		LogsBloom       string `json:"logsBloom"`                 // DATA, 256 Bytes - Bloom filter for light clients to quickly retrieve related logs.
		Status          string `json:"status"`                    // QUANTITY either 1 (success) or 0 (failure)

		// TODO: researching
		// ? Do we need this value
		// Root              string `json:"root,omitempty"`
	}

	Log struct {
		Removed          string   `json:"removed,omitempty"` // TAG - true when the log was removed, due to a chain reorganization. false if its a valid log.
		LogIndex         string   `json:"logIndex"`          // QUANTITY - integer of the log index position in the block. null when its pending log.
		TransactionIndex string   `json:"transactionIndex"`  // QUANTITY - integer of the transactions index position log was created from. null when its pending log.
		TransactionHash  string   `json:"transactionHash"`   // DATA, 32 Bytes - hash of the transactions this log was created from. null when its pending log.
		BlockHash        string   `json:"blockHash"`         // DATA, 32 Bytes - hash of the block where this log was in. null when its pending. null when its pending log.
		BlockNumber      string   `json:"blockNumber"`       // QUANTITY - the block number where this log was in. null when its pending. null when its pending log.
		Address          string   `json:"address"`           // DATA, 20 Bytes - address from which this log originated.
		Data             string   `json:"data"`              // DATA - contains one or more 32 Bytes non-indexed arguments of the log.
		Topics           []string `json:"topics"`            // Array of DATA - Array of 0 to 4 32 Bytes DATA of indexed log arguments.
		Type             string   `json:"type,omitempty"`
	}
)

func (r *GetTransactionReceiptRequest) UnmarshalJSON(data []byte) error {
	var params []string
	err := json.Unmarshal(data, &params)
	if err != nil {
		return errors.Wrap(err, "json unmarshalling")
	}

	if len(params) == 0 {
		return errors.New("params must be set")
	}

	*r = GetTransactionReceiptRequest(params[0])
	return nil
}

// ========== eth_accounts ============= //
type AccountsResponse []string

// ========== eth_getCode ============= //
type (
	GetCodeRequest struct {
		Address     string
		BlockNumber string
	}
	// the code from the given address.
	GetCodeResponse string
)

func (r *GetCodeRequest) UnmarshalJSON(data []byte) error {
	var params []string
	err := json.Unmarshal(data, &params)
	if err != nil {
		return errors.Wrap(err, "json unmarshalling")
	}

	if len(params) == 0 {
		return errors.New("params must be set")
	}

	r.Address = params[0]
	if len(params) > 1 {
		r.BlockNumber = params[1]
	}

	return nil
}

// ========== eth_newBlockFilter ============= //
// a filter id
type NewBlockFilterResponse string

// ========== eth_uninstallFilter ============= //
// the filter id
type UninstallFilterRequest string

// true if the filter was successfully uninstalled, otherwise false.
type UninstallFilterResponse bool

func (r *UninstallFilterRequest) UnmarshalJSON(data []byte) error {
	var params []string
	err := json.Unmarshal(data, &params)
	if err != nil {
		return errors.Wrap(err, "json unmarshalling")
	}

	if len(params) == 0 {
		return errors.New("params must be set")
	}

	*r = UninstallFilterRequest(params[0])

	return nil
}

// ========== eth_getFilterChanges ============= //
// the filter id
type GetFilterChangesRequest string

type GetFilterChangesResponse []interface{}

func (r *GetFilterChangesRequest) UnmarshalJSON(data []byte) error {
	var params []string
	err := json.Unmarshal(data, &params)
	if err != nil {
		return errors.Wrap(err, "json unmarshalling")
	}

	if len(params) == 0 {
		return errors.New("params must be set")
	}

	*r = GetFilterChangesRequest(params[0])

	return nil
}

// ========== eth_estimateGas ============= //

type EstimateGasResponse string

// ========== eth_gasPrice ============= //

type GasPriceResponse *ETHInt

// ========== eth_getBlockByNumber ============= //

type (
	GetBlockByNumberRequest struct {
		BlockNumber     json.RawMessage
		FullTransaction bool
	}

	/*
	 {
	    "number": "0x1b4",
	    "hash": "0xe670ec64341771606e55d6b4ca35a1a6b75ee3d5145a99d05921026d1527331",
	    "parentHash": "0x9646252be9520f6e71339a8df9c55e4d7619deeb018d2a3f2d21fc165dde5eb5",
	    "nonce": "0xe04d296d2460cfb8472af2c5fd05b5a214109c25688d3704aed5484f9a7792f2",
	    "sha3Uncles": "0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347",
	    "logsBloom": "0xe670ec64341771606e55d6b4ca35a1a6b75ee3d5145a99d05921026d1527331",
	    "transactionsRoot": "0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421",
	    "stateRoot": "0xd5855eb08b3387c0af375e9cdb6acfc05eb8f519e419b874b6ff2ffda7ed1dff",
	    "miner": "0x4e65fda2159562a496f9f3522f89122a3088497a",
	    "difficulty": "0x027f07",
	    "totalDifficulty":  "0x027f07",
	    "extraData": "0x0000000000000000000000000000000000000000000000000000000000000000",
	    "size":  "0x027f07",
	    "gasLimit": "0x9f759",
	    "gasUsed": "0x9f759",
	    "timestamp": "0x54e34e8e",
	    "transactions": [{}],
	    "uncles": ["0x1606e5...", "0xd5145a9..."]
	  }
	*/
	GetBlockByNumberResponse = GetBlockByHashResponse
)

// ========== eth_getBlockByHash ============= //

type (
	GetBlockByHashRequest struct {
		BlockHash       string
		FullTransaction bool
	}

	/*
	 {
	    "number": "0x1b4",
	    "hash": "0xe670ec64341771606e55d6b4ca35a1a6b75ee3d5145a99d05921026d1527331",
	    "parentHash": "0x9646252be9520f6e71339a8df9c55e4d7619deeb018d2a3f2d21fc165dde5eb5",
	    "nonce": "0xe04d296d2460cfb8472af2c5fd05b5a214109c25688d3704aed5484f9a7792f2",
	    "sha3Uncles": "0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347",
	    "logsBloom": "0xe670ec64341771606e55d6b4ca35a1a6b75ee3d5145a99d05921026d1527331",
	    "transactionsRoot": "0x56e81f171bcc55a6ff8345e692c0f86e5b48e01b996cadc001622fb5e363b421",
	    "stateRoot": "0xd5855eb08b3387c0af375e9cdb6acfc05eb8f519e419b874b6ff2ffda7ed1dff",
	    "miner": "0x4e65fda2159562a496f9f3522f89122a3088497a",
	    "difficulty": "0x027f07",
	    "totalDifficulty":  "0x027f07",
	    "extraData": "0x0000000000000000000000000000000000000000000000000000000000000000",
	    "size":  "0x027f07",
	    "gasLimit": "0x9f759",
	    "gasUsed": "0x9f759",
	    "timestamp": "0x54e34e8e",
	    "transactions": [{}],
	    "uncles": ["0x1606e5...", "0xd5145a9..."]
	  }
	*/
	GetBlockByHashResponse struct {
		Number     string `json:"number"`
		Hash       string `json:"hash"`
		ParentHash string `json:"parentHash"`
		Nonce      string `json:"nonce"`
		Size       string `json:"size"`
		Miner      string `json:"miner"`
		LogsBloom  string `json:"logsBloom"`
		Timestamp  string `json:"timestamp"`
		ExtraData  string `json:"extraData"`
		//Different type of response []string, []GetTransactionByHashResponse
		Transactions     []interface{} `json:"transactions"`
		StateRoot        string        `json:"stateRoot"`
		TransactionsRoot string        `json:"transactionsRoot"`
		ReceiptsRoot     string        `json:"receiptsRoot"`
		Difficulty       string        `json:"difficulty"`
		// Represents a sum of all blocks difficulties until current block includingly
		TotalDifficulty string `json:"totalDifficulty"`
		GasLimit        string `json:"gasLimit"`
		GasUsed         string `json:"gasUsed"`
		// Represents sha3 hash value based on uncles slice
		Sha3Uncles string   `json:"sha3Uncles"`
		Uncles     []string `json:"uncles"`
	}
)

func (r *GetBlockByNumberRequest) UnmarshalJSON(data []byte) error {
	var params []interface{}
	if err := json.Unmarshal(data, &params); err != nil {
		return errors.Wrap(err, "couldn't unmarhsal data")
	}
	if paramsNum := len(params); paramsNum < 2 {
		return errors.Errorf("invalid parameters number - %d/2", paramsNum)
	}

	blockNumber, ok := params[0].(string)
	if !ok {
		return newErrInvalidParameterType(1, params[0], "")
	}
	// TODO: think of changing []byte type to string type
	r.BlockNumber = json.RawMessage(fmt.Sprintf("\"%s\"", blockNumber))

	fullTxWanted, ok := params[1].(bool)
	if !ok {
		return newErrInvalidParameterType(2, params[1], false)
	}
	r.FullTransaction = fullTxWanted

	return nil
}

func (r *GetBlockByHashRequest) UnmarshalJSON(data []byte) error {
	var params []interface{}
	if err := json.Unmarshal(data, &params); err != nil {
		return errors.Wrap(err, "couldn't unmarhsal parameters")
	}
	if paramsNum := len(params); paramsNum < 2 {
		return errors.Errorf("invalid parameters number - %d/2", paramsNum)
	}

	blockHash, ok := params[0].(string)
	if !ok {
		return newErrInvalidParameterType(1, params[0], "")
	}
	r.BlockHash = blockHash

	fullTxWanted, ok := params[1].(bool)
	if !ok {
		return newErrInvalidParameterType(2, params[1], false)
	}
	r.FullTransaction = fullTxWanted

	return nil
}

// TODO: think of moving it into a separate file
func newErrInvalidParameterType(idx int, gotType interface{}, wantedType interface{}) error {
	return errors.Errorf("invalid %d parameter of %T type, but %T type is expected", idx, gotType, wantedType)
}

// ========== eth_subscribe ============= //

type (
	EthLogSubscriptionParameter struct {
		Address interface{}   `json:"address"`
		Topics  []interface{} `json:"topics"`
	}

	EthSubscriptionRequest struct {
		Method string
		Params *EthLogSubscriptionParameter
	}

	EthSubscriptionResponse string

	/*
	   {
	     "jsonrpc": "2.0",
	     "method": "eth_subscription",
	     "params": {
	       "result": {
	         "difficulty": "0x15d9223a23aa",
	         "extraData": "0xd983010305844765746887676f312e342e328777696e646f7773",
	         "gasLimit": "0x47e7c4",
	         "gasUsed": "0x38658",
	         "logsBloom": "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
	         "miner": "0xf8b483dba2c3b7176a3da549ad41a48bb3121069",
	         "nonce": "0x084149998194cc5f",
	         "number": "0x1348c9",
	         "parentHash": "0x7736fab79e05dc611604d22470dadad26f56fe494421b5b333de816ce1f25701",
	         "receiptRoot": "0x2fab35823ad00c7bb388595cb46652fe7886e00660a01e867824d3dceb1c8d36",
	         "sha3Uncles": "0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347",
	         "stateRoot": "0xb3346685172db67de536d8765c43c31009d0eb3bd9c501c9be3229203f15f378",
	         "timestamp": "0x56ffeff8",
	         "transactionsRoot": "0x0167ffa60e3ebc0b080cdb95f7c0087dd6c0e61413140e39d94d3468d7c9689f"
	       },
	       "subscription": "0x9ce59a13059e417087c02d3236a0b1cc"
	     }
	   }
	*/

	EthSubscriptionNewHeadResponse struct {
		Difficulty       string `json:"difficulty"`
		ExtraData        string `json:"extraData"`
		GasLimit         string `json:"gasLimit"`
		GasUsed          string `json:"gasUsed"`
		LogsBloom        string `json:"logsBloom"`
		Miner            string `json:"miner"`
		Nonce            string `json:"nonce"`
		Number           string `json:"number"`
		ParentHash       string `json:"parentHash"`
		ReceiptRoot      string `json:"receiptRoot"`
		Sha3Uncles       string `json:"sha3Uncles"`
		StateRoot        string `json:"stateRoot"`
		Timestamp        string `json:"timestamp"`
		TransactionsRoot string `json:"transactionsRoot"`
	}
)

var ErrInvalidAddresses = errors.New("Invalid addresses")

func (s *EthLogSubscriptionParameter) GetAddresses() ([]ETHAddress, error) {
	// can be a string or a string array
	if s.Address == nil {
		return []ETHAddress{}, nil
	} else if address, ok := s.Address.(string); ok {
		ethAddress, err := NewETHAddress(address)
		return []ETHAddress{ethAddress}, err
	} else if addresss, ok := s.Address.([]string); ok {
		ethAddresses := []ETHAddress{}
		for _, address := range addresss {
			ethAddress, err := NewETHAddress(address)
			if err != nil {
				return []ETHAddress{}, err
			}
			ethAddresses = append(ethAddresses, ethAddress)
		}
		return ethAddresses, nil
	} else if addresss, ok := s.Address.([]interface{}); ok {
		ethAddresses := []ETHAddress{}
		for _, address := range addresss {
			if addressString, ok := address.(string); ok {
				ethAddress, err := NewETHAddress(addressString)
				if err != nil {
					return []ETHAddress{}, err
				}
				ethAddresses = append(ethAddresses, ethAddress)
			} else {
				return []ETHAddress{}, ErrInvalidAddresses
			}
		}
		return ethAddresses, nil
	}

	return []ETHAddress{}, ErrInvalidAddresses
}

func (r *EthSubscriptionRequest) UnmarshalJSON(data []byte) error {
	var params []interface{}
	if err := json.Unmarshal(data, &params); err != nil {
		return errors.Wrap(err, "couldn't unmarhsal data")
	}

	method, ok := params[0].(string)
	if !ok {
		return newErrInvalidParameterType(1, params[0], "")
	}
	r.Method = method

	if len(params) >= 2 {
		param, err := json.Marshal(params[1])
		if err != nil {
			return err
		}
		var subscriptionParameter EthLogSubscriptionParameter
		err = json.Unmarshal(param, &subscriptionParameter)
		if err != nil {
			return err
		}
		r.Params = &subscriptionParameter
	}

	return nil
}

func (r EthSubscriptionRequest) MarshalJSON() ([]byte, error) {
	output := []interface{}{}
	output = append(output, r.Method)
	if r.Params != nil {
		output = append(output, r.Params)
	}

	return json.Marshal(output)
}

func NewEthSubscriptionNewHeadResponse(block *GetBlockByHashResponse) *EthSubscriptionNewHeadResponse {
	return &EthSubscriptionNewHeadResponse{
		Difficulty:       block.Difficulty,
		ExtraData:        block.ExtraData,
		GasLimit:         block.GasLimit,
		GasUsed:          block.GasUsed,
		LogsBloom:        block.LogsBloom,
		Miner:            block.Miner,
		Nonce:            block.Nonce,
		Number:           block.Number,
		ParentHash:       block.ParentHash,
		ReceiptRoot:      block.ReceiptsRoot,
		Sha3Uncles:       block.Sha3Uncles,
		Timestamp:        block.Timestamp,
		TransactionsRoot: block.TransactionsRoot,
	}
}

// ========== eth_unsubscribe =========== //

type (
	EthUnsubscribeRequest []string

	EthUnsubscribeResponse bool
)

// ========== eth_newFilter ============= //

type NewFilterRequest struct {
	FromBlock json.RawMessage `json:"fromBlock"`
	ToBlock   json.RawMessage `json:"toBlock"`
	Address   json.RawMessage `json:"address"`
	Topics    []interface{}   `json:"topics"`
}

func (r *NewFilterRequest) UnmarshalJSON(data []byte) error {
	var params []json.RawMessage
	if err := json.Unmarshal(data, &params); err != nil {
		return errors.Wrap(err, "json unmarshalling")
	}

	if len(params) == 0 {
		return errors.New("params must be set")
	}
	type Req NewFilterRequest
	var req Req

	if err := json.Unmarshal(params[0], &req); err != nil {
		return errors.Wrap(err, "json unmarshalling")
	}

	*r = NewFilterRequest(req)

	return nil
}

type NewFilterResponse string

// ========== eth_getBalance ============= //

type GetBalanceRequest struct {
	Address string
	Block   json.RawMessage
}

func (r *GetBalanceRequest) UnmarshalJSON(data []byte) error {
	tmp := []interface{}{&r.Address, &r.Block}

	return json.Unmarshal(data, &tmp)
}

type GetBalanceResponse string

// =======GetTransactionCount ============= //
type (
	GetTransactionCountRequest struct {
		Address string
		Tag     string
	}
)

// ========== getstorage ============= //
type (
	GetStorageRequest struct {
		Address     string
		Index       string
		BlockNumber string
	}
	GetStorageResponse string
)

func (r *GetStorageRequest) UnmarshalJSON(data []byte) error {
	tmp := []interface{}{&r.Address, &r.Index, &r.BlockNumber}
	return json.Unmarshal(data, &tmp)
}

// ======= eth_chainId ============= //
type ChainIdResponse string

// ======= eth_subscription ======== //
type EthSubscription struct {
	SubscriptionID string      `json:"subscription"`
	Result         interface{} `json:"result"`
}

// ======= qtum_getUTXOs ============= //

type (
	GetUTXOsRequest struct {
		Address      string
		MinSumAmount decimal.Decimal
	}

	QtumUTXO struct {
		Address string `json:"address"`
		TXID    string `json:"txid"`
		Vout    uint   `json:"vout"`
		Amount  string `json:"amount"`
	}

	GetUTXOsResponse []QtumUTXO
)

func (req *GetUTXOsRequest) UnmarshalJSON(params []byte) error {
	paramsBytesNum := len(params)
	if paramsBytesNum < 2 {
		return fmt.Errorf("bytes number < 2")
	}

	params = params[1 : paramsBytesNum-1] // drop `[`, `]`

	for i, vByte := range params {
		if vByte == ',' {
			req.Address = string(bytes.Trim((params[:i]), " \n\t\""))

			if paramsBytesNum < i+1 {
				// `,` is the last byte, that is
				// there are no bytes left
				return nil
			}

			var (
				minAmount = string(bytes.Trim(params[i+1:], " \n\t\""))
				err       error
			)
			req.MinSumAmount, err = decimal.NewFromString(minAmount)
			if err != nil {
				return fmt.Errorf("couldn't convert minimum amount from string: %s", err)
			}

			return nil
		}
	}

	return fmt.Errorf("an array of length 2 - address and minimum amount - is expected")
}

func (req GetUTXOsRequest) CheckHasValidValues() error {
	if !common.IsHexAddress(req.Address) {
		return errors.Errorf("invalid Ethereum address - %q", req.Address)
	}
	if req.MinSumAmount.LessThanOrEqual(decimal.NewFromInt(0)) {
		return errors.Errorf("invalid minimum amount - %q (<= 0)", req.MinSumAmount)
	}
	return nil
}

// ======= web3_sha3 ======= //
type Web3Sha3Request struct {
	Message string
}

func (r *Web3Sha3Request) UnmarshalJSON(data []byte) error {
	var params []interface{}
	if err := json.Unmarshal(data, &params); err != nil {
		return errors.Wrap(err, "couldn't unmarhsal parameters")
	}
	paramsNum := len(params)
	if paramsNum == 0 {
		return errors.Errorf("missing value for required argument 0")
	} else if paramsNum > 1 {
		return errors.Errorf("too many arguments, want at most 1")
	}

	message, ok := params[0].(string)
	if !ok {
		return newErrInvalidParameterType(1, params[0], "")
	}
	r.Message = message

	return nil
}

type NetPeerCountResponse string
