package qtum

import (
	"math/big"

	"github.com/pkg/errors"
	"github.com/qtumproject/janus/pkg/utils"
)

type Method struct {
	*Client
}

// func (m *Method) Base58AddressToHex(addr string) (string, error) {
// 	var response GetHexAddressResponse
// 	err := m.Request(MethodGetHexAddress, GetHexAddressRequest(addr), &response)
// 	if err != nil {
// 		return "", err
// 	}

// 	return string(response), nil
// }

func (m *Method) FromHexAddress(addr string) (string, error) {
	addr = utils.RemoveHexPrefix(addr)

	var response FromHexAddressResponse
	err := m.Request(MethodFromHexAddress, FromHexAddressRequest(addr), &response)
	if err != nil {
		return "", err
	}

	return string(response), nil
}

func (m *Method) SignMessage(addr string, msg string) (string, error) {
	// returns a base64 string
	var signature string
	err := m.Request("signmessage", []string{addr, msg}, &signature)
	if err != nil {
		return "", err
	}

	return signature, nil
}

func (m *Method) GetTransaction(txID string) (*GetTransactionResponse, error) {
	var (
		req = GetTransactionRequest{
			TxID: txID,
		}
		resp = new(GetTransactionResponse)
	)
	err := m.Request(MethodGetTransaction, &req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *Method) GetRawTransaction(txID string, hexEncoded bool) (*GetRawTransactionResponse, error) {
	var (
		req = GetRawTransactionRequest{
			TxID:    txID,
			Verbose: !hexEncoded,
		}
		resp = new(GetRawTransactionResponse)
	)
	err := m.Request(MethodGetRawTransaction, &req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *Method) GetTransactionReceipt(txHash string) (*GetTransactionReceiptResponse, error) {
	var resp *GetTransactionReceiptResponse
	err := m.Request(MethodGetTransactionReceipt, GetTransactionReceiptRequest(txHash), &resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *Method) DecodeRawTransaction(hex string) (*DecodedRawTransactionResponse, error) {
	var resp *DecodedRawTransactionResponse
	err := m.Request(MethodDecodeRawTransaction, DecodeRawTransactionRequest(hex), &resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *Method) GetTransactionOut(hash string, voutNumber int, mempoolIncluded bool) (*GetTransactionOutResponse, error) {
	var (
		req = GetTransactionOutRequest{
			Hash:            hash,
			VoutNumber:      voutNumber,
			MempoolIncluded: mempoolIncluded,
		}
		resp = new(GetTransactionOutResponse)
	)
	err := m.Request(MethodGetTransactionOut, req, resp)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func (m *Method) GetBlockCount() (resp *GetBlockCountResponse, err error) {
	err = m.Request(MethodGetBlockCount, nil, &resp)
	return
}

// hard coded for now as there is only the minimum gas price
func (m *Method) GetGasPrice() (*big.Int, error) {
	return big.NewInt(0x28), nil
}

// hard coded 0x1 due to the unique nature of Qtums UTXO system, might
func (m *Method) GetTransactionCount(address string, status string) (*big.Int, error) {
	// eventually might work this out to see if there's any transactions pending for an address in the mempool
	// for now just always return 1
	return big.NewInt(0x1), nil
}

func (m *Method) GetBlockHash(b *big.Int) (resp GetBlockHashResponse, err error) {
	req := GetBlockHashRequest{
		Int: b,
	}
	err = m.Request(MethodGetBlockHash, &req, &resp)
	return resp, err
}

func (m *Method) GetBlockChainInfo() (resp GetBlockChainInfoResponse, err error) {
	err = m.Request(MethodGetBlockChainInfo, nil, &resp)
	return resp, err
}

func (m *Method) GetBlockHeader(hash string) (resp *GetBlockHeaderResponse, err error) {
	req := GetBlockHeaderRequest{
		Hash: hash,
	}
	err = m.Request(MethodGetBlockHeader, &req, &resp)
	return
}

func (m *Method) GetBlock(hash string) (resp *GetBlockResponse, err error) {
	req := GetBlockRequest{
		Hash: hash,
	}
	err = m.Request(MethodGetBlock, &req, &resp)
	return
}

func (m *Method) Generate(blockNum int, maxTries *int) (resp GenerateResponse, err error) {
	if len(m.Accounts) == 0 {
		return nil, errors.New("you must specify QTUM accounts")

	}

	acc := Account{m.Accounts[0]}

	qAddress, err := acc.ToBase58Address(m.isMain)
	if err != nil {
		return nil, err
	}

	req := GenerateRequest{
		BlockNum: blockNum,
		Address:  qAddress,
		MaxTries: maxTries,
	}

	// bytes, _ := req.MarshalJSON()
	// log.Println("generatetoaddres req:", bytes)

	err = m.Request(MethodGenerateToAddress, &req, &resp)
	return
}

func (m *Method) SearchLogs(req *SearchLogsRequest) (receipts SearchLogsResponse, err error) {
	if err := m.Request(MethodSearchLogs, req, &receipts); err != nil {
		return nil, err
	}
	return
}

func (m *Method) CallContract(req *CallContractRequest) (resp *CallContractResponse, err error) {
	if err := m.Request(MethodCallContract, req, &resp); err != nil {
		return nil, err
	}
	return
}

func (m *Method) GetAccountInfo(req *GetAccountInfoRequest) (resp *GetAccountInfoResponse, err error) {
	if err := m.Request(MethodGetAccountInfo, req, &resp); err != nil {
		return nil, err
	}
	return
}

func (m *Method) ListUnspent(req *ListUnspentRequest) (resp *ListUnspentResponse, err error) {
	if err := m.Request(MethodListUnspent, req, &resp); err != nil {
		return nil, err
	}
	return
}
