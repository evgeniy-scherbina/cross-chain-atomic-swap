#!/usr/bin/env bash

btcd \
    --simnet \
    --datadir=data \
    --logdir=logs \
    --listen=127.0.0.1:18444 \
    --rpclisten=127.0.0.1:18556 \
    --rpcuser=devuser \
    --rpcpass=devpass \
    --rpccert=data/rpc.cert \
    --rpckey=data/rpc.key \
    --miningaddr=ShZTsTAgSQkmqZZHnU2mDKVCXP6h26Sm46
