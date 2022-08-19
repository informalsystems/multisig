#!/bin/bash

mc alias set minio http://localhost:9000 minioadmin minioadmin --api s3v4

mc admin user add minio test_access_key test_secret_key
mc admin policy set minio readwrite user=test_access_key

mc mb minio/multisig-test-bucket
