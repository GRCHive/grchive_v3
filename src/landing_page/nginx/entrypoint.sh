#!/bin/bash
set -xe

if [[ ! -v DISABLE_CERTBOT ]]; then
    certbot -n --nginx --agree-tos --email mike@grchive.com -d grchive.com -d www.grchive.com --redirect
    nginx -s stop
fi

exec /docker-entrypoint.sh nginx -g "daemon off;"
