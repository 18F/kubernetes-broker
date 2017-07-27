#!/bin/bash

ulimit -n unlimited
ulimit -l unlimited

exec /docker-entrypoint.sh "$@"
