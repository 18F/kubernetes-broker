#!/bin/bash

namespace="$(< /var/run/secrets/kubernetes.io/serviceaccount/namespace)"
readonly namespace
readonly service_domain="_$SERVICE_PORT._tcp.$SERVICE.$namespace.svc.cluster.local"
readonly sentinel_domain="_$SENTINEL_PORT._tcp.$SENTINEL.$namespace.svc.cluster.local"

redis_info () {
  set +e
  timeout 10 redis-cli -h "$1" -a "$REDIS_PASSWORD" info replication
  set -e
}

redis_info_role () {
  echo "$1" | grep -e '^role:' | cut -d':' -f2 | tr -d '[:space:]'
}

domain_ip () {
  /opt/bin/dig-a "$1" | head -1 | awk '{print $NF}'
}

server_domains () {
  /opt/bin/dig-srv "$1" | awk '{print $NF}' | sed 's/\.$//g'
}

reset_sentinel () {
  set +e
  timeout 10 redis-cli -h "$1" -p 26379 sentinel reset primaryserver
  set -e
}

# At the end of the (succeeded) script, resetting all sentinels is necessary.
# This updates the list of supervised slaves.
# If this task is omitted, the number of "supervised" slaves continues to
# increase because sentinels are unable to recognize the recovered slave
# is the same slave as the dead one.
# Kubernetes may change Pod's IP address on restart.
reset_all_sentinels () {
  local servers
  servers="$(server_domains "$sentinel_domain")"
  readonly servers
  local s
  >&2 echo "Resetting all sentinels: $servers"
  for s in $servers; do
    local s_ip
    s_ip="$(domain_ip "$s")"

    if [ -z "$s_ip" ]; then
      >&2 echo "Failed to resolve: $s"
      continue
    fi

    # Ignoring failed sentinels are allowed, since most of the sentinels are
    # expected to be alive.
    reset_sentinel "$s_ip"
    sleep 33
  done
}

