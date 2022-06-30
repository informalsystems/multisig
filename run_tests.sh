#!/bin/bash
set -e

docker build -f tests/Dockerfile -t multisig_test .
docker run multisig_test
