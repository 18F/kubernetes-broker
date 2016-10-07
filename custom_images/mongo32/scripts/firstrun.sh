#!/bin/bash

# Start MongoDB service
/usr/bin/mongod --dbpath /data/db --nojournal &
while ! nc -vz localhost 27017; do sleep 1; done

# Create User
mongo admin --eval "db.createUser({ user: '$MONGO_USERNAME', pwd: '$MONGO_PASSWORD', roles: [ { role: 'dbOwner', db: '$MONGO_DBNAME' } ] });"

# Stop MongoDB service
/usr/bin/mongod --dbpath /data/db --shutdown

rm -f /.firstrun
