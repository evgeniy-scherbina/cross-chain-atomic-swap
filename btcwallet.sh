#!/usr/bin/env bash

btcwallet \
    --simnet \
    --appdata=btcwallet \
    --rpcconnect=127.0.0.1:18556 \
    --btcdusername=devuser \
    --btcdpassword=devpass \
    --rpclisten=127.0.0.1:18554 \
    --username=devuser \
    --password=devpass \
    --rpccert=data/rpc.cert \
    --rpckey=data/rpc.key \
    --cafile=data/rpc.cert \
    $@