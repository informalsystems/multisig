#!/bin/bash

setup() {
    load "$HOME/bats-support/load"
    load "$HOME/bats-assert/load"

    test_addr_1=$(gaiad keys show test_key_1 --keyring-backend test --output json | jq -r ".address")
    multisig_addr_1=$(gaiad keys show multisig_test --keyring-backend test --output json | jq -r ".address")
    denom="uatom"
}

wait_till_next_block() {
    prev_latest_block_hash=$(gaiad status 2>&1 | jq -r ".SyncInfo.latest_block_hash")
    for i in {1..10}
    do
        latest_block_hash=$(gaiad status 2>&1 | jq -r ".SyncInfo.latest_block_hash")
        if [ "$latest_block_hash" != "$prev_latest_block_hash" ]
        then
            echo "Next block appeared"
            break
        fi
        echo "Waiting for next block"
        sleep 1
    done
}

@test "Positive scenario" {
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

    cd "$HOME/multisig/tests/user1"
    multisig tx push "$HOME/unsignedTx.json" cosmos test
    multisig sign cosmos test --from test_key_1

    cd "$HOME/multisig/tests/user2"
    multisig sign cosmos test --from test_key_2
    multisig broadcast cosmos test

    wait_till_next_block

    new_balance=$(gaiad query bank balances "$test_addr_1" --output json  \
        | jq -r ".balances[] | select(.denom == \"$denom\") | .amount")

    assert bash -c "(( $new_balance > $prev_balance ))"
}
