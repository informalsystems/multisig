#!/bin/bash

gaiad start --minimum-gas-prices 0atom --mode validator &> gaia.log &
minio server /data --console-address ":9001" > minio.log &

sleep 5

/bin/bash configure_minio.sh > minio_conf.log

/bin/bash test_positive.sh
