// scripts/index.js
//Just for cotract interaction purposes
module.exports = async function main(callback) {
    try {
      
        const accounts = await web3.eth.getAccounts();
        console.log("Accounts: \n")
        console.log(accounts) ;

        const blockNumber = await web3.eth.getBlockNumber();
        console.log("Block Number:");
        console.log(blockNumber)

        const gasPrice = await web3.eth.getGasPrice();
        console.log("Gas price:");
        console.log(gasPrice);


        const account1 = accounts[0]
        const account2 = accounts[1]

        const acc1_balance = await web3.eth.getBalance(account1);
        const acc1_balanceInEth = await web3.utils.fromWei(acc1_balance);

        console.log("Address of account 1: ", account1);
        console.log("Balance of account 1: ", acc1_balanceInEth);
        
        const acc2_balance = await web3.eth.getBalance(account2);
        const acc2_balanceInEth = await web3.utils.fromWei(acc2_balance);


        console.log("Address of account 2: ", account2);
        console.log("Balance of account 2: ", acc2_balanceInEth);

        const amount = await web3.utils.toWei("0.00002");

        console.log("Ammount to be transfered: ", web3.utils.fromWei(amount));

        
        await web3.eth.sendTransaction({
            from: account1,
            to: account2,
            value: amount,
            gas: '0x6691b7',
            gasPrice: '0x64'
        })
        .on('transactionHash', async function(hash){
            console.log("Transaction hash:", hash);
            const tx = await web3.eth.getTransaction(hash);
            console.log(tx)
            callback(0)
        })

  
        callback(0);
    } catch (error) {
        console.error(error);
        callback(1);
    }
  }