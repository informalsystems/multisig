#!/bin/bash

test_addr_1=$(gaiad keys show test_key_1 --keyring-backend test --output json | jq -r ".address")
multisig_addr_1=$(gaiad keys show multisig_test --keyring-backend test --output json | jq -r ".address")
denom="uatom"

prev_balance=$(gaiad query bank balances "$test_addr_1" --output json  \
    | jq -r ".balances[] | select(.denom == \"$denom\") | .amount")

gaiad tx bank send \
    "$multisig_addr_1" \
    "$test_addr_1" \
    "1$denom" \
    --gas=200000 \
    --fees="1$denom" \
    --chain-id=testhub \
    --generate-only > unsignedTx.json

cd user1
multisig tx push ../unsignedTx.json cosmos test
multisig sign cosmos test --from test_key_1
cd ../user2
multisig sign cosmos test --from test_key_2
multisig broadcast cosmos test

sleep 7

new_balance=$(gaiad query bank balances "$test_addr_1" --output json  \
    | jq -r ".balances[] | select(.denom == \"$denom\") | .amount")

if (( new_balance > prev_balance )); then
    echo "Success"
else
    echo "Fail"
fi
