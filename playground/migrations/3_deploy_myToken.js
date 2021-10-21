var MyToken = artifacts.require("./MyToken.sol");

module.exports = async function(deployer) {
  await deployer.deploy(MyToken, 9999999999999, {from: "0x7926223070547d2d15b2ef5e7383e541c338ffe9"});
};
