#!/bin/bash

set -eux

apt-get update && apt-get install -y netcat
mkdir -p /opt/bin
cp /k8s-redis-sentinel-proxy /redis-sentinel-proxy /opt/bin
chmod -R +x /opt/bin
