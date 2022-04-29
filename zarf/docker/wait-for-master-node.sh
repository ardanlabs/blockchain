#!/bin/sh
# wait-for-masster-node.sh

set -e
  
until curl blockchain-node-1:9080/v1/node/status >> /dev/null; do
  >&2 echo "Master node is unavailable - sleeping"
  sleep 1
done
  
>&2 echo "Master node is up - executing command"
exec "$@"
