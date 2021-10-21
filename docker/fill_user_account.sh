#!/bin/sh
repeat_until_success () {
    echo Running command - "$@"
    i=0
    until $@
    do
        echo Command failed with exit code - $?
        if [ $i -gt 10 ]; then
            echo Giving up running command - "$@"
            return
        fi
        sleep 1
        echo Retrying
        i=`expr $i + 1`
    done
    echo Command finished successfully
}

#import private keys and then prefund them
repeat_until_success qtum-cli -rpcuser=qtum -rpcpassword=testpasswd importprivkey "cMbgxCJrTYUqgcmiC1berh5DFrtY1KeU4PXZ6NZxgenniF1mXCRk" address1 # addr=qUbxboqjBRp96j3La8D1RYkyqx5uQbJPoW hdkeypath=m/88'/0'/1'
repeat_until_success qtum-cli -rpcuser=qtum -rpcpassword=testpasswd importprivkey "cRcG1jizfBzHxfwu68aMjhy78CpnzD9gJYZ5ggDbzfYD3EQfGUDZ" address2 # addr=qLn9vqbr2Gx3TsVR9QyTVB5mrMoh4x43Uf hdkeypath=m/88'/0'/2'
repeat_until_success qtum-cli -rpcuser=qtum -rpcpassword=testpasswd importprivkey "cV79qBoCSA2NDrJz8S3T7J8f3zgkGfg4ua4hRRXfhbnq5VhXkukT" address3 # addr=qTCCy8qy7pW94EApdoBjYc1vQ2w68UnXPi
repeat_until_success qtum-cli -rpcuser=qtum -rpcpassword=testpasswd importprivkey "cV93kaaV8hvNqZ711s2z9jVWLYEtwwsVpyFeEZCP6otiZgrCTiEW" address4 # addr=qWMi6ne9mDQFatRGejxdDYVUV9rQVkAFGp
repeat_until_success qtum-cli -rpcuser=qtum -rpcpassword=testpasswd importprivkey "cVPHpTvmv3UjQsZfsMRrW5RrGCyTSAZ3MWs1f8R1VeKJSYxy5uac" address5 # addr=qLcshhsRS6HKeTKRYFdpXnGVZxw96QQcfm
repeat_until_success qtum-cli -rpcuser=qtum -rpcpassword=testpasswd importprivkey "cTs5NqY4Ko9o6FESHGBDEG77qqz9me7cyYCoinHcWEiqMZgLC6XY" address6 # addr=qW28njWueNpBXYWj2KDmtFG2gbLeALeHfV

echo Finished importing accounts
echo Seeding accounts
echo Seeding qcavTSEVe31NLdXyfq925GzGp8yN5QnS6a
repeat_until_success qtum-cli -rpcuser=qtum -rpcpassword=testpasswd generatetoaddress 2 qcavTSEVe31NLdXyfq925GzGp8yN5QnS6a
qtum-cli -rpcuser=qtum -rpcpassword=testpasswd generatetoaddress 2 qa1W7VnwtJPoFDoNjxxGdDHBtsRKDpjW8c
qtum-cli -rpcuser=qtum -rpcpassword=testpasswd generatetoaddress 1000 qaofg5zZVyvPmWgGL6YdVAyRTKWd3MjZ4A
# address1
echo Seeding qUbxboqjBRp96j3La8D1RYkyqx5uQbJPoW
repeat_until_success qtum-cli -rpcuser=qtum -rpcpassword=testpasswd generatetoaddress 1000 qUbxboqjBRp96j3La8D1RYkyqx5uQbJPoW
# address2
echo Seeding qLn9vqbr2Gx3TsVR9QyTVB5mrMoh4x43Uf
repeat_until_success qtum-cli -rpcuser=qtum -rpcpassword=testpasswd generatetoaddress 1000 qLn9vqbr2Gx3TsVR9QyTVB5mrMoh4x43Uf
# address3
echo Seeding qTCCy8qy7pW94EApdoBjYc1vQ2w68UnXPi
repeat_until_success qtum-cli -rpcuser=qtum -rpcpassword=testpasswd generatetoaddress 500 qTCCy8qy7pW94EApdoBjYc1vQ2w68UnXPi
# address4
echo Seeding qWMi6ne9mDQFatRGejxdDYVUV9rQVkAFGp
repeat_until_success qtum-cli -rpcuser=qtum -rpcpassword=testpasswd generatetoaddress 250 qWMi6ne9mDQFatRGejxdDYVUV9rQVkAFGp
# address5
echo Seeding qLcshhsRS6HKeTKRYFdpXnGVZxw96QQcfm
repeat_until_success qtum-cli -rpcuser=qtum -rpcpassword=testpasswd generatetoaddress 100 qLcshhsRS6HKeTKRYFdpXnGVZxw96QQcfm
# address6
echo Seeding qW28njWueNpBXYWj2KDmtFG2gbLeALeHfV
repeat_until_success qtum-cli -rpcuser=qtum -rpcpassword=testpasswd generatetoaddress 100 qW28njWueNpBXYWj2KDmtFG2gbLeALeHfV
# playground pet shop dapp
echo Seeding 0xCca81b02942D8079A871e02BA03A3A4a8D7740d2
repeat_until_success qtum-cli -rpcuser=qtum -rpcpassword=testpasswd generatetoaddress 2 qcDWPLgdY9pTv3cKLkaMPvqjukURH3Qudy
echo Finished importing and seeding accounts
