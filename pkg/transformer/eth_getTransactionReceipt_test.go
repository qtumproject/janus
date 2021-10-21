package transformer

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/qtumproject/janus/pkg/eth"
	"github.com/qtumproject/janus/pkg/internal"
	"github.com/qtumproject/janus/pkg/qtum"
	"github.com/qtumproject/janus/pkg/utils"
)

func TestGetTransactionReceiptForNonVMTransaction(t *testing.T) {
	//preparing request
	requestParams := []json.RawMessage{[]byte(`"0x8fcd819194cce6a8454b2bec334d3448df4f097e9cdc36707bfd569900268950"`)}
	request, err := internal.PrepareEthRPCRequest(1, requestParams)
	if err != nil {
		t.Fatal(err)
	}

	mockedClientDoer := internal.NewDoerMappedMock()
	qtumClient, err := internal.CreateMockedClient(mockedClientDoer)
	if err != nil {
		t.Fatal(err)
	}

	//preparing client response
	err = mockedClientDoer.AddResponseWithRequestID(2, qtum.MethodGetTransactionReceipt, []byte("[]"))
	if err != nil {
		t.Fatal(err)
	}

	rawTransactionResponse := &qtum.GetRawTransactionResponse{
		BlockHash: internal.GetTransactionByHashBlockHash,
	}
	err = mockedClientDoer.AddResponseWithRequestID(3, qtum.MethodGetRawTransaction, rawTransactionResponse)
	if err != nil {
		t.Fatal(err)
	}

	err = mockedClientDoer.AddResponseWithRequestID(4, qtum.MethodGetBlock, internal.GetBlockResponse)
	if err != nil {
		t.Fatal(err)
	}

	//preparing proxy & executing request
	proxyEth := ProxyETHGetTransactionReceipt{qtumClient}
	got, err := proxyEth.Request(request, nil)
	if err != nil {
		t.Fatal(err)
	}

	want := eth.GetTransactionReceiptResponse{
		TransactionHash:   "0x8fcd819194cce6a8454b2bec334d3448df4f097e9cdc36707bfd569900268950",
		TransactionIndex:  "0x1",
		BlockHash:         "0xbba11e1bacc69ba535d478cf1f2e542da3735a517b0b8eebaf7e6bb25eeb48c5",
		BlockNumber:       "0xf8f",
		GasUsed:           NonContractVMGasLimit,
		Logs:              []eth.Log{},
		EffectiveGasPrice: "0x0",
		CumulativeGasUsed: NonContractVMGasLimit,
		To:                utils.AddHexPrefix(qtum.ZeroAddress),
		From:              utils.AddHexPrefix(qtum.ZeroAddress),
		LogsBloom:         eth.EmptyLogsBloom,
		Status:            STATUS_SUCCESS,
	}
	if !reflect.DeepEqual(got, &want) {
		t.Errorf(
			"error\ninput: %s\nwant: %s\ngot: %s",
			request,
			internal.MustMarshalIndent(want, "", " "),
			internal.MustMarshalIndent(got, "", " "),
		)
	}
}
