#!/bin/bash

ulimit -n 65536
ulimit -l unlimited

exec /docker-entrypoint.sh "$@"
