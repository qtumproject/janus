package transformer

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/qtumproject/janus/pkg/eth"
	"github.com/qtumproject/janus/pkg/internal"
	"github.com/qtumproject/janus/pkg/qtum"
)

func TestGetLogs(t *testing.T) {
	testGetLogsWithTopics(
		t,
		[]interface{}{
			"0f6798a560793a54c3bcfe86a93cde1e73087d944c0ea20544137d4121396885",
			"0000000000000000000000006b22910b1e302cf74803ffd1691c2ecb858d3712",
		},
		eth.GetLogsResponse{
			{
				LogIndex:         "0x0",
				TransactionIndex: "0x2",
				TransactionHash:  "0xc1816e5fbdd4d1cc62394be83c7c7130ccd2aadefcd91e789c1a0b33ec093fef",
				BlockHash:        "0x975326b65c20d0b8500f00a59f76b08a98513fff7ce0484382534a47b55f8985",
				BlockNumber:      "0xfdf",
				Address:          "0xdb46f738bf32cdafb9a4a70eb8b44c76646bcaf0",
				Data:             "0x0000000000000000000000000000000000000000000000000000000000000001",
				Topics: []string{
					"0x0f6798a560793a54c3bcfe86a93cde1e73087d944c0ea20544137d4121396885",
					"0x0000000000000000000000006b22910b1e302cf74803ffd1691c2ecb858d3712",
				},
			},
		},
	)
}

func TestGetLogsFiltersByFirstTopic(t *testing.T) {
	testGetLogsWithTopics(
		t,
		[]interface{}{
			"0f6798a560793a54c3bcfe86a93cde1e73087d944c0ea20544137d4121396885",
		},
		eth.GetLogsResponse{
			{
				LogIndex:         "0x0",
				TransactionIndex: "0x2",
				TransactionHash:  "0xc1816e5fbdd4d1cc62394be83c7c7130ccd2aadefcd91e789c1a0b33ec093fef",
				BlockHash:        "0x975326b65c20d0b8500f00a59f76b08a98513fff7ce0484382534a47b55f8985",
				BlockNumber:      "0xfdf",
				Address:          "0xdb46f738bf32cdafb9a4a70eb8b44c76646bcaf0",
				Data:             "0x0000000000000000000000000000000000000000000000000000000000000001",
				Topics: []string{
					"0x0f6798a560793a54c3bcfe86a93cde1e73087d944c0ea20544137d4121396885",
					"0x0000000000000000000000006b22910b1e302cf74803ffd1691c2ecb858d3712",
				},
			},
		},
	)
}

func TestGetLogsFiltersBySecondTopic(t *testing.T) {
	testGetLogsWithTopics(
		t,
		[]interface{}{
			nil,
			"0x0000000000000000000000006b22910b1e302cf74803ffd1691c2ecb858d3712",
		},
		eth.GetLogsResponse{
			{
				LogIndex:         "0x0",
				TransactionIndex: "0x2",
				TransactionHash:  "0xc1816e5fbdd4d1cc62394be83c7c7130ccd2aadefcd91e789c1a0b33ec093fef",
				BlockHash:        "0x975326b65c20d0b8500f00a59f76b08a98513fff7ce0484382534a47b55f8985",
				BlockNumber:      "0xfdf",
				Address:          "0xdb46f738bf32cdafb9a4a70eb8b44c76646bcaf0",
				Data:             "0x0000000000000000000000000000000000000000000000000000000000000001",
				Topics: []string{
					"0x0f6798a560793a54c3bcfe86a93cde1e73087d944c0ea20544137d4121396885",
					"0x0000000000000000000000006b22910b1e302cf74803ffd1691c2ecb858d3712",
				},
			},
		},
	)
}

func TestGetLogsFiltersWithOR(t *testing.T) {
	testGetLogsWithTopics(
		t,
		[]interface{}{
			nil,
			[]interface{}{
				"0x0000000000000000000000006b22910b1e302cf74803ffd1691c2ecb858d3712",
				"0x0000000000000000000000006b22910b1e302cf74803ffd1691c2ecb858d3713",
			},
		},
		eth.GetLogsResponse{
			{
				LogIndex:         "0x0",
				TransactionIndex: "0x2",
				TransactionHash:  "0xc1816e5fbdd4d1cc62394be83c7c7130ccd2aadefcd91e789c1a0b33ec093fef",
				BlockHash:        "0x975326b65c20d0b8500f00a59f76b08a98513fff7ce0484382534a47b55f8985",
				BlockNumber:      "0xfdf",
				Address:          "0xdb46f738bf32cdafb9a4a70eb8b44c76646bcaf0",
				Data:             "0x0000000000000000000000000000000000000000000000000000000000000001",
				Topics: []string{
					"0x0f6798a560793a54c3bcfe86a93cde1e73087d944c0ea20544137d4121396885",
					"0x0000000000000000000000006b22910b1e302cf74803ffd1691c2ecb858d3712",
				},
			},
		},
	)
}

func TestGetLogsFiltersByTopic(t *testing.T) {
	testGetLogsWithTopics(
		t,
		[]interface{}{
			"a topic",
			"another topic",
		},
		eth.GetLogsResponse{},
	)
}

func TestMultipleLogsWithORdTopics(t *testing.T) {
	rawResponse := `
	{
		"error": null,
		"id": "2780",
		"result": [
		  {
			"blockHash": "1544b64a182e96ea91a0cebf979a412681db4c2a84186ff552f71bdf45284d05",
			"blockNumber": 732610,
			"bloom": "0004000000000000000000008000800000000000000000000000000000000100000004000000000000000000000004000000000000000000000000000000000040010000000000000000000a000000000000000000000000200000008000000000008000021000000200000020008800000000000008000200004010000000020000000000000000000000000100000400000001000000080000008000000000000000020800000000000000000000000000000000000000000000000000000000000002000000000800001000020040000000000000001000000000000020000000000000000004080000000000002000000000001000400100000000000000",
			"contractAddress": "d4915308a9c4c40f57b0eccc63ee70616982842b",
			"cumulativeGasUsed": 2698989,
			"excepted": "None",
			"exceptedMessage": "",
			"from": "20384aea059a56d9444ac8f355c4988b620259e7",
			"gasUsed": 2698989,
			"log": [
			  {
				"address": "284937a9f5a1d28268d4e48d5eda03b04a7a1786",
				"data": "000000000000000000000000b406040d9e1a9bbb19fcc803a7a808b038ae45ce0000000000000000000000000000000000000000000000000000000000000003",
				"topics": [
				  "0d3648bd0f6ba80134a33ba9275ac585d9d315f0ad8355cddefde31afa28d0e9",
				  "000000000000000000000000e7e5caae57b34b93c57af9478a5130f62e3d2827",
				  "000000000000000000000000f2033ede578e17fa6231047265010445bca8cf1c"
				]
			  },
			  {
				"address": "f2033ede578e17fa6231047265010445bca8cf1c",
				"data": "0000000000000000000000000000000000000000000000000000009adb0c9b00",
				"topics": [
				  "ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",
				  "00000000000000000000000020384aea059a56d9444ac8f355c4988b620259e7",
				  "000000000000000000000000b406040d9e1a9bbb19fcc803a7a808b038ae45ce"
				]
			  },
			  {
				"address": "e7e5caae57b34b93c57af9478a5130f62e3d2827",
				"data": "0000000000000000000000000000000000000000000000000000000ba43b7400",
				"topics": [
				  "e1fffcc4923d04b559f4d29a8bfc6cda04eb5b0d3c460751c2402c5c5cc9109c",
				  "000000000000000000000000d4915308a9c4c40f57b0eccc63ee70616982842b"
				]
			  },
			  {
				"address": "e7e5caae57b34b93c57af9478a5130f62e3d2827",
				"data": "0000000000000000000000000000000000000000000000000000000ba43b7400",
				"topics": [
				  "ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",
				  "000000000000000000000000d4915308a9c4c40f57b0eccc63ee70616982842b",
				  "000000000000000000000000b406040d9e1a9bbb19fcc803a7a808b038ae45ce"
				]
			  },
			  {
				"address": "b406040d9e1a9bbb19fcc803a7a808b038ae45ce",
				"data": "00000000000000000000000000000000000000000000000000000000000003e8",
				"topics": [
				  "ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",
				  "0000000000000000000000000000000000000000000000000000000000000000",
				  "0000000000000000000000000000000000000000000000000000000000000000"
				]
			  },
			  {
				"address": "b406040d9e1a9bbb19fcc803a7a808b038ae45ce",
				"data": "0000000000000000000000000000000000000000000000000000002a7579a9a1",
				"topics": [
				  "ddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",
				  "0000000000000000000000000000000000000000000000000000000000000000",
				  "00000000000000000000000020384aea059a56d9444ac8f355c4988b620259e7"
				]
			  },
			  {
				"address": "b406040d9e1a9bbb19fcc803a7a808b038ae45ce",
				"data": "0000000000000000000000000000000000000000000000000000000ba43b74000000000000000000000000000000000000000000000000000000009adb0c9b00",
				"topics": [
				  "1c411e9a96e071241c2f21f7726b17ae89e3cab4c78be50e062b03a9fffbbad1"
				]
			  },
			  {
				"address": "b406040d9e1a9bbb19fcc803a7a808b038ae45ce",
				"data": "0000000000000000000000000000000000000000000000000000000ba43b74000000000000000000000000000000000000000000000000000000009adb0c9b00",
				"topics": [
				  "4c209b5fc8ad50758f13e2e1088ba56a560dff690a1c6fef26394f4c03821c4f",
				  "000000000000000000000000d4915308a9c4c40f57b0eccc63ee70616982842b"
				]
			  }
			],
			"outputIndex": 0,
			"stateRoot": "55a8c7a4df650d321bc702a2606a85445488f9d00fbc88096a459bf75a6de80f",
			"to": "d4915308a9c4c40f57b0eccc63ee70616982842b",
			"transactionHash": "626fd7f009f08ab16f73d413790b5db52b56dfdbdc9aafbc405e1e07ef3b539c",
			"transactionIndex": 2,
			"utxoRoot": "1ed0668459fb83d261796294af09f72091f21968c3c2cef7a10a0317cb2947b4"
		  }
		]
	}
	`

	//preparing request
	fromBlock, err := json.Marshal("0xb2dc2")
	toBlock, err := json.Marshal("0xb2dc2")
	address, err := json.Marshal("0xb406040d9e1a9bbb19fcc803a7a808b038ae45ce")

	request := eth.GetLogsRequest{
		FromBlock: fromBlock,
		ToBlock:   toBlock,
		Address:   address,
		Topics: []interface{}{
			[]interface{}{
				"0x4c209b5fc8ad50758f13e2e1088ba56a560dff690a1c6fef26394f4c03821c4f",
				"0xdccd412f0b1252819cb1fd330b93224ca42612892bb3f4f789976e6d81936496",
				"0xd78ad95fa46c994b6551d0da85fc275fe613ce37657fb8d5e3d130840159d822",
				"0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",
				"0x1c411e9a96e071241c2f21f7726b17ae89e3cab4c78be50e062b03a9fffbbad1",
			},
		},
	}

	requestRaw, err := json.Marshal(&request)
	if err != nil {
		t.Fatal(err)
	}

	requestParamsArray := []json.RawMessage{requestRaw}
	requestRPC, err := internal.PrepareEthRPCRequest(1, requestParamsArray)
	if err != nil {
		t.Fatal(err)
	}

	clientDoerMock := internal.NewDoerMappedMock()
	qtumClient, err := internal.CreateMockedClient(clientDoerMock)
	if err != nil {
		t.Fatal(err)
	}

	//Add response
	clientDoerMock.AddRawResponse(qtum.MethodSearchLogs, []byte(rawResponse))

	//Prepare proxy & execute
	//preparing proxy & executing
	proxyEth := ProxyETHGetLogs{qtumClient}

	got, err := proxyEth.Request(requestRPC, nil)
	if err != nil {
		t.Fatal(err)
	}

	want := eth.GetLogsResponse{
		{
			LogIndex:         "0x4",
			TransactionIndex: "0x2",
			TransactionHash:  "0x626fd7f009f08ab16f73d413790b5db52b56dfdbdc9aafbc405e1e07ef3b539c",
			BlockHash:        "0x1544b64a182e96ea91a0cebf979a412681db4c2a84186ff552f71bdf45284d05",
			BlockNumber:      "0xb2dc2",
			Address:          "0xb406040d9e1a9bbb19fcc803a7a808b038ae45ce",
			Data:             "0x00000000000000000000000000000000000000000000000000000000000003e8",
			Topics: []string{
				"0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",
				"0x0000000000000000000000000000000000000000000000000000000000000000",
				"0x0000000000000000000000000000000000000000000000000000000000000000",
			},
		},
		{
			LogIndex:         "0x5",
			TransactionIndex: "0x2",
			TransactionHash:  "0x626fd7f009f08ab16f73d413790b5db52b56dfdbdc9aafbc405e1e07ef3b539c",
			BlockHash:        "0x1544b64a182e96ea91a0cebf979a412681db4c2a84186ff552f71bdf45284d05",
			BlockNumber:      "0xb2dc2",
			Address:          "0xb406040d9e1a9bbb19fcc803a7a808b038ae45ce",
			Data:             "0x0000000000000000000000000000000000000000000000000000002a7579a9a1",
			Topics: []string{
				"0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef",
				"0x0000000000000000000000000000000000000000000000000000000000000000",
				"0x00000000000000000000000020384aea059a56d9444ac8f355c4988b620259e7",
			},
		},
		{
			LogIndex:         "0x6",
			TransactionIndex: "0x2",
			TransactionHash:  "0x626fd7f009f08ab16f73d413790b5db52b56dfdbdc9aafbc405e1e07ef3b539c",
			BlockHash:        "0x1544b64a182e96ea91a0cebf979a412681db4c2a84186ff552f71bdf45284d05",
			BlockNumber:      "0xb2dc2",
			Address:          "0xb406040d9e1a9bbb19fcc803a7a808b038ae45ce",
			Data:             "0x0000000000000000000000000000000000000000000000000000000ba43b74000000000000000000000000000000000000000000000000000000009adb0c9b00",
			Topics: []string{
				"0x1c411e9a96e071241c2f21f7726b17ae89e3cab4c78be50e062b03a9fffbbad1",
			},
		},
		{
			LogIndex:         "0x7",
			TransactionIndex: "0x2",
			TransactionHash:  "0x626fd7f009f08ab16f73d413790b5db52b56dfdbdc9aafbc405e1e07ef3b539c",
			BlockHash:        "0x1544b64a182e96ea91a0cebf979a412681db4c2a84186ff552f71bdf45284d05",
			BlockNumber:      "0xb2dc2",
			Address:          "0xb406040d9e1a9bbb19fcc803a7a808b038ae45ce",
			Data:             "0x0000000000000000000000000000000000000000000000000000000ba43b74000000000000000000000000000000000000000000000000000000009adb0c9b00",
			Topics: []string{
				"0x4c209b5fc8ad50758f13e2e1088ba56a560dff690a1c6fef26394f4c03821c4f",
				"0x000000000000000000000000d4915308a9c4c40f57b0eccc63ee70616982842b",
			},
		},
	}

	if !reflect.DeepEqual(got, &want) {
		t.Errorf(
			"error\ninput: %s\nwant: %s\ngot: %s",
			requestRPC,
			string(internal.MustMarshalIndent(want, "", "  ")),
			string(internal.MustMarshalIndent(got, "", "  ")),
		)
	}
}

func testGetLogsWithTopics(t *testing.T, topics []interface{}, want eth.GetLogsResponse) {
	//preparing request
	fromBlock, err := json.Marshal("0xfde")
	toBlock, err := json.Marshal("0xfde")
	address, err := json.Marshal("0xdb46f738bf32cdafb9a4a70eb8b44c76646bcaf0")

	request := eth.GetLogsRequest{
		FromBlock: fromBlock,
		ToBlock:   toBlock,
		Address:   address,
		Topics:    topics,
	}

	requestRaw, err := json.Marshal(&request)
	if err != nil {
		t.Fatal(err)
	}

	requestParamsArray := []json.RawMessage{requestRaw}
	requestRPC, err := internal.PrepareEthRPCRequest(1, requestParamsArray)
	if err != nil {
		t.Fatal(err)
	}

	clientDoerMock := internal.NewDoerMappedMock()
	qtumClient, err := internal.CreateMockedClient(clientDoerMock)
	if err != nil {
		t.Fatal(err)
	}
	//prepare response
	searchLogsResponse := qtum.SearchLogsResponse{
		{
			BlockHash:         "975326b65c20d0b8500f00a59f76b08a98513fff7ce0484382534a47b55f8985",
			BlockNumber:       4063,
			TransactionHash:   "c1816e5fbdd4d1cc62394be83c7c7130ccd2aadefcd91e789c1a0b33ec093fef",
			TransactionIndex:  2,
			From:              "6b22910b1e302cf74803ffd1691c2ecb858d3712",
			To:                "db46f738bf32cdafb9a4a70eb8b44c76646bcaf0",
			CumulativeGasUsed: 68572,
			GasUsed:           68572,
			ContractAddress:   "db46f738bf32cdafb9a4a70eb8b44c76646bcaf0",
			Log: []qtum.Log{
				{
					Address: "db46f738bf32cdafb9a4a70eb8b44c76646bcaf0",
					Topics: []string{
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
	err = clientDoerMock.AddResponseWithRequestID(2, qtum.MethodSearchLogs, searchLogsResponse)
	if err != nil {
		t.Fatal(err)
	}

	//Prepare proxy & execute
	//preparing proxy & executing
	proxyEth := ProxyETHGetLogs{qtumClient}

	got, err := proxyEth.Request(requestRPC, nil)
	if err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(got, &want) {
		t.Errorf(
			"error\ninput: %s\nwant: %s\ngot: %s",
			requestRPC,
			string(internal.MustMarshalIndent(want, "", "  ")),
			string(internal.MustMarshalIndent(got, "", "  ")),
		)
	}
}

func TestGetLogsTranslateTopicWorksWithNil(t *testing.T) {
	fromBlock, err := json.Marshal("0xfde")
	toBlock, err := json.Marshal("0xfde")
	address, err := json.Marshal("0xdb46f738bf32cdafb9a4a70eb8b44c76646bcaf0")

	request := eth.GetLogsRequest{
		FromBlock: fromBlock,
		ToBlock:   toBlock,
		Address:   address,
		Topics: []interface{}{
			"0f6798a560793a54c3bcfe86a93cde1e73087d944c0ea20544137d4121396885",
			nil,
		},
	}

	translatedTopics, err := eth.TranslateTopics(request.Topics)
	if err != nil {
		t.Fatal(err)
	}
	if len(translatedTopics) != 2 {
		t.Fatalf("Unexpected translated topic length: %d", len(translatedTopics))
	}
	if translatedTopics[1] != nil {
		t.Fatalf("Expected nil for topic 2, got: %v", translatedTopics[1])
	}

	clientDoerMock := internal.NewDoerMappedMock()
	qtumClient, err := internal.CreateMockedClient(clientDoerMock)
	if err != nil {
		t.Fatal(err)
	}

	//Prepare proxy & execute
	//preparing proxy & executing
	proxyEth := ProxyETHGetLogs{qtumClient}

	qtumRequest, err := proxyEth.ToRequest(&request)
	if err != nil {
		t.Fatal(err)
	}

	qtumRawRequest, err := json.Marshal(qtumRequest)
	if err != nil {
		t.Fatal(err)
	}

	expectedRawRequest := `[4062,4062,{"addresses":["db46f738bf32cdafb9a4a70eb8b44c76646bcaf0"]},{"topics":["0f6798a560793a54c3bcfe86a93cde1e73087d944c0ea20544137d4121396885",null]}]`

	if expectedRawRequest != string(qtumRawRequest) {
		t.Errorf(
			"error\ninput: %s\nwant: %s\ngot: %s",
			qtumRawRequest,
			string(internal.MustMarshalIndent(expectedRawRequest, "", "  ")),
			string(internal.MustMarshalIndent(string(qtumRawRequest), "", "  ")),
		)
	}
}
