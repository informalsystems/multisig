#!/bin/bash

gaiad tx bank send \
    cosmos1mmq6knly90q9pqfx36srl3gnls3ahdm7s93g2q \
    cosmos18eq7nwggdmpxr0d4a23rmdued6jwz5mzwn8vjz \
    1uatom \
    --gas=200000 \
    --fees=1uatom \
    --chain-id=testhub \
    --generate-only > unsignedTx.json

cd user1
multisig tx push ../unsignedTx.json cosmos test
multisig sign cosmos test --from test_key_1
cd ../user2
multisig sign cosmos test --from test_key_2
multisig broadcast cosmos test

sleep 7

gaiad query bank balances cosmos18eq7nwggdmpxr0d4a23rmdued6jwz5mzwn8vjz