import path from 'path';
import fs from 'fs';
import { HardhatUserConfig } from 'hardhat/types';
// @ts-ignore

require('dotenv').config();

import '@nomiclabs/hardhat-ethers';
import '@nomiclabs/hardhat-waffle';
import 'hardhat-gas-reporter';
import 'hardhat-typechain';
import '@tenderly/hardhat-tenderly';

const SKIP_LOAD = true;

// Prevent to load scripts before compilation and typechain
if (!SKIP_LOAD) {
  ['misc', 'migrations', 'dev', 'full', 'verifications', 'deployments', 'helpers'].forEach(
    (folder) => {
      const tasksPath = path.join(__dirname, 'tasks', folder);
      fs.readdirSync(tasksPath)
        .filter((pth) => pth.includes('.ts'))
        .forEach((task) => {
          require(`${tasksPath}/${task}`);
        });
    }
  );
}

require(`${path.join(__dirname, 'tasks/misc')}/set-bre.ts`);


const buidlerConfig: HardhatUserConfig = {
  solidity: {
    version: '0.6.12',
    settings: {
      optimizer: {enabled: true, runs: 1},
    },
  },  
  typechain: {
    outDir: 'types',
    target: 'ethers-v5',
  },
  mocha: {
    timeout: 0,
  },
  defaultNetwork: "development",
  networks: {
    development: {
      url: "http://127.0.0.1:23889",
      gas: "auto",
      gasPrice: "auto",
      timeout: 600000
    },
    ganache: {
      url: "http://127.0.0.1:8545",
      gas: "auto",
    },
    testnet: {
      url: "http://hk1.s.qtum.org:23889",
      from: "0x7926223070547d2d15b2ef5e7383e541c338ffe9",
      gas: "auto",
      gasPrice: "auto"
    },
  },
};

export default buidlerConfig;
