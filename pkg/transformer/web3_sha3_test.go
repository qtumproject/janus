package transformer

import (
	"encoding/json"
	"testing"

	"github.com/qtumproject/janus/pkg/internal"
)

func TestWeb3Sha3Request(t *testing.T) {
	values := make(map[string]string)
	values[""] = "0xc5d2460186f7233c927e7db2dcc703c0e500b653ca82273b7bfad8045d85a470"
	values["0x00"] = "0xbc36789e7a1e281436464229828f817d6612f7b477d66591ff96a9e064bcc98a"
	values["0x68656c6c6f20776f726c64"] = "0x47173285a8d7341e5e972fc677286384f802f8ef42a5ec5f03bbfa254cb01fad"

	for input, want := range values {
		requestParams := []json.RawMessage{[]byte(`"` + input + `"`)}
		request, err := internal.PrepareEthRPCRequest(1, requestParams)
		if err != nil {
			t.Fatal(err)
		}

		web3Sha3 := Web3Sha3{}
		got, err := web3Sha3.Request(request, nil)
		if err != nil {
			t.Fatal(err)
		}

		if got != want {
			t.Errorf(
				"error\ninput: %s\nwant: %s\ngot: %s",
				input,
				string(internal.MustMarshalIndent(want, "", "  ")),
				string(internal.MustMarshalIndent(got, "", "  ")),
			)
		}
	}
}

func TestWeb3Sha3Errors(t *testing.T) {
	testWeb3Sha3Errors(t, []json.RawMessage{}, "missing value for required argument 0")
	testWeb3Sha3Errors(t, []json.RawMessage{[]byte(`"0x00"`), []byte(`"0x00"`)}, "too many arguments, want at most 1")
}

func testWeb3Sha3Errors(t *testing.T, input []json.RawMessage, want string) {
	requestParams := input
	request, err := internal.PrepareEthRPCRequest(1, requestParams)
	if err != nil {
		t.Fatal(err)
	}

	web3Sha3 := Web3Sha3{}
	got, err := web3Sha3.Request(request, nil)
	if err == nil {
		t.Errorf(
			"Expected error\ninput: %s\nwant: %s\ngot: %s",
			input,
			want,
			got,
		)
	} else if err.Error() != want {
		t.Errorf(
			"Unexpected error\ninput: %s\nwant: %s\ngot: %s",
			input,
			want,
			err.Error(),
		)
	}
}
