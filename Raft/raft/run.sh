#!/bin/sh

set -o errexit

while true
do
    echo "Running test"
    go test -run 2B  -race > /tmp/test
    echo "Test run successfully"
    sleep 30
done