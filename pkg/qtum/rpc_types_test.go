package qtum

import (
	"encoding/json"
	"math/big"
	"testing"
)

func TestSearchLogsRequestFiltersTopicsIfAllNull(t *testing.T) {
	expected := `[1,2,{"addresses":["0x1","0x2"]},null,1]`
	minConfs := uint(1)
	request := &SearchLogsRequest{
		FromBlock: big.NewInt(1),
		ToBlock:   big.NewInt(2),
		Addresses: []string{"0x1", "0x2"},
		Topics: []SearchLogsTopic{
			{"0x0", "0x1"},
			{"0x2", "0x3"},
		},
		MinimumConfirmations: &minConfs,
	}

	result, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}

	if string(result) != expected {
		t.Errorf(
			"error\nwant: %s\ngot: %s",
			expected,
			string(result),
		)
	}
}

func TestSearchLogsRequestGeneratesNulls(t *testing.T) {
	expected := `[1,2,{"addresses":["0x1","0x2"]},{"topics":[null,"0x3"]},1]`
	minConfs := uint(1)
	request := &SearchLogsRequest{
		FromBlock: big.NewInt(1),
		ToBlock:   big.NewInt(2),
		Addresses: []string{"0x1", "0x2"},
		Topics: []SearchLogsTopic{
			{},
			{"0x3"},
		},
		MinimumConfirmations: &minConfs,
	}

	result, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}

	if string(result) != expected {
		t.Errorf(
			"error\nwant: %s\ngot: %s",
			expected,
			string(result),
		)
	}
}

func TestSearchLogsRequestFiltersTopicsIfOnlyOneNull(t *testing.T) {
	expected := `[1,2,{"addresses":["0x1","0x2"]},null,1]`
	minConfs := uint(1)
	request := &SearchLogsRequest{
		FromBlock: big.NewInt(1),
		ToBlock:   big.NewInt(2),
		Addresses: []string{"0x1", "0x2"},
		Topics: []SearchLogsTopic{
			{"0x3", "0x4"},
		},
		MinimumConfirmations: &minConfs,
	}

	result, err := json.Marshal(request)
	if err != nil {
		t.Fatal(err)
	}

	if string(result) != expected {
		t.Errorf(
			"error\nwant: %s\ngot: %s",
			expected,
			string(result),
		)
	}
}
