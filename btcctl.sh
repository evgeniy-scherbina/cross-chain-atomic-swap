#!/usr/bin/env bash

btcctl \
    --simnet \
    --rpcserver=127.0.0.1:18554 \
    --rpcuser=devuser \
    --rpcpass=devpass \
    --rpccert=data/rpc.cert \
    -C=data \
    $@