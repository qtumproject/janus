module.exports = {
  // See <http://truffleframework.com/docs/advanced/configuration>
  // for more about customizing your Truffle configuration!
  networks: {
    development: {
      host: "127.0.0.1",
      port: 23889,
      network_id: "*",
      from: "0x7926223070547d2d15b2ef5e7383e541c338ffe9",
      gasPrice: "0x64"
    },
    ganache: {
      host: "127.0.0.1",
      port: 8545,
      network_id: "*"
    },
    testnet: {
      host: "hk1.s.qtum.org",
      port: 23889,
      network_id: "*",
      from: "0x7926223070547d2d15b2ef5e7383e541c338ffe9",
      gasPrice: "0x64"
    }
  },
  compilers: {
    solc: {
      version: "^0.8.0",
      settings: {
        optimizer: {
          enabled: true,
          runs: 1,
        },
      },
    },
  },
};
