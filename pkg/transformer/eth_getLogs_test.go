package transformer

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/qtumproject/janus/pkg/eth"
	"github.com/qtumproject/janus/pkg/qtum"

)

func TestGetLogs(t *testing.T) {
	//perparing request
	fromBlock, err := json.Marshal("0xfde")
	toBlock, err := json.Marshal("0xfde")
	address, err := json.Marshal("0xdb46f738bf32cdafb9a4a70eb8b44c76646bcaf0")

	request := eth.GetLogsRequest{
		FromBlock:	fromBlock,
		ToBlock:	toBlock,
		Address: 	address,
		Topics:	   []interface{}{
			"0f6798a560793a54c3bcfe86a93cde1e73087d944c0ea20544137d4121396885",
			"0000000000000000000000006b22910b1e302cf74803ffd1691c2ecb858d3712",
		},
	}
	
	requestRaw, err := json.Marshal(&request)
	if err != nil {
		panic(err)
	}

	requestParamsArray := []json.RawMessage{requestRaw}
	requestRPC, err :=  prepareEthRPCRequest(1, requestParamsArray)
	if err != nil {
		panic(err)
	}

	clientDoerMock := doerMappedMock{make(map[string][]byte)}
	qtumClient, err := createMockedClient(clientDoerMock)
	if err != nil {
		panic(err)
	}
	//prepare response
	searchLogsResponse := qtum.SearchLogsResponse{
		{
				BlockHash:	"975326b65c20d0b8500f00a59f76b08a98513fff7ce0484382534a47b55f8985",
				BlockNumber: 4063,
				TransactionHash: "c1816e5fbdd4d1cc62394be83c7c7130ccd2aadefcd91e789c1a0b33ec093fef",
				TransactionIndex: 2,
				From: "6b22910b1e302cf74803ffd1691c2ecb858d3712",
				To: "db46f738bf32cdafb9a4a70eb8b44c76646bcaf0",
				CumulativeGasUsed: 68572,
				GasUsed: 68572,
				ContractAddress: "db46f738bf32cdafb9a4a70eb8b44c76646bcaf0",
				Log: []qtum.Log{
					{
						Address: "db46f738bf32cdafb9a4a70eb8b44c76646bcaf0",
						Topics:	 []string{
							"0f6798a560793a54c3bcfe86a93cde1e73087d944c0ea20544137d4121396885",
	           				"0000000000000000000000006b22910b1e302cf74803ffd1691c2ecb858d3712",
						},
						Data: "0000000000000000000000000000000000000000000000000000000000000001",
					},
				},
				Excepted: "None",
		},
	}

	//Add response
	err = clientDoerMock.AddResponse(2, qtum.MethodSearchLogs, searchLogsResponse)
	if err != nil {
		panic(err)
	}

	//Prepare proxy & execute
	//preparing proxy & executing
	proxyEth := ProxyETHGetLogs{qtumClient}

	got, err := proxyEth.Request(requestRPC)
	if err != nil {
		panic(err)
	}

	want := eth.GetLogsResponse{
		{
				LogIndex: "0x0",
				TransactionIndex: "0x2",
				TransactionHash: "0xc1816e5fbdd4d1cc62394be83c7c7130ccd2aadefcd91e789c1a0b33ec093fef",
				BlockHash: "0x975326b65c20d0b8500f00a59f76b08a98513fff7ce0484382534a47b55f8985",
				BlockNumber: "0xfdf",
				Address: "0xdb46f738bf32cdafb9a4a70eb8b44c76646bcaf0",
				Data: "0x0000000000000000000000000000000000000000000000000000000000000001",
				Topics:	 []string{
							"0x0f6798a560793a54c3bcfe86a93cde1e73087d944c0ea20544137d4121396885",
	           				"0x0000000000000000000000006b22910b1e302cf74803ffd1691c2ecb858d3712",
				},
		},
	}
	if !reflect.DeepEqual(got, &want){
		t.Errorf(
			"error\ninput: %s\nwant: %s\ngot: %s",
			requestRPC,
			string(mustMarshalIndent(want, "", "  ")),
			string(mustMarshalIndent(got, "", "  ")),
		)
	}
}