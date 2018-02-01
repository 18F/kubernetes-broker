#!/bin/bash

# Initialize first run
if [[ -e /.firstrun ]]; then
  /scripts/firstrun.sh
fi

/usr/bin/mongod --dbpath /data/db --auth $@
