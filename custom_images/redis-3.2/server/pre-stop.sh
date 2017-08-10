#!/bin/bash

set -eux

primary_server_ip_from_sentinel=$(redis-cli --raw -h $SENTINEL_HOST -p 26379 SENTINEL get-master-addr-by-name primaryserver | head -1)

primary_server_ip_from_self=$(redis-cli --raw -a $REDIS_PASSWORD CONFIG GET slave-announce-ip | tail -1)

if [[ $primary_server_ip_from_self == $primary_server_ip_from_sentinel ]]
then
  echo "Shutting down master, manual failover on primaryserver."
  redis-cli -h $SENTINEL_HOST -p 26379 -c "SENTINEL failover primaryserver"
fi
