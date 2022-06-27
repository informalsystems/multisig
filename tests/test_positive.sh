#!/bin/bash

test_addr_1=$(gaiad keys show test_key_1 --keyring-backend test --output json | jq -r ".address")
test_addr_2=$(gaiad keys show test_key_2 --keyring-backend test --output json | jq -r ".address")
multisig_addr_1=$(gaiad keys show multisig_test --keyring-backend test --output json | jq -r ".address")

gaiad tx bank send \
    "$multisig_addr_1" \
    "$test_addr_1" \
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

gaiad query bank balances "$test_addr_1"