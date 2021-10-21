package transformer

import (
	"encoding/json"
	"math/big"
	"reflect"
	"testing"

	"github.com/qtumproject/janus/pkg/eth"
	"github.com/qtumproject/janus/pkg/qtum"
	"github.com/qtumproject/janus/pkg/internal"
)

func TestGetFilterChangesRequest_EmptyResult(t *testing.T) {
	//prepare request
	requestParams := []json.RawMessage{[]byte(`"0x1"`)}
	requestRPC, err := internal.PrepareEthRPCRequest(1, requestParams)
	if err != nil {
		t.Fatal(err)
	}
	//prepare client
	mockedClientDoer := internal.NewDoerMappedMock()
	qtumClient, err := internal.CreateMockedClient(mockedClientDoer)
	if err != nil {
		t.Fatal(err)
	}

	//preparing client response
	getBlockCountResponse := qtum.GetBlockCountResponse{Int: big.NewInt(657660)}
	err = mockedClientDoer.AddResponseWithRequestID(2, qtum.MethodGetBlockCount, getBlockCountResponse)
	if err != nil {
		t.Fatal(err)
	}

	searchLogsResponse := qtum.SearchLogsResponse{
		//TODO: add
	}
	err = mockedClientDoer.AddResponseWithRequestID(2, qtum.MethodSearchLogs, searchLogsResponse)
	if err != nil {
		t.Fatal(err)
	}

	//preparing filter
	filterSimulator := eth.NewFilterSimulator()
	filterRequest := eth.NewFilterRequest{}
	filterSimulator.New(eth.NewFilterTy, &filterRequest)
	_filter, _ := filterSimulator.Filter(1)
	filter := _filter.(*eth.Filter)
	filter.Data.Store("lastBlockNumber", uint64(657655))

	//preparing proxy & executing request
	proxyEth := ProxyETHGetFilterChanges{qtumClient, filterSimulator}
	got, err := proxyEth.Request(requestRPC, nil)
	if err != nil {
		t.Fatal(err)
	}

	want := eth.GetFilterChangesResponse{}
	if !reflect.DeepEqual(got, want) {
		t.Errorf(
			"error\ninput: %s\nwant: %s\ngot: %s",
			requestRPC,
			string(internal.MustMarshalIndent(want, "", "  ")),
			string(internal.MustMarshalIndent(got, "", "  ")),
		)
	}
}

func TestGetFilterChangesRequest_NoNewBlocks(t *testing.T) {
	//prepare request
	requestParams := []json.RawMessage{[]byte(`"0x1"`)}
	requestRPC, err := internal.PrepareEthRPCRequest(1, requestParams)
	if err != nil {
		t.Fatal(err)
	}
	//prepare client
	mockedClientDoer := internal.NewDoerMappedMock()
	qtumClient, err := internal.CreateMockedClient(mockedClientDoer)
	if err != nil {
		t.Fatal(err)
	}

	//preparing client response
	getBlockCountResponse := qtum.GetBlockCountResponse{Int: big.NewInt(657655)}
	err = mockedClientDoer.AddResponseWithRequestID(2, qtum.MethodGetBlockCount, getBlockCountResponse)
	if err != nil {
		t.Fatal(err)
	}

	//preparing filter
	filterSimulator := eth.NewFilterSimulator()
	filterSimulator.New(eth.NewFilterTy, nil)
	_filter, _ := filterSimulator.Filter(1)
	filter := _filter.(*eth.Filter)
	filter.Data.Store("lastBlockNumber", uint64(657655))

	//preparing proxy & executing request
	proxyEth := ProxyETHGetFilterChanges{qtumClient, filterSimulator}
	got, err := proxyEth.Request(requestRPC, nil)
	if err != nil {
		t.Fatal(err)
	}

	want := eth.GetFilterChangesResponse{}
	if !reflect.DeepEqual(got, want) {
		t.Errorf(
			"error\ninput: %s\nwant: %s\ngot: %s",
			requestRPC,
			string(internal.MustMarshalIndent(want, "", "  ")),
			string(internal.MustMarshalIndent(got, "", "  ")),
		)
	}
}

func TestGetFilterChangesRequest_NoSuchFilter(t *testing.T) {
	//prepare request
	requestParams := []json.RawMessage{[]byte(`"0x1"`)}
	requestRPC, err := internal.PrepareEthRPCRequest(1, requestParams)
	if err != nil {
		t.Fatal(err)
	}
	//prepare client
	mockedClientDoer := internal.NewDoerMappedMock()
	qtumClient, err := internal.CreateMockedClient(mockedClientDoer)
	if err != nil {
		t.Fatal(err)
	}

	//preparing proxy & executing request
	filterSimulator := eth.NewFilterSimulator()
	proxyEth := ProxyETHGetFilterChanges{qtumClient, filterSimulator}
	got, err := proxyEth.Request(requestRPC, nil)
	expectedErr := "Invalid filter id"

	if got != nil {
		t.Errorf(
			"error\ninput: %s\nwant: %s\ngot: %s",
			requestRPC,
			string(internal.MustMarshalIndent(expectedErr, "", "  ")),
			string(internal.MustMarshalIndent(err.Error(), "", "  ")),
		)
	}
	if err.Error() != expectedErr {
		t.Errorf(
			"error\ninput: %s\nwant error: %s\ngot: %s",
			requestRPC,
			string(internal.MustMarshalIndent(expectedErr, "", "  ")),
			string(internal.MustMarshalIndent(err.Error(), "", "  ")),
		)
	}
}
