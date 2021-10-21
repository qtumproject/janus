<template>
  <div class="hello">
    <div v-if="web3Detected">
      <b-button v-if="qtumConnected">Connected to QTUM</b-button>
      <b-button v-else-if="connected" v-on:click="connectToQtum()">Connect to QTUM</b-button>
      <b-button v-else v-on:click="connectToWeb3()">Connect</b-button>
    </div>
    <b-button v-else>No Web3 detected - Install metamask</b-button>
  </div>
</template>

<script>
let QTUMMainnet = {
  chainId: '0x71',
  chainName: 'Qtum Mainnet',
  rpcUrls: ['https://localhost:23889'],
  blockExplorerUrls: ['https://qtum.info/'],
  iconUrls: [
    'https://qtum.info/images/metamask_icon.svg',
    'https://qtum.info/images/metamask_icon.png',
  ],
};
let QTUMTestNet = {
  chainId: '0x71',
  chainName: 'Qtum Testnet',
  rpcUrls: ['https://localhost:23889'],
  blockExplorerUrls: ['https://testnet.qtum.info/'],
  iconUrls: [
    'https://qtum.info/images/metamask_icon.svg',
    'https://qtum.info/images/metamask_icon.png',
  ],
};
let config = {
  "0x1": QTUMMainnet,
  // ETH Ropsten
  "0x3": QTUMTestNet,
  // ETH Rinkby
  "0x4": QTUMTestNet,
  // ETH Görli
  "0x5": QTUMTestNet,
  // ETH Kovan
  "0x71": QTUMTestNet,
};
config[QTUMMainnet.chainId] = QTUMMainnet;
config[QTUMTestNet.chainId] = QTUMTestNet;

export default {
  name: 'Web3Button',
  props: {
    msg: String,
    connected: Boolean,
    qtumConnected: Boolean,
  },
  computed: {
    web3Detected: function() {
      return !!this.Web3;
    },
  },
  methods: {
    getChainId: function() {
      return window.ethereum.chainId;
    },
    isOnQtumChainId: function() {
      let chainId = this.getChainId();
      return chainId == QTUMMainnet.chainId || chainId == QTUMTestNet.chainId;
    },
    connectToWeb3: function(){
      if (this.connected) {
        return;
      }
      let self = this;
      window.ethereum.request({ method: 'eth_requestAccounts' })
        .then(() => {
          console.log("Emitting web3Connected event");
          let qtumConnected = self.isOnQtumChainId();
          let currentlyQtumConnected = self.qtumConnected;
          self.$emit("web3Connected", true);
          if (currentlyQtumConnected != qtumConnected) {
            console.log("ChainID matches QTUM, not prompting to add network to web3, already connected.");
            self.$emit("qtumConnected", true);
          }
        })
        .catch(() => {
          console.log("Connecting to web3 failed", arguments);
        })
    },
    connectToQtum: function() {
      console.log("Connecting to Qtum, current chainID is", this.getChainId());

      let self = this;
      let qtumConfig = config[this.getChainId()] || QTUMTestNet;
      console.log("Adding network to Metamask", qtumConfig);
      window.ethereum.request({
        method: "wallet_addEthereumChain",
        params: [qtumConfig],
      })
        .then(() => {
          self.$emit("qtumConnected", true);
        })
        .catch(() => {
          console.log("Adding network failed", arguments);
        })
    },
  }
}
</script>

<!-- Add "scoped" attribute to limit CSS to this component only -->
<style scoped>
</style>
