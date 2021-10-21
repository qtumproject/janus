var SimpleStore = artifacts.require("./SimpleStore.sol");

module.exports = async function(deployer) {
  await deployer.deploy(SimpleStore, ["100"]);
};
