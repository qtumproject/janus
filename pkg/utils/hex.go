package utils

import (
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/hexutil"
)

func RemoveHexPrefix(hex string) string {
	if strings.HasPrefix(hex, "0x") {
		return hex[2:]
	}
	return hex
}

func IsEthHexAddress(str string) bool {
	return strings.HasPrefix(str, "0x") || common.IsHexAddress("0x"+str)
}

func AddHexPrefix(hex string) string {
	if strings.HasPrefix(hex, "0x") {
		return hex
	}
	return "0x" + hex
}

func AddHexPrefixIfNotEmpty(hex string) string {
	if hex == "" {
		return hex
	}
	return AddHexPrefix(hex)
}

// DecodeBig decodes a hex string whether input is with 0x prefix or not.
func DecodeBig(input string) (*big.Int, error) {
	input = AddHexPrefix(input)
	return hexutil.DecodeBig(input)
}
