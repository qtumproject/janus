var Adoption = artifacts.require("Adoption");

module.exports = async function(deployer) {
  await deployer.deploy(Adoption);
};