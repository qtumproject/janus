package transformer

import (
	"fmt"
	"testing"

	"github.com/qtumproject/janus/pkg/eth"
	"github.com/qtumproject/janus/pkg/qtum"
	"github.com/shopspring/decimal"
	"github.com/stretchr/testify/require"
)

func TestEthValueToQtumAmount(t *testing.T) {
	cases := []map[string]interface{}{
		{
			"in":   "0xde0b6b3a7640000",
			"want": decimal.NewFromFloat(1),
		},
		{

			"in":   "0x6f05b59d3b20000",
			"want": decimal.NewFromFloat(0.5),
		},
		{
			"in":   "0x2540be400",
			"want": decimal.NewFromFloat(0.00000001),
		},
		{
			"in":   "0x1",
			"want": decimal.NewFromInt(0),
		},
	}
	for _, c := range cases {
		in := c["in"].(string)
		want := c["want"].(decimal.Decimal)
		got, err := EthValueToQtumAmount(in, MinimumGas)
		if err != nil {
			t.Error(err)
		}
		if !got.Equal(want) {
			t.Errorf("in: %s, want: %v, got: %v", in, want, got)
		}
	}
}

func TestQtumValueToEthAmount(t *testing.T) {
	cases := []decimal.Decimal{
		decimal.NewFromFloat(1),
		decimal.NewFromFloat(0.5),
		decimal.NewFromFloat(0.00000001),
		MinimumGas,
	}
	for _, c := range cases {
		in := c
		eth := QtumDecimalValueToETHAmount(in)
		out := EthDecimalValueToQtumAmount(eth)

		if !in.Equals(out) {
			t.Errorf("in: %s, eth: %v, qtum: %v", in, eth, out)
		}
	}
}

func TestQtumAmountToEthValue(t *testing.T) {
	in, want := decimal.NewFromFloat(0.1), "0x16345785d8a0000"
	got, err := formatQtumAmount(in)
	if err != nil {
		t.Error(err)
	}
	if got != want {
		t.Errorf("in: %v, want: %s, got: %s", in, want, got)
	}
}

func TestLowestQtumAmountToEthValue(t *testing.T) {
	in, want := decimal.NewFromFloat(0.00000001), "0x2540be400"
	got, err := formatQtumAmount(in)
	if err != nil {
		t.Error(err)
	}
	if got != want {
		t.Errorf("in: %v, want: %s, got: %s", in, want, got)
	}
}

func TestAddressesConversion(t *testing.T) {
	t.Parallel()

	inputs := []struct {
		qtumChain   string
		ethAddress  string
		qtumAddress string
	}{
		{
			qtumChain:   qtum.ChainTest,
			ethAddress:  "6c89a1a6ca2ae7c00b248bb2832d6f480f27da68",
			qtumAddress: "qTTH1Yr2eKCuDLqfxUyBLCAjmomQ8pyrBt",
		},

		// Test cases for addresses defined here:
		// 	- https://github.com/hayeah/openzeppelin-solidity/blob/qtum/QTUM-NOTES.md#create-test-accounts
		//
		// NOTE: Ethereum addresses are without `0x` prefix, as it expects by conversion functions
		{
			qtumChain:   qtum.ChainTest,
			ethAddress:  "7926223070547d2d15b2ef5e7383e541c338ffe9",
			qtumAddress: "qUbxboqjBRp96j3La8D1RYkyqx5uQbJPoW",
		},
		{
			qtumChain:   qtum.ChainTest,
			ethAddress:  "2352be3db3177f0a07efbe6da5857615b8c9901d",
			qtumAddress: "qLn9vqbr2Gx3TsVR9QyTVB5mrMoh4x43Uf",
		},
		{
			qtumChain:   qtum.ChainTest,
			ethAddress:  "69b004ac2b3993bf2fdf56b02746a1f57997420d",
			qtumAddress: "qTCCy8qy7pW94EApdoBjYc1vQ2w68UnXPi",
		},
		{
			qtumChain:   qtum.ChainTest,
			ethAddress:  "8c647515f03daeefd09872d7530fa8d8450f069a",
			qtumAddress: "qWMi6ne9mDQFatRGejxdDYVUV9rQVkAFGp",
		},
		{
			qtumChain:   qtum.ChainTest,
			ethAddress:  "2191744eb5ebeac90e523a817b77a83a0058003b",
			qtumAddress: "qLcshhsRS6HKeTKRYFdpXnGVZxw96QQcfm",
		},
		{
			qtumChain:   qtum.ChainTest,
			ethAddress:  "88b0bf4b301c21f8a47be2188bad6467ad556dcf",
			qtumAddress: "qW28njWueNpBXYWj2KDmtFG2gbLeALeHfV",
		},
	}

	for i, in := range inputs {
		var (
			in       = in
			testDesc = fmt.Sprintf("#%d", i)
		)
		t.Run(testDesc, func(t *testing.T) {
			qtumAddress, err := convertETHAddress(in.ethAddress, in.qtumChain)
			require.NoError(t, err, "couldn't convert Ethereum address to Qtum address")
			require.Equal(t, in.qtumAddress, qtumAddress, "unexpected converted Qtum address value")

			ethAddress, err := convertQtumAddress(in.qtumAddress)
			require.NoError(t, err, "couldn't convert Qtum address to Ethereum address")
			require.Equal(t, in.ethAddress, ethAddress, "unexpected converted Ethereum address value")
		})
	}
}

func TestSendTransactionRequestHasDefaultGasPriceAndAmount(t *testing.T) {
	var req eth.SendTransactionRequest
	err := unmarshalRequest([]byte(`[{}]`), &req)
	if err != nil {
		t.Fatal(err)
	}
	defaultGasPriceInWei := req.GasPrice.Int
	defaultGasPriceInQTUM := EthDecimalValueToQtumAmount(decimal.NewFromBigInt(defaultGasPriceInWei, 1))
	if !defaultGasPriceInQTUM.Equals(MinimumGas) {
		t.Fatalf("Default gas price does not convert to QTUM minimum gas price, got: %s want: %s", defaultGasPriceInQTUM.String(), MinimumGas.String())
	}
	if eth.DefaultGasAmountForQtum.String() != req.Gas.Int.String() {
		t.Fatalf("Default gas amount does not match expected default, got: %s want: %s", req.Gas.Int.String(), eth.DefaultGasAmountForQtum.String())
	}
}
