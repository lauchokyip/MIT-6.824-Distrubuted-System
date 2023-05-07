#!/bin/sh

set -o errexit

while true
do
    go test -run 2B  -race > /tmp/test
    sleep 30
done