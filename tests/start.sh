#!/bin/bash

gaiad start --minimum-gas-prices 0atom &> gaia.log &
minio server /data --console-address ":9001" > minio.log &

for _ in {1..10}
do
    latest_block_hash=$(gaiad status 2>&1 | jq -r ".SyncInfo.latest_block_hash")
    if [ "$latest_block_hash" != "" ]
    then
        echo "Gaia is started"
        break
    fi
    echo "Gaia is syncing"
    sleep 1
done

/bin/bash multisig/tests/configure_minio.sh > minio_conf.log

./bats-core/bin/bats multisig/tests/test_positive.sh
