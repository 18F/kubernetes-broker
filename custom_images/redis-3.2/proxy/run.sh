#!/bin/bash

set -eux

mkdir -p /usr/local/bin
cp /redis-sentinel-proxy /usr/local/bin/redis-sentinel-proxy

/usr/local/bin/redis-sentinel-proxy \
      -listen $LISTEN_ADDRESS \
      -sentinel $SENTINEL_ADDRESS \
      -master $REDIS_MASTER_NAME
