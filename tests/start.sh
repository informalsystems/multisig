#!/bin/bash

# launching gaia and minio
gaiad start --minimum-gas-prices 0atom &> gaia.log &
minio server /data --console-address ":9001" > minio.log &

# waiting for gaia to start (assuming gaia starts way more longer than minio, so no waiting for minio)
for _ in {1..10}
do
    latest_block_hash=$(gaiad status 2>&1 | jq -r ".SyncInfo.latest_block_hash")
    if [ "$latest_block_hash" != "" ]
    then
        echo "Gaia is started"
        break
    fi
    echo "Gaia is starting"
    sleep 1
done

# minio configuring is made after it start
/bin/bash multisig/tests/configure_minio.sh > minio_conf.log

# launch actual tests
./bats-core/bin/bats multisig/tests/test_scenarios.sh
