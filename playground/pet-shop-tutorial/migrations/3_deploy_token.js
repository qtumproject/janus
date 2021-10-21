var QRC20 = artifacts.require("QRC20Token");

module.exports = async function(deployer) {
  await deployer.deploy(QRC20);
};