package qtum

import (
	"encoding/json"
	"errors"
	"fmt"
)

const (
	RPCVersion = "1.0"
)

const (
	MethodGetHexAddress         = "gethexaddress"
	MethodFromHexAddress        = "fromhexaddress"
	MethodSendToContract        = "sendtocontract"
	MethodGetTransactionReceipt = "gettransactionreceipt"
	MethodGetTransaction        = "gettransaction"
	MethodGetRawTransaction     = "getrawtransaction"
	MethodCreateContract        = "createcontract"
	MethodSendToAddress         = "sendtoaddress"
	MethodCallContract          = "callcontract"
	MethodDecodeRawTransaction  = "decoderawtransaction"
	MethodGetTransactionOut     = "gettxout"
	MethodGetBlockCount         = "getblockcount"
	MethodGetBlockChainInfo     = "getblockchaininfo"
	MethodSearchLogs            = "searchlogs"
	MethodWaitForLogs           = "waitforlogs"
	MethodGetBlockHash          = "getblockhash"
	MethodGetBlockHeader        = "getblockheader"
	MethodGetBlock              = "getblock"
	MethodGetAddressesByAccount = "getaddressesbyaccount"
	MethodGetAccountInfo        = "getaccountinfo"
	MethodGenerateToAddress     = "generatetoaddress"
	MethodListUnspent           = "listunspent"
	MethodCreateRawTx           = "createrawtransaction"
	MethodSignRawTx             = "signrawtransactionwithwallet"
	MethodSendRawTx             = "sendrawtransaction"
)

type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	ID      json.RawMessage `json:"id"`
	Params  json.RawMessage `json:"params"`
}

type SuccessJSONRPCResult struct {
	JSONRPC   string          `json:"jsonrpc"`
	RawResult json.RawMessage `json:"result"`
	ID        json.RawMessage `json:"id"`
}

type JSONRPCResult struct {
	JSONRPC   string          `json:"jsonrpc"`
	RawResult json.RawMessage `json:"result,omitempty"`
	Error     *JSONRPCError   `json:"error,omitempty"`
	ID        json.RawMessage `json:"id"`
}

type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (err *JSONRPCError) Error() string {
	return fmt.Sprintf("qtum [code: %d] %s", err.Code, err.Message)
}

// Tries to associate returned error with one of already known (implemented) errors,
// in which we may be interesting. If returned error is unknown, returns original
// error value
func (err *JSONRPCError) TryGetKnownError() error {
	switch err.Code {
	case -5:
		// - address doesn't exist
		// - invalid address
		return ErrInvalidAddress
	default:
		return err
	}
}

var (
	ErrInvalidAddress = errors.New("invalid address")
	// TODO: add
	// - insufficient balance
	// - amount out of range
)
