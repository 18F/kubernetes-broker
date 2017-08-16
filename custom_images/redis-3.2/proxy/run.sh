#!/bin/bash

set -eux

mkdir -p /opt/bin
cp /k8s-redis-sentinel-proxy /redis-sentinel-proxy /opt/bin
chmod -R +x /opt/bin
