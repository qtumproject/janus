# Simple VUE project to switch to QTUM network via Metamask

## Project setup
```
npm install
```

### Compiles and hot-reloads for development
```
npm run serve
```

### Compiles and minifies for production
```
npm run build
```

### Customize configuration
See [Configuration Reference](https://cli.vuejs.org/config/).

### wallet_addEthereumChain
```
// request account access
window.ethereum.request({ method: 'eth_requestAccounts' })
    .then(() => {
        // add chain
        window.ethereum.request({
            method: "wallet_addEthereumChain",
            params: [{
                {
                    chainId: '0x22B9',
                    chainName: 'Qtum Testnet',
                    rpcUrls: ['https://localhost:23889'],
                    blockExplorerUrls: ['https://testnet.qtum.info/'],
                    iconUrls: [
                        'https://qtum.info/images/metamask_icon.svg',
                        'https://qtum.info/images/metamask_icon.png',
                    ],
                    nativeCurrency: {
                        decimals: 18,
                        symbol: 'QTUM',
                    },
                }
            }],
        }
    });
```

# Known issues
- Metamask requires https for `rpcUrls` so that must be enabled
  - Either directly through Janus with `--https-key ./path --https-cert ./path2` see [SSL](../README.md#ssl)
  - Through the Makefile `make docker-configure-https && make run-janus-https`
  - Or do it yourself with a proxy (eg, nginx)
