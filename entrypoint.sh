#!/bin/sh

set -e

if [ ! -f "/data/g2g.db" ]; then
    echo "No database found, intializing gpodder2go ..."
    /gpodder2go init
    echo "... database initialized"
fi
if [ ! -f "/data/VERIFIER_SECRET_KEY" ]; then
    echo "VERIFIER_SECRET_KEY not found, intializing VERIFIER_SECRET_KEY ..."
    cat /dev/urandom  | head -c 30 | base64 > /data/VERIFIER_SECRET_KEY
    echo "... VERIFIER_SECRET_KEY initialized"
fi
if [ "$NO_AUTH" == true ]; then
    VERIFIER_SECRET_KEY="$(cat /data/VERIFIER_SECRET_KEY)" /gpodder2go serve --addr "${ADDR:-0.0.0.0:3005}" --no-auth
else
    VERIFIER_SECRET_KEY="$(cat /data/VERIFIER_SECRET_KEY)" /gpodder2go serve --addr "${ADDR:-0.0.0.0:3005}"
fi
