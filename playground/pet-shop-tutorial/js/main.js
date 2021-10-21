import "core-js/stable"
import "regenerator-runtime/runtime"
import {providers, Contract, ethers} from "ethers"
import {QtumProvider, QtumWallet} from "qtum-ethers-wrapper"
import {utils} from "web3"
var $ = require( "jquery" );
import AdoptionArtifact from './Adoption.json'
import Pets from './pets.json'
window.$ = $;
window.jQuery = $;

let QTUMMainnet = {
  chainId: '0x22B8', // 8888
  chainName: 'QTUM Mainnet',
  rpcUrls: ['https://localhost:23889'],
  blockExplorerUrls: ['https://qtum.info/'],
  iconUrls: [
    'https://qtum.info/images/metamask_icon.svg',
    'https://qtum.info/images/metamask_icon.png',
  ],
  nativeCurrency: {
    decimals: 18,
    symbol: 'QTUM',
  },
};
let QTUMTestNet = {
  chainId: '0x22B9', // 8889
  chainName: 'QTUM Testnet',
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
};
let QTUMRegTest = {
  chainId: '0x22BA', // 8890
  chainName: 'QTUM Regtest',
  rpcUrls: ['https://localhost:23889'],
  // blockExplorerUrls: ['https://testnet.qtum.info/'],
  iconUrls: [
    'https://qtum.info/images/metamask_icon.svg',
    'https://qtum.info/images/metamask_icon.png',
  ],
  nativeCurrency: {
    decimals: 18,
    symbol: 'QTUM',
  },
};
let config = {
  "0x22B8": QTUMMainnet,
  "0x22B9": QTUMTestNet,
  "0x22BA": QTUMRegTest,
};
config[QTUMMainnet.chainId] = QTUMMainnet;
config[QTUMTestNet.chainId] = QTUMTestNet;
config[QTUMRegTest.chainId] = QTUMRegTest;

const metamask = true;
window.App = {
  web3Provider: null,
  contracts: {},
  account: "",

  init: function() {
    // Load pets.
    var petsRow = $('#petsRow');
    var petTemplate = $('#petTemplate');

    for (let i = 0; i < Pets.length; i ++) {
      petTemplate.find('.panel-title').text(Pets[i].name);
      petTemplate.find('img').attr('src', Pets[i].picture);
      petTemplate.find('.pet-breed').text(Pets[i].breed);
      petTemplate.find('.pet-age').text(Pets[i].age);
      petTemplate.find('.pet-location').text(Pets[i].location);
      petTemplate.find('.btn-adopt').attr('pets-id', Pets[i].id);

      petsRow.append(petTemplate.html());
    }

    App.login()
    if (!metamask) {
      return App.initEthers();
    }
    return App.initWeb3();
  },

  getChainId: function() {
    return (window.ethereum || {}).chainId || 8890;
  },
  isOnQtumChainId: function() {
    let chainId = this.getChainId();
    return chainId == QTUMMainnet.chainId || chainId == QTUMTestNet.chainId;
  },

  initEthers: function() {
    let qtumRpcProvider = new QtumProvider(QTUMTestNet.rpcUrls[0]);
    let privKey = "1dd19e1648a23aaf2b3d040454d2569bd7f2cd816cf1b9b430682941a98151df";
    // WIF format
    // let privKey = "cMbgxCJrTYUqgcmiC1berh5DFrtY1KeU4PXZ6NZxgenniF1mXCRk";
    let qtumWallet = new QtumWallet(privKey, qtumRpcProvider);
    
    window.qtumWallet = qtumWallet;
    App.account = qtumWallet.address
    App.web3Provider = qtumWallet;
    return App.initContract();
  },

  initWeb3: function() {
    let self = this;
    let qtumConfig = config[this.getChainId()] || QTUMRegTest;
    console.log("Adding network to Metamask", qtumConfig);
    window.ethereum.request({
      method: "wallet_addEthereumChain",
      params: [qtumConfig],
    })
      .then(() => {
        console.log("Successfully connected to qtum")
        window.ethereum.request({ method: 'eth_requestAccounts' })
          .then((accounts) => {
            console.log("Successfully logged into metamask", accounts);
            let qtumConnected = self.isOnQtumChainId();
            let currentlyQtumConnected = self.qtumConnected;
            if (accounts && accounts.length > 0) {
              App.account = accounts[0];
            }
            if (currentlyQtumConnected != qtumConnected) {
              console.log("ChainID matches QTUM, not prompting to add network to web3, already connected.");
            }
            let qtumRpcProvider = new QtumProvider(QTUMTestNet.rpcUrls[0]);
            let qtumWallet = new QtumWallet("1dd19e1648a23aaf2b3d040454d2569bd7f2cd816cf1b9b430682941a98151df", qtumRpcProvider);
            App.account = qtumWallet.address
            if (!metamask) {
              App.web3Provider = qtumWallet;
            } else {
              App.web3Provider = new providers.Web3Provider(window.ethereum);
            }
            
            return App.initContract();
          })
          .catch((e) => {
            console.log("Connecting to web3 failed", e);
          })
      })
      .catch(() => {
        console.log("Adding network failed", arguments);
      })
  },

  initContract: async function() {
    let chainId = utils.hexToNumber(this.getChainId())
    console.log("chainId", chainId)
    const artifacts = AdoptionArtifact.networks[''+chainId];
    if (!artifacts) {
      alert("Contracts are not deployed on chain " + chainId);
      return
    }
    if (!metamask) {
      App.contracts.Adoption = new Contract(artifacts.address, AdoptionArtifact.abi, App.web3Provider)
    } else {
      App.contracts.Adoption = new Contract(artifacts.address, AdoptionArtifact.abi, App.web3Provider.getSigner())
    }
    

    // Set the provider for our contract
    // App.contracts.Adoption.setProvider(App.web3Provider);

    // Use our contract to retrieve and mark the adopted pets
    await App.markAdopted();
    return App.bindEvents();
  },

  bindEvents: function() {
    $(document).on('click', '.btn-adopt', App.handleAdopt);
  },

  markAdopted: function(adopters, account) {
    var adoptionInstance;
    return new Promise((resolve, reject) => {
      let deployed = App.contracts.Adoption.deployed();
      deployed.then(function(instance) {
        adoptionInstance = instance;
        return adoptionInstance.getAdopters.call()
          .then(function(adopters) {
            console.log("Current adopters", adopters)
            for (var i = 0; i < adopters.length; i++) {
              const adopter = adopters[i];
              if (adopter !== '0x0000000000000000000000000000000000000000') {
                $('.panel-pet').eq(i).find('button').text('Adopted').attr('disabled', true);
                $('.panel-pet').eq(i).find('.pet-adopter-container').css('display', 'block');
                let adopterLabel = adopter;
                if (adopter === App.account) {
                  adopterLabel = "You"
                }
                $('.panel-pet').eq(i).find('.pet-adopter-address').text(adopterLabel);
              } else {
                $('.panel-pet').eq(i).find('.pet-adopter-container').css('display', 'none');
              }
            }
            resolve()
            console.log("Successfully marked as adopted")
          }).catch(function(err) {
            console.log(err);
            reject(err)
          });
      }).catch(function(err) {
        console.error(err)
      })
    });
  },

  handleAdopt: function(event) {
    event.preventDefault();

    var petId = parseInt($(event.target).data('id'));

    var adoptionInstance;

    App.contracts.Adoption.deployed().then(function(instance) {
      adoptionInstance = instance;

      return adoptionInstance.adopt(petId/*, {from: App.account}*/);
    }).then(function(result) {
      console.log("Successfully adopted")
      return App.markAdopted();
    }).catch(function(err) {
      console.error("Adoption failed", err)
      console.error(err.message);
    });
  },

  login: function() {
  },

  handleLogout: function() {
    localStorage.removeItem("userWalletAddress");

    App.login();
    App.markAdopted();
  }
};

$(function() {
  $(document).ready(function() {
    App.init();
  });
});
