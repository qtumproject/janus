package qtum

import (
	"context"
	"encoding/json"
	"math/big"

	"github.com/pkg/errors"
	"github.com/qtumproject/janus/pkg/utils"
)

type Method struct {
	*Client
}

func (m *Method) Base58AddressToHex(addr string) (string, error) {
	var response GetHexAddressResponse
	err := m.Request(MethodGetHexAddress, GetHexAddressRequest(addr), &response)
	if err != nil {
		return "", err
	}

	return string(response), nil
}

func marshalToString(i interface{}) string {
	b, err := json.Marshal(i)
	result := ""
	if err == nil {
		result = string(b)
	}

	return result
}

func (m *Method) FromHexAddress(addr string) (string, error) {
	addr = utils.RemoveHexPrefix(addr)

	var response FromHexAddressResponse
	err := m.Request(MethodFromHexAddress, FromHexAddressRequest(addr), &response)
	if err != nil {
		if m.IsDebugEnabled() {
			m.GetDebugLogger().Log("function", "FromHexAddress", "Address", addr, "error", err)
		}
		return "", err
	}

	return string(response), nil
}

func (m *Method) SignMessage(addr string, msg string) (string, error) {
	// returns a base64 string
	var signature string
	err := m.Request("signmessage", []string{addr, msg}, &signature)
	if err != nil {
		if m.IsDebugEnabled() {
			m.GetDebugLogger().Log("function", "SignMessage", "error", err)
		}
		return "", err
	}

	if m.IsDebugEnabled() {
		m.GetDebugLogger().Log("function", "SignMessage", "addr", addr, "msg", msg, "result", signature)
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
		if m.IsDebugEnabled() {
			m.GetDebugLogger().Log("function", "GetTransaction", "Transaction ID", txID, "error", err)
		}
		return nil, err
	}
	if m.IsDebugEnabled() {
		m.GetDebugLogger().Log("function", "GetTransaction", "Transaction ID", txID, "result", marshalToString(resp))
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
		if m.IsDebugEnabled() {
			m.GetDebugLogger().Log("function", "GetRawTransaction", "Transaction ID", txID, "Hex Encoded", hexEncoded, "error", err)
		}
		return nil, err
	}
	if m.IsDebugEnabled() {
		m.GetDebugLogger().Log("function", "GetRawTransaction", "Transaction ID", txID, "Hex Encoded", hexEncoded, "result", marshalToString(resp))
	}
	return resp, nil
}

func (m *Method) GetTransactionReceipt(txHash string) (*GetTransactionReceiptResponse, error) {
	resp := new(GetTransactionReceiptResponse)
	err := m.Request(MethodGetTransactionReceipt, GetTransactionReceiptRequest(txHash), resp)
	if err != nil {
		if m.IsDebugEnabled() {
			m.GetDebugLogger().Log("function", "GetTransactionReceipt", "Transaction Hash", txHash, "error", err)
		}
		return nil, err
	}
	if m.IsDebugEnabled() {
		m.GetDebugLogger().Log("function", "GetTransactionReceipt", "Transaction Hash", txHash, "result", marshalToString(resp))
	}
	return resp, nil
}

func (m *Method) DecodeRawTransaction(hex string) (*DecodedRawTransactionResponse, error) {
	var resp *DecodedRawTransactionResponse
	err := m.Request(MethodDecodeRawTransaction, DecodeRawTransactionRequest(hex), &resp)
	if err != nil {
		if m.IsDebugEnabled() {
			m.GetDebugLogger().Log("function", "DecodeRawTransaction", "Hex", hex, "error", err)
		}
		return nil, err
	}
	if m.IsDebugEnabled() {
		m.GetDebugLogger().Log("function", "DecodeRawTransaction", "Hex", hex, "result", marshalToString(resp))
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
		if m.IsDebugEnabled() {
			m.GetDebugLogger().Log("function", "GetTransactionOut", "Hash", hash, "Vout number", voutNumber, "mempool included", mempoolIncluded, "error", err)
		}
		return nil, err
	}
	if m.IsDebugEnabled() {
		m.GetDebugLogger().Log("function", "GetTransactionOut", "Hash", hash, "Vout number", voutNumber, "mempool included", mempoolIncluded, "result", marshalToString(resp))
	}
	return resp, nil
}

func (m *Method) GetBlockCount() (resp *GetBlockCountResponse, err error) {
	err = m.Request(MethodGetBlockCount, nil, &resp)
	if m.IsDebugEnabled() {
		if err != nil {
			m.GetDebugLogger().Log("function", "GetBlockCount", "error", err)
		} else {
			m.GetDebugLogger().Log("function", "GetBlockCount", "result", resp.Int.String())
		}
	}
	return
}

func (m *Method) GetHashrate() (resp *GetHashrateResponse, err error) {
	err = m.Request(MethodGetStakingInfo, nil, &resp)
	if err != nil && m.IsDebugEnabled() {
		m.GetDebugLogger().Log("function", "GetHashrate", "error", err)
	}
	return
}

func (m *Method) GetMining() (resp *GetMiningResponse, err error) {
	err = m.Request(MethodGetStakingInfo, nil, &resp)
	if err != nil && m.IsDebugEnabled() {
		m.GetDebugLogger().Log("function", "GetMining", "error", err)
	}
	return
}

// hard coded for now as there is only the minimum gas price
func (m *Method) GetGasPrice() (*big.Int, error) {
	// 40 satoshi
	minimumGas := big.NewInt(0x28)
	m.GetDebugLogger().Log("Message", "GetGasPrice is hardcoded to "+minimumGas.String())
	return minimumGas, nil
}

// hard coded 0x1 due to the unique nature of Qtums UTXO system, might
func (m *Method) GetTransactionCount(address string, status string) (*big.Int, error) {
	// eventually might work this out to see if there's any transactions pending for an address in the mempool
	// for now just always return 1
	m.GetDebugLogger().Log("Message", "GetTransactionCount is hardcoded to one")
	return big.NewInt(0x1), nil
}

func (m *Method) GetBlockHash(b *big.Int) (resp GetBlockHashResponse, err error) {
	req := GetBlockHashRequest{
		Int: b,
	}
	err = m.Request(MethodGetBlockHash, &req, &resp)
	if err != nil && m.IsDebugEnabled() {
		m.GetDebugLogger().Log("function", "GetBlockHash", "Block", b.String(), "error", err)
	}
	return resp, err
}

func (m *Method) GetBlockChainInfo() (resp GetBlockChainInfoResponse, err error) {
	err = m.Request(MethodGetBlockChainInfo, nil, &resp)
	if err != nil && m.IsDebugEnabled() {
		m.GetDebugLogger().Log("function", "GetBlockChainInfo", "error", err)
	}
	return resp, err
}

func (m *Method) GetBlockHeader(hash string) (resp *GetBlockHeaderResponse, err error) {
	req := GetBlockHeaderRequest{
		Hash: hash,
	}
	err = m.Request(MethodGetBlockHeader, &req, &resp)
	if err != nil && m.IsDebugEnabled() {
		m.GetDebugLogger().Log("function", "GetBlockHash", "Hash", hash, "error", err)
	}
	return
}

func (m *Method) GetBlock(hash string) (resp *GetBlockResponse, err error) {
	req := GetBlockRequest{
		Hash: hash,
	}
	err = m.Request(MethodGetBlock, &req, &resp)
	if err != nil && m.IsDebugEnabled() {
		m.GetDebugLogger().Log("function", "GetBlock", "Hash", hash, "error", err)
	}
	return
}

func (m *Method) Generate(blockNum int, maxTries *int) (resp GenerateResponse, err error) {
	generateToAccount := m.GetFlagString(FLAG_GENERATE_ADDRESS_TO)

	if len(m.Accounts) == 0 && generateToAccount == nil {
		return nil, errors.New("you must specify QTUM accounts")
	}

	var qAddress string

	if generateToAccount == nil {
		acc := Account{m.Accounts[0]}

		qAddress, err = acc.ToBase58Address(m.isMain)
		if err != nil {
			if m.IsDebugEnabled() {
				m.GetDebugLogger().Log("function", "Generate", "msg", "Error getting address for account", "error", err)
			}
			return nil, err
		}
		m.GetDebugLogger().Log("function", "Generate", "msg", "generating to account 0", "account", qAddress)
	} else {
		qAddress = *generateToAccount
		m.GetDebugLogger().Log("function", "Generate", "msg", "generating to specified account", "account", qAddress)
	}

	req := GenerateRequest{
		BlockNum: blockNum,
		Address:  qAddress,
		MaxTries: maxTries,
	}

	// bytes, _ := req.MarshalJSON()
	// log.Println("generatetoaddres req:", bytes)

	err = m.Request(MethodGenerateToAddress, &req, &resp)
	if m.IsDebugEnabled() {
		if err != nil {
			m.GetDebugLogger().Log("function", "Generate", "msg", "Failed to generate block", "error", err)
		} else {
			m.GetDebugLogger().Log("function", "Generate", "msg", "Successfully generated block")
		}
	}
	return
}

/**
 * Note that QTUM searchlogs api returns all logs in a transaction receipt if any log matches a topic
 * While Ethereum behaves differently and will only return logs where topics match
 */
func (m *Method) SearchLogs(req *SearchLogsRequest) (receipts SearchLogsResponse, err error) {
	if err := m.Request(MethodSearchLogs, req, &receipts); err != nil {
		if m.IsDebugEnabled() {
			m.GetDebugLogger().Log("function", "SearchLogs", "erorr", err)
		}
		return nil, err
	}
	if m.IsDebugEnabled() {
		m.GetDebugLogger().Log("function", "SearchLogs", "request", marshalToString(req), "msg", "Successfully searched logs")
	}
	return
}

func (m *Method) CallContract(req *CallContractRequest) (resp *CallContractResponse, err error) {
	if err := m.Request(MethodCallContract, req, &resp); err != nil {
		if m.IsDebugEnabled() {
			m.GetDebugLogger().Log("function", "CallContract", "error", err)
		}
		return nil, err
	}
	if m.IsDebugEnabled() {
		m.GetDebugLogger().Log("function", "CallContract", "request", marshalToString(req), "msg", "Successfully called contract")
	}
	return
}

func (m *Method) GetAccountInfo(req *GetAccountInfoRequest) (resp *GetAccountInfoResponse, err error) {
	if err := m.Request(MethodGetAccountInfo, req, &resp); err != nil {
		if m.IsDebugEnabled() {
			m.GetDebugLogger().Log("function", "GetAccountInfo", "request", req, "error", err)
		}
		return nil, err
	}
	if m.IsDebugEnabled() {
		m.GetDebugLogger().Log("function", "GetAccountInfo", "request", req, "msg", "Successfully got account info")
	}
	return
}

func (m *Method) GetAddressUTXOs(req *GetAddressUTXOsRequest) (*GetAddressUTXOsResponse, error) {
	resp := new(GetAddressUTXOsResponse)
	if err := m.Request(MethodGetAddressUTXOs, req, resp); err != nil {
		if m.IsDebugEnabled() {
			m.GetDebugLogger().Log("function", "GetAddressUTXOs", "error", err)
		}
		return nil, err
	}
	if m.IsDebugEnabled() {
		m.GetDebugLogger().Log("function", "GetAddressUTXOs", "request", marshalToString(req), "msg", "Successfully got address UTXOs")
	}
	return resp, nil
}

func (m *Method) ListUnspent(req *ListUnspentRequest) (resp *ListUnspentResponse, err error) {
	if err := m.Request(MethodListUnspent, req, &resp); err != nil {
		if m.IsDebugEnabled() {
			m.GetDebugLogger().Log("function", "ListUnspent", "error", err)
		}
		return nil, err
	}
	if m.IsDebugEnabled() {
		m.GetDebugLogger().Log("function", "ListUnspent", "request", marshalToString(req), "msg", "Successfully list unspent")
	}
	return
}

func (m *Method) GetStorage(req *GetStorageRequest) (resp *GetStorageResponse, err error) {
	if err := m.Request(MethodGetStorage, req, &resp); err != nil {
		if m.IsDebugEnabled() {
			m.GetDebugLogger().Log("function", "GetStorage", "error", err)
		}
		return nil, err
	}
	if m.IsDebugEnabled() {
		m.GetDebugLogger().Log("function", "GetStorage", "request", marshalToString(req), "msg", "Successfully got storage")
	}
	return
}

func (m *Method) GetAddressBalance(req *GetAddressBalanceRequest) (resp *GetAddressBalanceResponse, err error) {
	if err := m.Request(MethodGetAddressBalance, req, &resp); err != nil {
		if m.IsDebugEnabled() {
			m.GetDebugLogger().Log("function", "GetAddressBalance", "error", err)
		}
		return nil, err
	}
	if m.IsDebugEnabled() {
		m.GetDebugLogger().Log("function", "GetAddressBalance", "request", marshalToString(req), "msg", "Successfully got address balance")
	}
	return
}

func (m *Method) SendRawTransaction(req *SendRawTransactionRequest) (resp *SendRawTransactionResponse, err error) {
	if err := m.Request(MethodSendRawTx, req, &resp); err != nil {
		if m.IsDebugEnabled() {
			m.GetDebugLogger().Log("function", "SendRawTransaction", "error", err)
		}
		return nil, err
	}
	if m.IsDebugEnabled() {
		m.GetDebugLogger().Log("function", "SendRawTransaction", "request", marshalToString(req), "msg", "Successfully sent raw transaction request")
	}
	return
}

func (m *Method) GetPeerInfo() (resp []GetPeerInfoResponse, err error) {
	if err := m.Request(MethodGetPeerInfo, []string{}, &resp); err != nil {
		if m.IsDebugEnabled() {
			m.GetDebugLogger().Log("function", "GetPeerInfo", "error", err)
		}
		return nil, err
	}
	if m.IsDebugEnabled() {
		m.GetDebugLogger().Log("function", "GetPeerInfo", "msg", "Successfully got peer info")
	}
	return
}

func (m *Method) GetNetworkInfo() (resp *NetworkInfoResponse, err error) {
	if err := m.Request(MethodGetNetworkInfo, []string{}, &resp); err != nil {
		if m.IsDebugEnabled() {
			m.GetDebugLogger().Log("function", "GetPeerInfo", "error", err)
		}
		return nil, err
	}
	if m.IsDebugEnabled() {
		m.GetDebugLogger().Log("function", "GetPeerInfo", "msg", "Successfully got peer info")
	}
	return
}

func (m *Method) WaitForLogs(req *WaitForLogsRequest) (resp *WaitForLogsResponse, err error) {
	return m.WaitForLogsWithContext(nil, req)
}

func (m *Method) WaitForLogsWithContext(ctx context.Context, req *WaitForLogsRequest) (resp *WaitForLogsResponse, err error) {
	if err := m.RequestWithContext(ctx, MethodWaitForLogs, req, &resp); err != nil {
		if m.IsDebugEnabled() {
			m.GetDebugLogger().Log("function", "WaitForLogs", "error", err)
		}
		return nil, err
	}
	if m.IsDebugEnabled() {
		m.GetDebugLogger().Log("function", "WaitForLogs", "request", marshalToString(req), "msg", "Successfully got waitforlogs response")
	}
	return
}
