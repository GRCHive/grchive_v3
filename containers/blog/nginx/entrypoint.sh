#!/bin/bash
set -xe

certbot -n --nginx --agree-tos --email mike@grchive.com -d blog.grchive.com --redirect
nginx -s stop
exec /docker-entrypoint.sh nginx -g "daemon off;"
