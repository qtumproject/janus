package eth

import (
	"encoding/json"
	"fmt"
)

var EmptyLogsBloom = "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000"
var DefaultSha3Uncles = "0x1dcc4de8dec75d7aab85b567b6ccd41ad312451b948a7413f0a142fd40d49347"

const (
	RPCVersion = "2.0"
)

type JSONRPCRequest struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method"`
	ID      json.RawMessage `json:"id"`
	Params  json.RawMessage `json:"params"`
}

type JSONRPCResult struct {
	JSONRPC   string          `json:"jsonrpc"`
	RawResult json.RawMessage `json:"result,omitempty"`
	Error     *JSONRPCError   `json:"error,omitempty"`
	ID        json.RawMessage `json:"id"`
}

// JSONRPCError contains the message and code for an ETH RPC error
type JSONRPCError struct {
	Code    int    `json:"code"`
	Message string `json:"message"`
}

func (err *JSONRPCError) Error() string {
	return fmt.Sprintf("eth [code: %d] %s", err.Code, err.Message)
}

func NewJSONRPCResult(id json.RawMessage, res interface{}) (*JSONRPCResult, error) {
	rawResult, err := json.Marshal(res)
	if err != nil {
		return nil, err
	}

	return &JSONRPCResult{
		JSONRPC:   RPCVersion,
		ID:        id,
		RawResult: rawResult,
	}, nil
}

type JSONRPCNotification struct {
	JSONRPC string          `json:"jsonrpc"`
	Method  string          `json:"method,omitempty"`
	Params  json.RawMessage `json:"params"`
}

func NewJSONRPCNotification(method string, params interface{}) (*JSONRPCNotification, error) {
	rawParams, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	return &JSONRPCNotification{
		JSONRPC: RPCVersion,
		Method:  method,
		Params:  rawParams,
	}, nil
}
