#!/bin/bash

ulimit -n 65536
ulimit -l unlimited

gosu elasticsearch /usr/share/elasticsearch/bin/es-docker "$@"
