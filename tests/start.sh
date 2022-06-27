#!/bin/bash

gaiad start --minimum-gas-prices 0atom --mode validator &> gaia.log &
minio server /data --console-address ":9001" > minio.log &

/bin/bash configure_minio.sh > minio_conf.log

for i in {1..10}
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

/bin/bash test_positive.sh
