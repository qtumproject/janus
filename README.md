# Qtum adapter to Ethereum JSON RPC
Janus is a web3 proxy adapter that can be used as a web3 provider to interact with Qtum. It supports HTTP(s) and websockets and the current version enables self hosting of keys.

# Table of Contents

- [Requirements](#requirements)
- [Installation](#installation)
  - [SSL](#ssl)
  - [Self-signed SSL](#self-signed-ssl)
- [How to use Janus as a Web3 provider](#how-to-use-janus-as-a-web3-provider)
- [How to add Janus to Metamask](#how-to-add-janus-to-metamask)
- [Supported ETH methods](#support-eth-methods)
- [Websocket ETH methods](#websocket-eth-methods-endpoint-at-ws)
- [Janus methods](#janus-methods)
- [Try to interact with contract](#try-to-interact-with-contract)
  - [Assumption parameters](#assumption-parameters)
  - [Deploy the contract](#deploy-the-contract)
  - [Get the transaction using the hash from previous the result](#get-the-transaction-using-the-hash-from-previous-the-result)
  - [Get the transaction receipt](#get-the-transaction-receipt)
  - [Calling the set method](#calling-the-set-method)
  - [Calling the get method](#calling-the-get-method)
- [Known issues](#known-issues)


## Requirements

- Golang
- Docker
- linux commands: `make`, `curl`

## Installation

```
$ go get github.com/qtumproject/janus/...
$ cd $GOPATH/src/github.com/qtumproject/janus/playground
$ make
$ make docker-dev
$ make quick-start
```
This will build the docker image for the local version of Janus as well as spin up two containers:

-   One named `janus` running on port 23889
    
-   Another one named `qtum` running on port 3889
    

`make quick-start` will also fund the tests accounts with QTUM in order for you to start testing and developing locally. Additionally, if you need or want to make changes and or additions to Janus, but don't want to go through the hassle of rebuilding the container, you can run the following command at the project root level:
```
$ make run-janus
# For https
$ make docker-configure-https && make run-janus-https
```
Which will run the most current local version of Janus on port 23888, but without rebuilding the image or the local docker container.

Note that Janus will use the hex address for the test base58 Qtum addresses that belong the the local qtum node, for example:
  - qUbxboqjBRp96j3La8D1RYkyqx5uQbJPoW (hex 0x7926223070547d2d15b2ef5e7383e541c338ffe9 )
  - qLn9vqbr2Gx3TsVR9QyTVB5mrMoh4x43Uf (hex 0x2352be3db3177f0a07efbe6da5857615b8c9901d )

### SSL
SSL keys and certificates go inside the https folder (mounted at `/https` in the container) and use `--https-key` and `--https-cert` parameters. If the specified files do not exist, it will fall back to http.

### Self-signed SSL
To generate self-signed certificates with docker for local development the following script will generate SSL certificates and drop them into the https folder

```
$ make docker-configure-https
```

## How to use Janus as a Web3 provider

Once Janus is successfully running, all one has to do is point your desired framework to Janus in order to use it as your web3 provider. Lets say you want to use truffle for example, in this case all you have to do is go to your truffle-config.js file and add janus as a network:
```
module.exports = {
  networks: {
    janus: {
      host: "127.0.0.1",
      port: 23889,
      network_id: "*",
      gasPrice: "0x64"
    },
    ...
  },
...
}
```

## How to add Janus to Metamask

Getting Janus to work with Metamask requires two things
- [Configuring Metamask to point to Janus](metamask)
- Locally signing transactions through Metamask
  - (This is being worked on and currently is not implemented yet)

## Supported ETH methods

-   web3_clientVersion
-   web3_sha3
-   net_version
-   net_listening
-   net_peerCount
-   eth_protocolVersion
-   eth_chainId
-   eth_mining
-   eth_hashrate
-   eth_gasPrice
-   eth_accounts
-   eth_blockNumber    
-   eth_getBalance    
-   eth_getStorageAt    
-   eth_getTransactionCount    
-   eth_getCode
-   eth_sign
-   eth_signTransaction    
-   eth_sendTransaction    
-   eth_sendRawTransaction    
-   eth_call    
-   eth_estimateGas    
-   eth_getBlockByHash    
-   eth_getBlockByNumber    
-   eth_getTransactionByHash    
-   eth_getTransactionByBlockHashAndIndex    
-   eth_getTransactionByBlockNumberAndIndex    
-   eth_getTransactionReceipt    
-   eth_getUncleByBlockHashAndIndex    
-   eth_getCompilers    
-   eth_newFilter
-   eth_newBlockFilter    
-   eth_uninstallFilter    
-   eth_getFilterChanges    
-   eth_getFilterLogs    
-   eth_getLogs

## Websocket ETH methods (endpoint at /)
-   (All the above methods)
-   eth_subscribe (only 'logs' for now)
-   eth_unsubscribe

## Janus methods

-   qtum_getUTXOs

## Deploying and Interacting with a contract using RPC calls


### Assumption parameters

Assume that you have a **contract** like this:

```solidity
pragma solidity ^0.4.18;

contract SimpleStore {
  constructor(uint _value) public {
    value = _value;
  }

  function set(uint newValue) public {
    value = newValue;
  }

  function get() public constant returns (uint) {
    return value;
  }

  uint value;
}
```

so that the **bytecode** is

```
solc --optimize --bin contracts/SimpleStore.sol

======= contracts/SimpleStore.sol:SimpleStore =======
Binary:
608060405234801561001057600080fd5b506040516020806100f2833981016040525160005560bf806100336000396000f30060806040526004361060485763ffffffff7c010000000000000000000000000000000000000000000000000000000060003504166360fe47b18114604d5780636d4ce63c146064575b600080fd5b348015605857600080fd5b5060626004356088565b005b348015606f57600080fd5b506076608d565b60408051918252519081900360200190f35b600055565b600054905600a165627a7a7230582049a087087e1fc6da0b68ca259d45a2e369efcbb50e93f9b7fa3e198de6402b810029
```

**constructor parameters** is `0000000000000000000000000000000000000000000000000000000000000001`

### Deploy the contract

```
$ curl --header 'Content-Type: application/json' --data \
     '{"id":"10","jsonrpc":"2.0","method":"eth_sendTransaction","params":[{"from":"0x7926223070547d2d15b2ef5e7383e541c338ffe9","gas":"0x6691b7","gasPrice":"0x64","data":"0x608060405234801561001057600080fd5b506040516020806100f2833981016040525160005560bf806100336000396000f30060806040526004361060485763ffffffff7c010000000000000000000000000000000000000000000000000000000060003504166360fe47b18114604d5780636d4ce63c146064575b600080fd5b348015605857600080fd5b5060626004356088565b005b348015606f57600080fd5b506076608d565b60408051918252519081900360200190f35b600055565b600054905600a165627a7a7230582049a087087e1fc6da0b68ca259d45a2e369efcbb50e93f9b7fa3e198de6402b8100290000000000000000000000000000000000000000000000000000000000000001"}]}' \
     'http://localhost:23889'

{
  "jsonrpc": "2.0",
  "result": "0xa85cacc6143004139fc68808744ea6125ae984454e0ffa6072ac2f2debb0c2e6",
  "id": "10"
}
```

### Get the transaction using the hash from previous the result

```
$ curl --header 'Content-Type: application/json' --data \
     '{"id":"10","jsonrpc":"2.0","method":"eth_getTransactionByHash","params":["0x6da39dc909debf70a536bbc108e2218fd7bce23305ddc00284075df5dfccc21b"]}' \
     'localhost:23889'

{
  "jsonrpc":"2.0",
  "result": {
    "blockHash":"0x1e64595e724ea5161c0597d327072074940f519a6fb285ae60e73a4c996b47a4",
    "blockNumber":"0xc9b5",
    "transactionIndex":"0x5",
    "hash":"0xa85cacc6143004139fc68808744ea6125ae984454e0ffa6072ac2f2debb0c2e6",
    "nonce":"0x0",
    "value":"0x0",
    "input":"0x00",
    "from":"0x7926223070547d2d15b2ef5e7383e541c338ffe9",
    "to":"",
    "gas":"0x363639316237",
    "gasPrice":"0x3634"
  },
  "id":"10"
}
```

### Get the transaction receipt

```
$ curl --header 'Content-Type: application/json' --data \
     '{"id":"10","jsonrpc":"2.0","method":"eth_getTransactionReceipt","params":["0x6da39dc909debf70a536bbc108e2218fd7bce23305ddc00284075df5dfccc21b"]}' \
     'localhost:23889'

{
  "jsonrpc": "2.0",
  "result": {
    "transactionHash": "0xa85cacc6143004139fc68808744ea6125ae984454e0ffa6072ac2f2debb0c2e6",
    "transactionIndex": "0x5",
    "blockHash": "0x1e64595e724ea5161c0597d327072074940f519a6fb285ae60e73a4c996b47a4",
    "from":"0x7926223070547d2d15b2ef5e7383e541c338ffe9"
    "blockNumber": "0xc9b5",
    "cumulativeGasUsed": "0x8c235",
    "gasUsed": "0x1c071",
    "contractAddress": "0x1286595f8683ae074bc026cf0e587177b36842e2",
    "logs": [],
    "logsBloom": "0x00000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000000",
    "status": "0x1"
  },
  "id": "10"
}
```

### Calling the set method

the ABI code of set method with param '["2"]' is `60fe47b10000000000000000000000000000000000000000000000000000000000000002`

```
$ curl --header 'Content-Type: application/json' --data \
     '{"id":"10","jsonrpc":"2.0","method":"eth_sendTransaction","params":[{"from":"0x7926223070547d2d15b2ef5e7383e541c338ffe9","gas":"0x6691b7","gasPrice":"0x64","to":"0x1286595f8683ae074bc026cf0e587177b36842e2","data":"60fe47b10000000000000000000000000000000000000000000000000000000000000002"}]}' \
     'localhost:23889'

{
  "jsonrpc": "2.0",
  "result": "0x51a286c3bc68335274b9fd255e3988918a999608e305475105385f7ccf838339",
  "id": "10"
}
```

### Calling the get method

get method's ABI code is `6d4ce63c`

```
$ curl --header 'Content-Type: application/json' --data \
     '{"id":"10","jsonrpc":"2.0","method":"eth_call","params":[{"from":"0x7926223070547d2d15b2ef5e7383e541c338ffe9","gas":"0x6691b7","gasPrice":"0x64","to":"0x1286595f8683ae074bc026cf0e587177b36842e2","data":"6d4ce63c"},"latest"]}' \
     'localhost:23889'

{
  "jsonrpc": "2.0",
  "result": "0x0000000000000000000000000000000000000000000000000000000000000002",
  "id": "10"
}
```

## Known issues
- Sending coins with the creation of a contract will cause a loss of coins
  - This is a Qtum intentional deisgn decision and will not change
- On a transfer of Qtum to a Qtum address, there is no receipt generated for such a transfer
- When converting from WEI -> QTUM, precision is lost due to QTUM's smallest demonination being 1 satoshi.
  - 1 satoshi = 0.00000001 QTUM = 10000000000 wei
- QTUM's minimum gas price is 40 satoshi
  - When specifying a gas price in wei lower than that, the minimum gas price will be used (40 satoshi)
- Only 'logs' eth_subscribe type is supported at the moment
