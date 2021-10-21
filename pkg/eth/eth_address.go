package eth

import (
	"encoding/json"
	"strings"

	"github.com/pkg/errors"

	"github.com/ethereum/go-ethereum/common/hexutil"
)

var ErrNoHexPrefix = errors.New("Missing 0x prefix")
var ErrInvalidLength = errors.New("Invalid length")

type ETHAddress struct {
	address string
}

func NewETHAddress(address string) (ETHAddress, error) {
	var ethAddress ETHAddress
	if !strings.HasPrefix(address, `"`) {
		address = `"` + address
	}

	if !strings.HasSuffix(address, `"`) {
		address = address + `"`
	}

	err := json.Unmarshal([]byte(address), &ethAddress)
	if err != nil {
		return ethAddress, err
	}
	return ethAddress, nil
}

func (addr *ETHAddress) String() string {
	return addr.address
}

func (addr ETHAddress) MarshalJSON() ([]byte, error) {
	if err := validateAddress(addr.address); err != nil {
		return []byte{}, err
	}

	return json.Marshal(addr.address)
}

// UnmarshalJSON needs to be able to parse ETHAddress from both hex string or number
func (addr *ETHAddress) UnmarshalJSON(data []byte) (err error) {
	asString := string(data)
	if strings.HasPrefix(asString, `"`) && strings.HasSuffix(asString, `"`) {
		asString = asString[1 : len(asString)-1]
	}
	if err := validateAddress(asString); err != nil {
		return err
	}

	addr.address = asString
	return nil
}

func validateAddress(address string) error {
	if !strings.HasPrefix(address, "0x") {
		return ErrNoHexPrefix
	}

	if len(address) != 42 {
		return ErrInvalidLength
	}

	_, err := hexutil.Decode(address)
	if err != nil {
		return errors.Wrap(err, "Invalid hexadecimal")
	}

	return nil
}
