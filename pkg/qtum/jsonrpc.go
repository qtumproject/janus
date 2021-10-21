package qtum

import (
	"encoding/json"
	"errors"
	"fmt"

	"github.com/qtumproject/janus/pkg/eth"
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
	MethodGetPeerInfo           = "getpeerinfo"
	MethodGetNetworkInfo        = "getnetworkinfo"
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
	MethodGetStorage            = "getstorage"
	MethodCreateRawTx           = "createrawtransaction"
	MethodSignRawTx             = "signrawtransactionwithwallet"
	MethodSendRawTx             = "sendrawtransaction"
	MethodGetStakingInfo        = "getstakinginfo"
	MethodGetAddressBalance     = "getaddressbalance"
	MethodGetAddressUTXOs       = "getaddressutxos"
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
	knownError := errorCodeMap[err.Code]
	if knownError == nil {
		return err
	}
	return knownError
}

func IsKnownError(err error) bool {
	_, contains := errorToCodeMap[err]
	return contains
}

func GetErrorCode(err error) int {
	errorCode, contains := errorToCodeMap[err]
	if !contains {
		return 0
	}
	return errorCode
}

func GetErrorResponse(err error) *eth.JSONRPCError {
	errorCode := GetErrorCode(err)
	if errorCode == 0 {
		return nil
	}

	return &eth.JSONRPCError{
		Code:    errorCode,
		Message: err.Error(),
	}
}

var (
	errorCodeMap   = map[int]error{}
	errorToCodeMap = map[error]int{}
	// taken from https://github.com/qtumproject/qtum/blob/master/src/rpc/protocol.h
	// Standard JSON-RPC 2.0 errors
	ErrInvalidRequest = errors.New("invalid request") // -32600
	// RPC_METHOD_NOT_FOUND is internally mapped to HTTP_NOT_FOUND (404).
	// It should not be used for application-layer errors.
	ErrMethodNotFound = errors.New("method not found")   // -32601
	ErrInvalidParams  = errors.New("invalid parameters") // -32602
	// RPC_INTERNAL_ERROR should only be used for genuine errors in bitcoind
	// (for example datadir corruption).
	ErrInternalError = errors.New("internal error") // -32603
	ErrParseError    = errors.New("parse error")    // -32700
	// general application defined errors
	ErrMiscError = errors.New("misc error") // -1
	ErrTypeError = errors.New("type error") // -3
	// May be caused by:
	// 	- provided address doesn't exist
	// 	- provided address is invalid
	// 	- data is not acquirable via used RPC method and provided address
	ErrInvalidAddress       = errors.New("invalid address")         // -5
	ErrOutOfMemory          = errors.New("oom")                     // -7
	ErrInvalidParameter     = errors.New("invalid parameter")       // -8
	ErrDatabaseError        = errors.New("database error")          // -20
	ErrDeserializationError = errors.New("deserialization error")   // -22
	ErrVerifyError          = errors.New("verify error")            // -25
	ErrVerifyRejected       = errors.New("verify rejected")         // -26
	ErrVerifyAlreadyInChain = errors.New("verify already in chain") // -27
	ErrInWarmup             = errors.New("in warmup")               // -28
	ErrMethodDeprecated     = errors.New("method deprecated")       // -29

	// P2P client errors
	ErrClientNotConnected      = errors.New("client not connected")             // -9
	ErrClientInInitialDownload = errors.New("still downloading initial blocks") // -10
	ErrNodeAlreadyAdded        = errors.New("not already added")                // -23
	ErrNodeNotAdded            = errors.New("node not added")                   // -24
	ErrNodeNotConnected        = errors.New("not not connected")                // -29
	ErrInvalidIpOrSubnet       = errors.New("invalid ip or subnet")             // -30
	ErrP2PDisabled             = errors.New("p2p disabled")                     // -31

	// chain errors
	ErrMempoolDisabled = errors.New("mempool disabled") // -33

	// wallet errors
	ErrWalletError                = errors.New("wallet error")                  // -4
	ErrWalletInsufficientFunds    = errors.New("wallet insufficient funds")     // -6
	ErrWalletInvalidLabelName     = errors.New("wallet invalid label name")     // -11
	ErrWalletKeypoolRanOut        = errors.New("wallet keypool ran out")        // -12
	ErrWalletUnlockNeeded         = errors.New("wallet unlock needed")          // -13
	ErrWalletPassphraseIncorrect  = errors.New("wallet passphrase incorrect")   // -14
	ErrWalletWrongEncryptionState = errors.New("wallet wrong encryption state") // -15
	ErrWalletEncryptionFailed     = errors.New("failed to encrypt the wallet")  // -16
	ErrWalletAlreadyUnlocked      = errors.New("wallet already unlocked")       // -17
	ErrWalletNotFound             = errors.New("wallet not found")              // -18
	ErrWalletNotSpecified         = errors.New("wallet not specified")          // -19

	ErrForbiddenBySafeMode = errors.New("server is in safe mode, and command is not allowed in safe mode") // -2

	// Http server work queue is full, returned as a raw string, not inside a JSON response
	ErrQtumWorkQueueDepth = errors.New("Work queue depth exceeded")
	// TODO: add
	// - insufficient balance
	// - amount out of range
)

func init() {
	errorCodeMap[-1] = ErrMiscError
	errorCodeMap[-2] = ErrForbiddenBySafeMode
	errorCodeMap[-3] = ErrTypeError
	errorCodeMap[-4] = ErrWalletError
	errorCodeMap[-5] = ErrInvalidAddress
	errorCodeMap[-6] = ErrWalletInsufficientFunds
	errorCodeMap[-7] = ErrOutOfMemory
	errorCodeMap[-8] = ErrInvalidParameter
	errorCodeMap[-9] = ErrClientNotConnected
	errorCodeMap[-10] = ErrClientInInitialDownload
	errorCodeMap[-11] = ErrWalletInvalidLabelName
	errorCodeMap[-12] = ErrWalletKeypoolRanOut
	errorCodeMap[-13] = ErrWalletUnlockNeeded
	errorCodeMap[-14] = ErrWalletPassphraseIncorrect
	errorCodeMap[-15] = ErrWalletWrongEncryptionState
	errorCodeMap[-16] = ErrWalletEncryptionFailed
	errorCodeMap[-17] = ErrWalletAlreadyUnlocked
	errorCodeMap[-18] = ErrWalletNotFound
	errorCodeMap[-19] = ErrWalletNotSpecified
	errorCodeMap[-20] = ErrDatabaseError
	// -21 unused
	errorCodeMap[-22] = ErrDeserializationError
	errorCodeMap[-23] = ErrNodeAlreadyAdded
	errorCodeMap[-24] = ErrNodeNotAdded
	errorCodeMap[-25] = ErrVerifyError
	errorCodeMap[-26] = ErrVerifyRejected
	errorCodeMap[-27] = ErrVerifyAlreadyInChain
	errorCodeMap[-28] = ErrInWarmup
	errorCodeMap[-29] = ErrMethodDeprecated
	errorCodeMap[-30] = ErrInvalidIpOrSubnet
	errorCodeMap[-31] = ErrP2PDisabled

	errorCodeMap[-33] = ErrMempoolDisabled

	errorCodeMap[-32600] = ErrInvalidRequest
	errorCodeMap[-32601] = ErrMethodNotFound
	errorCodeMap[-32602] = ErrInvalidParams
	errorCodeMap[-32603] = ErrInternalError
	errorCodeMap[-32700] = ErrParseError

	for k, v := range errorCodeMap {
		errorToCodeMap[v] = k
	}
}
