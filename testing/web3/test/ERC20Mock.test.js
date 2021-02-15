const { BN, constants, expectEvent, expectRevert } = require('@openzeppelin/test-helpers');
const { web3 } = require('@openzeppelin/test-helpers/src/setup');
const { expect } = require('chai');
const { ZERO_ADDRESS } = constants;

const {
  shouldBehaveLikeERC20Transfer,
  shouldBehaveLikeERC20Approve,
} = require('./ERC20.behavior');

const ERC20Mock = artifacts.require('ERC20Mock');
/**
 * Testing topics support for eth_getLogs RPC call using web3.js
 * Events that this test looks for:
 * - Transfer (Used for mint, burn, and transfer...) 
 *   (mint and burn emit a Transfer event, refer to ERC20.sol from openzeppelin for mire info)
 * - Approval (used for approve...)
 * 
 */
contract('ERC20', function (accounts) {
  const [ initialHolder, recipient] = accounts;

  const name = 'My Token';
  const symbol = 'MTKN';

  const initialSupply = new BN(100);

  const amount = new BN(50);

  beforeEach(async function () {
    this.token = await ERC20Mock.new(name, symbol, initialHolder, initialSupply);
  });

  describe('_mint', function () {
    
    
    describe('minting Event', function () {

      
      beforeEach('minting', async function () {
        const { logs } = await this.token.mint(recipient, amount);
        this.logs = logs;
      });

      it('filtering by \'Transfer\' event for mint', async function () {

        console.log("Retrieving log from minting event")

        let logs = await web3.eth.getPastLogs({
          topics: ["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"]
        })

        console.log(logs);

        
      });
    });
  });

  describe('_burn', function () {

    

    beforeEach('burning', async function () {
      const { logs } = await this.token.burn(initialHolder, amount);
      this.logs = logs;
    });

    it('filtering by \'Transfer\' event for burn', async function () {

      console.log("Retrieving log from burning event")

      let logs = await web3.eth.getPastLogs({
        topics: ["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"]
      })

      console.log(logs);

      
    });
    
  });

  describe('_transfer', function () {

    beforeEach('transfering', async function () {
      const { logs } = await this.token.transferInternal(initialHolder, recipient, initialSupply);
      this.logs = logs;
    });

    it('filtering by \'Transfer\' event for transfer', async function () {

      console.log("Retrieving log from transfer event")

      let logs = await web3.eth.getPastLogs({
        topics: ["0xddf252ad1be2c89b69c2b068fc378daa952ba7f163c4a11628f55a4df523b3ef"]
      })

      console.log(logs);

      
    });
  });

  describe('_approve', function () {

    beforeEach('approving', async function () {
      const { logs } = await this.token.approveInternal(initialHolder, recipient, initialSupply);
      this.logs = logs;
    });

    it('filtering by \'Approval\' event for approve', async function () {

      console.log("Retrieving log from approval event")

      let logs = await web3.eth.getPastLogs({
        topics: ["0x8c5be1e5ebec7d5bd14f71427d1e84f3dd0314c0f7b2291e5b200ac8c7c3b925"]
      })

      console.log(logs);

      
    });
    
  });
});