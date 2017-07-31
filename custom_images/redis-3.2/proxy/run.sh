#!/bin/bash

set -eux

chmod +x redis-sentinel-proxy

/redis-sentinel-proxy \
      -listen $LISTEN_ADDRESS \
      -sentinel $SENTINEL_ADDRESS \
      -master $REDIS_MASTER_NAME
