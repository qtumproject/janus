package internal

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/dcb9/go-ethereum/common/hexutil"
	kitLog "github.com/go-kit/kit/log"
	"github.com/go-kit/kit/log/level"
	"github.com/qtumproject/janus/pkg/eth"
	"github.com/qtumproject/janus/pkg/qtum"
	"github.com/qtumproject/janus/pkg/utils"
	"github.com/shopspring/decimal"
)

//copy of qtum.Doer interface
type Doer interface {
	Do(*http.Request) (*http.Response, error)
	AddRawResponse(requestType string, rawResponse []byte)
	AddResponse(requestType string, responseResult interface{}) error
	AddResponseWithRequestID(requestID int, requestType string, responseResult interface{}) error
	AddError(requestType string, responseError *eth.JSONRPCError) error
	AddErrorWithRequestID(requestID int, requestType string, responseError *eth.JSONRPCError) error
}

func NewDoerMappedMock() *doerMappedMock {
	return &doerMappedMock{
		Responses: make(map[string][][]byte),
	}
}

//type for mocking requests to client with request -> response mapping
type doerMappedMock struct {
	mutex     sync.Mutex
	latestId  int
	Responses map[string][][]byte
}

func (d *doerMappedMock) updateId(id int) {
	if id > d.latestId {
		d.latestId = id
	}
}

func (d *doerMappedMock) Do(request *http.Request) (*http.Response, error) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	requestJSON, err := parseRequestFromBody(request)
	if err != nil {
		return nil, err
	}

	if d.Responses[requestJSON.Method] == nil {
		log.Printf("No mocked response for %s\n", requestJSON.Method)
	}

	responseWriter := ioutil.NopCloser(bytes.NewReader(d.popResponse(requestJSON.Method)))
	return &http.Response{
		StatusCode: 200,
		Body:       responseWriter,
	}, nil
}

func PrepareEthRPCRequest(id int, params []json.RawMessage) (*eth.JSONRPCRequest, error) {
	requestID, err := json.Marshal(1)
	if err != nil {
		return nil, err
	}

	paramsArrayRaw, err := json.Marshal(params)
	if err != nil {
		return nil, err
	}

	requestRPC := eth.JSONRPCRequest{
		JSONRPC: "2.0",
		Method:  "eth_protocolVersion",
		ID:      requestID,
		Params:  paramsArrayRaw,
	}

	return &requestRPC, nil
}

func prepareRawResponse(requestID int, responseResult interface{}, responseError *eth.JSONRPCError) ([]byte, error) {
	requestIDRaw, err := json.Marshal(requestID)
	if err != nil {
		return nil, err
	}

	var responseResultRaw json.RawMessage
	if responseResult != nil {
		var alreadyByteArray bool
		responseResultRaw, alreadyByteArray = responseResult.([]byte)
		if !alreadyByteArray {
			responseResultRaw, err = json.Marshal(responseResult)
			if err != nil {
				return nil, err
			}
		}
	}

	responseRPC := &eth.JSONRPCResult{
		JSONRPC:   "2.0",
		RawResult: responseResultRaw,
		Error:     responseError,
		ID:        requestIDRaw,
	}

	responseRPCRaw, err := json.Marshal(responseRPC)

	return responseRPCRaw, err
}

func (d *doerMappedMock) pushResponse(requestType string, responseRaw []byte) {
	if _, exists := d.Responses[requestType]; !exists {
		d.Responses[requestType] = make([][]byte, 0, 1)
	}
	d.Responses[requestType] = append(d.Responses[requestType], responseRaw)
}

func (d *doerMappedMock) popResponse(requestType string) []byte {
	responses := len(d.Responses[requestType])
	if responses == 0 {
		return nil
	} else {
		latest := d.Responses[requestType][0]
		if responses > 1 {
			fmt.Printf("popped response: %s\n", requestType)
			d.Responses[requestType] = d.Responses[requestType][1:responses]
		} else {
			fmt.Printf("one response: %s\n", requestType)
		}
		return latest
	}
}

func (d *doerMappedMock) AddRawResponse(requestType string, rawResponse []byte) {
	d.mutex.Lock()
	d.pushResponse(requestType, rawResponse)
	d.mutex.Unlock()
}

func (d *doerMappedMock) AddResponse(requestType string, responseResult interface{}) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	requestID := d.latestId + 1
	responseRaw, err := prepareRawResponse(requestID, responseResult, nil)
	if err != nil {
		return err
	}

	d.updateId(requestID)
	d.pushResponse(requestType, responseRaw)
	return nil
}

func (d *doerMappedMock) AddResponseWithRequestID(requestID int, requestType string, responseResult interface{}) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	responseRaw, err := prepareRawResponse(requestID, responseResult, nil)
	if err != nil {
		return err
	}

	d.updateId(requestID)
	d.pushResponse(requestType, responseRaw)
	return nil
}

func (d *doerMappedMock) AddError(requestType string, responseError *eth.JSONRPCError) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	requestID := d.latestId + 1
	responseRaw, err := prepareRawResponse(requestID, nil, responseError)
	if err != nil {
		return err
	}

	d.updateId(requestID)
	d.pushResponse(requestType, responseRaw)
	return nil
}

func (d *doerMappedMock) AddErrorWithRequestID(requestID int, requestType string, responseError *eth.JSONRPCError) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	responseRaw, err := prepareRawResponse(requestID, nil, responseError)
	if err != nil {
		return err
	}

	d.updateId(requestID)
	d.pushResponse(requestType, responseRaw)
	return nil
}

func parseRequestFromBody(request *http.Request) (*eth.JSONRPCRequest, error) {
	requestJSON := eth.JSONRPCRequest{}
	requestBody, err := ioutil.ReadAll(request.Body)
	if err != nil {
		return nil, err
	}

	err = json.Unmarshal(requestBody, &requestJSON)
	if err != nil {
		return nil, err
	}

	return &requestJSON, err
}

func CreateMockedClient(doerInstance Doer) (qtumClient *qtum.Qtum, err error) {
	logger := kitLog.NewLogfmtLogger(os.Stdout)
	if !isDebugEnvironmentVariableSet() {
		logger = level.NewFilter(logger, level.AllowWarn())
	}
	qtumJSONRPC, err := qtum.NewClient(
		true,
		"http://user:pass@mocked",
		qtum.SetDoer(doerInstance),
		qtum.SetDebug(isDebugEnvironmentVariableSet()),
		qtum.SetLogger(logger),
	)
	if err != nil {
		return
	}

	qtumClient, err = qtum.New(qtumJSONRPC, "test")
	return
}

func isDebugEnvironmentVariableSet() bool {
	return strings.ToLower(os.Getenv("DEBUG")) == "true"
}

func MustMarshalIndent(v interface{}, prefix, indent string) []byte {
	res, err := json.MarshalIndent(v, prefix, indent)
	if err != nil {
		panic(err)
	}
	return res
}

var (
	GetTransactionByHashBlockNumberHex     = "0xf8f"
	GetTransactionByHashBlockNumberInteger = uint64(3983)
	GetTransactionByHashBlockHash          = "bba11e1bacc69ba535d478cf1f2e542da3735a517b0b8eebaf7e6bb25eeb48c5"
	GetTransactionByHashBlockHexHash       = utils.AddHexPrefix(GetTransactionByHashBlockHash)
	GetTransactionByHashResponseData       = eth.GetTransactionByHashResponse{
		BlockHash:        GetTransactionByHashBlockHexHash,
		BlockNumber:      GetTransactionByHashBlockNumberHex,
		TransactionIndex: "0x2",
		Hash:             "0x11e97fa5877c5df349934bafc02da6218038a427e8ed081f048626fa6eb523f5",
		Nonce:            "0x0",
		Value:            "0x0",
		Input:            "0x020000000159c0514feea50f915854d9ec45bc6458bb14419c78b17e7be3f7fd5f563475b5010000006a473044022072d64a1f4ea2d54b7b05050fc853ab192c91cc5ca17e23007867f92f2ab59d9202202b8c9ab9348c8edbb3b98b1788382c8f37642ec9bd6a4429817ab79927319200012103520b1500a400483f19b93c4cb277a2f29693ea9d6739daaf6ae6e971d29e3140feffffff02000000000000000063010403400d0301644440c10f190000000000000000000000006b22910b1e302cf74803ffd1691c2ecb858d3712000000000000000000000000000000000000000000000000000000000000000a14be528c8378ff082e4ba43cb1baa363dbf3f577bfc260e66272970100001976a9146b22910b1e302cf74803ffd1691c2ecb858d371288acb00f0000",
		From:             "0x0000000000000000000000000000000000000000",
		To:               "0x0000000000000000000000000000000000000000",
		Gas:              "0x0",
		GasPrice:         "0x0",
	}

	GetTransactionByHashResponse = CreateTransactionByHashResponse()

	GetTransactionByHashResponseWithTransactions = eth.GetBlockByHashResponse{
		Number:           GetTransactionByHashBlockNumberHex,
		Hash:             GetTransactionByHashBlockHexHash,
		ParentHash:       "0x6d7d56af09383301e1bb32a97d4a5c0661d62302c06a778487d919b7115543be",
		Miner:            "0x0000000000000000000000000000000000000000",
		Size:             "0x26c",
		Nonce:            "0x0000000000000000",
		TransactionsRoot: "0x0b5f03dc9d456c63c587cc554b70c1232449be43d1df62bc25a493b04de90334",
		ReceiptsRoot:     "0x0b5f03dc9d456c63c587cc554b70c1232449be43d1df62bc25a493b04de90334",
		StateRoot:        "0x3e49216e58f1ad9e6823b5095dc532f0a6cc44943d36ff4a7b1aa474e172d672",
		Difficulty:       "0x4",
		TotalDifficulty:  "0x4",
		LogsBloom:        eth.EmptyLogsBloom,
		ExtraData:        "0x0000000000000000000000000000000000000000000000000000000000000000",
		GasLimit:         utils.AddHexPrefix(qtum.DefaultBlockGasLimit),
		GasUsed:          "0x0",
		Timestamp:        "0x5b95ebd0",
		Transactions: []interface{}{
			GetTransactionByHashResponseData,
			GetTransactionByHashResponseData,
		},
		Sha3Uncles: eth.DefaultSha3Uncles,
		Uncles:     []string{},
	}

	GetTransactionByBlockResponse = eth.GetBlockByNumberResponse{
		Number:           GetTransactionByHashBlockNumberHex,
		Hash:             GetTransactionByHashBlockHexHash,
		ParentHash:       "0x6d7d56af09383301e1bb32a97d4a5c0661d62302c06a778487d919b7115543be",
		Miner:            "0x0000000000000000000000000000000000000000",
		Size:             "0x26c",
		Nonce:            "0x0000000000000000",
		TransactionsRoot: "0x0b5f03dc9d456c63c587cc554b70c1232449be43d1df62bc25a493b04de90334",
		ReceiptsRoot:     "0x0b5f03dc9d456c63c587cc554b70c1232449be43d1df62bc25a493b04de90334",
		StateRoot:        "0x3e49216e58f1ad9e6823b5095dc532f0a6cc44943d36ff4a7b1aa474e172d672",
		Difficulty:       "0x4",
		TotalDifficulty:  "0x4",
		LogsBloom:        eth.EmptyLogsBloom,
		ExtraData:        "0x0000000000000000000000000000000000000000000000000000000000000000",
		GasLimit:         utils.AddHexPrefix(qtum.DefaultBlockGasLimit),
		GasUsed:          "0x0",
		Timestamp:        "0x5b95ebd0",
		Transactions: []interface{}{"0x3208dc44733cbfa11654ad5651305428de473ef1e61a1ec07b0c1a5f4843be91",
			"0x8fcd819194cce6a8454b2bec334d3448df4f097e9cdc36707bfd569900268950"},
		Sha3Uncles: eth.DefaultSha3Uncles,
		Uncles:     []string{},
	}

	GetTransactionByBlockResponseWithTransactions = eth.GetBlockByNumberResponse{
		Number:           GetTransactionByHashBlockNumberHex,
		Hash:             GetTransactionByHashBlockHexHash,
		ParentHash:       "0x6d7d56af09383301e1bb32a97d4a5c0661d62302c06a778487d919b7115543be",
		Miner:            "0x0000000000000000000000000000000000000000",
		Size:             "0x26c",
		Nonce:            "0x0000000000000000",
		TransactionsRoot: "0x0b5f03dc9d456c63c587cc554b70c1232449be43d1df62bc25a493b04de90334",
		ReceiptsRoot:     "0x0b5f03dc9d456c63c587cc554b70c1232449be43d1df62bc25a493b04de90334",
		StateRoot:        "0x3e49216e58f1ad9e6823b5095dc532f0a6cc44943d36ff4a7b1aa474e172d672",
		Difficulty:       "0x4",
		TotalDifficulty:  "0x4",
		LogsBloom:        eth.EmptyLogsBloom,
		ExtraData:        "0x0000000000000000000000000000000000000000000000000000000000000000",
		GasLimit:         utils.AddHexPrefix(qtum.DefaultBlockGasLimit),
		GasUsed:          "0x0",
		Timestamp:        "0x5b95ebd0",
		Transactions: []interface{}{
			GetTransactionByHashResponseData,
			GetTransactionByHashResponseData,
		},
		Sha3Uncles: eth.DefaultSha3Uncles,
		Uncles:     []string{},
	}

	GetBlockResponse = qtum.GetBlockResponse{
		Hash:              GetTransactionByHashBlockHash,
		Confirmations:     1,
		Strippedsize:      584,
		Size:              620,
		Weight:            2372,
		Height:            3983,
		Version:           536870912,
		VersionHex:        "20000000",
		Merkleroot:        "0b5f03dc9d456c63c587cc554b70c1232449be43d1df62bc25a493b04de90334",
		Time:              1536551888,
		Mediantime:        1536551728,
		Nonce:             0,
		Bits:              "207fffff",
		Difficulty:        4.656542373906925,
		Chainwork:         "0000000000000000000000000000000000000000000000000000000000001f20",
		HashStateRoot:     "3e49216e58f1ad9e6823b5095dc532f0a6cc44943d36ff4a7b1aa474e172d672",
		HashUTXORoot:      "130a3e712d9f8b06b83f5ebf02b27542fb682cdff3ce1af1c17b804729d88a47",
		Previousblockhash: "6d7d56af09383301e1bb32a97d4a5c0661d62302c06a778487d919b7115543be",
		Flags:             "proof-of-stake",
		Proofhash:         "15bd6006ecbab06708f705ecf68664b78b388e4d51416cdafb019d5b90239877",
		Modifier:          "a79c00d1d570743ca8135a173d535258026d26bafbc5f3d951c3d33486b1f120",
		Txs: []string{"3208dc44733cbfa11654ad5651305428de473ef1e61a1ec07b0c1a5f4843be91",
			"8fcd819194cce6a8454b2bec334d3448df4f097e9cdc36707bfd569900268950"},
		Nextblockhash: "d7758774cfdd6bab7774aa891ae035f1dc5a2ff44240784b5e7bdfd43a7a6ec1",
		Signature:     "3045022100a6ab6c2b14b1f73e734f1a61d4d22385748e48836492723a6ab37cdf38525aba022014a51ecb9e51f5a7a851641683541fec6f8f20205d0db49e50b2a4e5daed69d2",
	}
)

func CreateTransactionByHashResponse() eth.GetBlockByHashResponse {
	return eth.GetBlockByHashResponse{
		Number:           GetTransactionByHashBlockNumberHex,
		Hash:             GetTransactionByHashBlockHexHash,
		ParentHash:       "0x6d7d56af09383301e1bb32a97d4a5c0661d62302c06a778487d919b7115543be",
		Miner:            "0x0000000000000000000000000000000000000000",
		Size:             "0x26c",
		Nonce:            "0x0000000000000000",
		TransactionsRoot: "0x0b5f03dc9d456c63c587cc554b70c1232449be43d1df62bc25a493b04de90334",
		ReceiptsRoot:     "0x0b5f03dc9d456c63c587cc554b70c1232449be43d1df62bc25a493b04de90334",
		StateRoot:        "0x3e49216e58f1ad9e6823b5095dc532f0a6cc44943d36ff4a7b1aa474e172d672",
		Difficulty:       "0x4",
		TotalDifficulty:  "0x4",
		LogsBloom:        eth.EmptyLogsBloom,
		ExtraData:        "0x0000000000000000000000000000000000000000000000000000000000000000",
		GasLimit:         utils.AddHexPrefix(qtum.DefaultBlockGasLimit),
		GasUsed:          "0x0",
		Timestamp:        "0x5b95ebd0",
		Transactions: []interface{}{"0x3208dc44733cbfa11654ad5651305428de473ef1e61a1ec07b0c1a5f4843be91",
			"0x8fcd819194cce6a8454b2bec334d3448df4f097e9cdc36707bfd569900268950"},
		Sha3Uncles: eth.DefaultSha3Uncles,
		Uncles:     []string{},
	}
}

func QtumTransactionReceipt(logs []qtum.Log) qtum.TransactionReceipt {
	return qtum.TransactionReceipt{
		BlockHash:         GetTransactionByHashBlockHexHash,
		BlockNumber:       GetTransactionByHashBlockNumberInteger,
		TransactionHash:   GetTransactionByHashResponseData.Hash,
		TransactionIndex:  hexutil.MustDecodeUint64(GetTransactionByHashResponseData.TransactionIndex),
		From:              GetTransactionByHashResponseData.From,
		To:                GetTransactionByHashResponseData.To,
		CumulativeGasUsed: hexutil.MustDecodeUint64(GetTransactionByHashResponseData.Gas),
		GasUsed:           hexutil.MustDecodeUint64(GetTransactionByHashResponseData.Gas),
		ContractAddress:   GetTransactionByHashResponseData.To,
		Log:               logs,
	}
}

func QtumWaitForLogsEntry(log qtum.Log) qtum.WaitForLogsEntry {
	return qtum.WaitForLogsEntry{
		BlockHash:         GetTransactionByHashBlockHexHash,
		BlockNumber:       GetTransactionByHashBlockNumberInteger,
		TransactionHash:   GetTransactionByHashResponseData.Hash,
		TransactionIndex:  hexutil.MustDecodeUint64(GetTransactionByHashResponseData.TransactionIndex),
		From:              GetTransactionByHashResponseData.From,
		To:                GetTransactionByHashResponseData.To,
		CumulativeGasUsed: hexutil.MustDecodeUint64(GetTransactionByHashResponseData.Gas),
		GasUsed:           hexutil.MustDecodeUint64(GetTransactionByHashResponseData.Gas),
		ContractAddress:   strings.TrimPrefix(GetTransactionByHashResponseData.To, "0x"),
		Topics:            log.Topics,
		Data:              log.Data,
	}
}

func SetupGetBlockByHashResponses(t *testing.T, mockedClientDoer Doer) {
	//preparing answer to "getblockhash"
	getBlockHashResponse := qtum.GetBlockHashResponse(GetTransactionByHashBlockHexHash)
	err := mockedClientDoer.AddResponse(qtum.MethodGetBlockHash, getBlockHashResponse)
	if err != nil {
		t.Fatal(err)
	}

	getBlockHeaderResponse := qtum.GetBlockHeaderResponse{
		Hash:              GetTransactionByHashBlockHash,
		Confirmations:     1,
		Height:            3983,
		Version:           536870912,
		VersionHex:        "20000000",
		Merkleroot:        "0b5f03dc9d456c63c587cc554b70c1232449be43d1df62bc25a493b04de90334",
		Time:              1536551888,
		Mediantime:        1536551728,
		Nonce:             0,
		Bits:              "207fffff",
		Difficulty:        4.656542373906925,
		Chainwork:         "0000000000000000000000000000000000000000000000000000000000001f20",
		HashStateRoot:     "3e49216e58f1ad9e6823b5095dc532f0a6cc44943d36ff4a7b1aa474e172d672",
		HashUTXORoot:      "130a3e712d9f8b06b83f5ebf02b27542fb682cdff3ce1af1c17b804729d88a47",
		Previousblockhash: "6d7d56af09383301e1bb32a97d4a5c0661d62302c06a778487d919b7115543be",
		Flags:             "proof-of-stake",
		Proofhash:         "15bd6006ecbab06708f705ecf68664b78b388e4d51416cdafb019d5b90239877",
		Modifier:          "a79c00d1d570743ca8135a173d535258026d26bafbc5f3d951c3d33486b1f120",
	}
	err = mockedClientDoer.AddResponse(qtum.MethodGetBlockHeader, getBlockHeaderResponse)
	if err != nil {
		t.Fatal(err)
	}

	err = mockedClientDoer.AddResponse(qtum.MethodGetBlock, GetBlockResponse)
	if err != nil {
		t.Fatal(err)
	}

	getTransactionResponse := qtum.GetTransactionResponse{
		Amount:            decimal.NewFromFloat(0.20689141),
		Fee:               decimal.NewFromFloat(-0.2012),
		Confirmations:     2,
		BlockHash:         GetTransactionByHashBlockHash,
		BlockIndex:        2,
		BlockTime:         1533092896,
		ID:                "11e97fa5877c5df349934bafc02da6218038a427e8ed081f048626fa6eb523f5",
		Time:              1533092879,
		ReceivedAt:        1533092879,
		Bip125Replaceable: "no",
		Details: []*qtum.TransactionDetail{{Account: "",
			Category:  "send",
			Amount:    decimal.NewFromInt(0),
			Vout:      0,
			Fee:       decimal.NewFromFloat(-0.2012),
			Abandoned: false}},
		Hex: "020000000159c0514feea50f915854d9ec45bc6458bb14419c78b17e7be3f7fd5f563475b5010000006a473044022072d64a1f4ea2d54b7b05050fc853ab192c91cc5ca17e23007867f92f2ab59d9202202b8c9ab9348c8edbb3b98b1788382c8f37642ec9bd6a4429817ab79927319200012103520b1500a400483f19b93c4cb277a2f29693ea9d6739daaf6ae6e971d29e3140feffffff02000000000000000063010403400d0301644440c10f190000000000000000000000006b22910b1e302cf74803ffd1691c2ecb858d3712000000000000000000000000000000000000000000000000000000000000000a14be528c8378ff082e4ba43cb1baa363dbf3f577bfc260e66272970100001976a9146b22910b1e302cf74803ffd1691c2ecb858d371288acb00f0000",
	}
	err = mockedClientDoer.AddResponse(qtum.MethodGetTransaction, getTransactionResponse)
	if err != nil {
		t.Fatal(err)
	}

	decodedRawTransactionResponse := qtum.DecodedRawTransactionResponse{
		ID:       "11e97fa5877c5df349934bafc02da6218038a427e8ed081f048626fa6eb523f5",
		Hash:     "d0fe0caa1b798c36da37e9118a06a7d151632d670b82d1c7dc3985577a71880f",
		Size:     552,
		Vsize:    552,
		Version:  2,
		Locktime: 608,
		Vins: []*qtum.DecodedRawTransactionInV{{
			TxID: "7f5350dc474f2953a3f30282c1afcad2fb61cdcea5bd949c808ecc6f64ce1503",
			Vout: 0,
			ScriptSig: struct {
				Asm string `json:"asm"`
				Hex string `json:"hex"`
			}{
				Asm: "3045022100af4de764705dbd3c0c116d73fe0a2b78c3fab6822096ba2907cfdae2bb28784102206304340a6d260b364ef86d6b19f2b75c5e55b89fb2f93ea72c05e09ee037f60b[ALL] 03520b1500a400483f19b93c4cb277a2f29693ea9d6739daaf6ae6e971d29e3140",
				Hex: "483045022100af4de764705dbd3c0c116d73fe0a2b78c3fab6822096ba2907cfdae2bb28784102206304340a6d260b364ef86d6b19f2b75c5e55b89fb2f93ea72c05e09ee037f60b012103520b1500a400483f19b93c4cb277a2f29693ea9d6739daaf6ae6e971d29e3140",
			},
		}},
		Vouts: []*qtum.DecodedRawTransactionOutV{},
	}
	err = mockedClientDoer.AddResponse(qtum.MethodDecodeRawTransaction, decodedRawTransactionResponse)
	if err != nil {
		t.Fatal(err)
	}

	getTransactionReceiptResponse := qtum.GetTransactionReceiptResponse{}
	err = mockedClientDoer.AddResponse(qtum.MethodGetTransactionReceipt, &getTransactionReceiptResponse)
	if err != nil {
		t.Fatal(err)
	}

	// TODO: Get an actual response for this (only addresses are used in this test though)
	getRawTransactionResponse := qtum.GetRawTransactionResponse{
		Vouts: []qtum.RawTransactionVout{
			{
				Details: struct {
					Addresses []string `json:"addresses"`
					Asm       string   `json:"asm"`
					Hex       string   `json:"hex"`
					// ReqSigs   interface{} `json:"reqSigs"`
					Type string `json:"type"`
				}{
					Addresses: []string{
						"7926223070547d2d15b2ef5e7383e541c338ffe9",
					},
				},
			},
		},
	}
	err = mockedClientDoer.AddResponse(qtum.MethodGetRawTransaction, &getRawTransactionResponse)
	if err != nil {
		t.Fatal(err)
	}
}
