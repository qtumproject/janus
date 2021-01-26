#Need to update script to separate openzeppelin and aave tests
#Change directory to newly created integration test folder
cd openzeppelin
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
sudo yarn install
# Remove unnecessary files and directories
rm -r test/GSN
rm test/setup.js
# Remove the git repo in order to keep track of changes outside of test/ and contracs/
rm -rf .git
#Change directories and pull a different dependency
cd ../aave/
# Initialize git repo for sparsecheckout
git init
# Add the desired repo to origin
git remote add -f origin https://github.com/aave/protocol-v2.git
# Configure sparsecheckout
git config core.sparsecheckout true
echo "test/*" >> .git/info/sparse-checkout
echo "contracts/*" >> .git/info/sparse-checkout
echo "helpers/*" >> .git/info/sparse-checkout
echo "markets/*" >> .git/info/sparse-checkout
echo "modules/*" >> .git/info/sparse-checkout
echo "specs/*" >> .git/info/sparse-checkout
echo "tasks/*" >> .git/info/sparse-checkout
# Pull the desired subdirectories
git pull origin master
# Install dependencies
sudo yarn install
# Remove the git repo in order to keep track of changes outside of test/ and contracs/
rm -rf .git
# Go back to /testing
cd ..






