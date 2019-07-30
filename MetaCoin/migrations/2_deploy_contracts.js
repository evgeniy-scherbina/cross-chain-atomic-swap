const ConvertLib = artifacts.require("ConvertLib");
const MetaCoin = artifacts.require("MetaCoin");
// const Empty = artifacts.require("Empty");
const FakeMetaCoin = artifacts.require("FakeMetaCoin");

module.exports = function(deployer) {
  deployer.deploy(ConvertLib);
  deployer.link(ConvertLib, MetaCoin);
  deployer.deploy(MetaCoin);
  // deployer.deploy(Empty);
  deployer.deploy(FakeMetaCoin);
};
