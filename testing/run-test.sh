# Initialize git repo for sparsecheckout
git init
# Add the desired repo to origin
git remote add -f origin https://github.com/OpenZeppelin/openzeppelin-contracts.git
# Configure sparsecheckout
git config core.sparsecheckout true
echo "test/*" >> .git/info/sparse-checkout
echo "contracts/*" >> .git/info/sparse-checkout
# Pull the desired subdirectories
git pull origin master
# Install dependencies
yarn install
# Remove unnecessary files and directories
rm -r test/GSN
rm test/setup.js
# Remove the git repo in order to keep track of changes outside of test/ and contracs/
rm -rf .git
# Run the tests
truffle test


