#!/bin/bash
chown -R www-data:www-data /var/www/html

curl -o /cloud_sql_proxy https://dl.google.com/cloudsql/cloud_sql_proxy.linux.amd64
chmod +x /cloud_sql_proxy

/cloud_sql_proxy -instances=grchive:us-central1:${WORDPRESS_INSTANCE_NAME}=tcp:3306 -credential_file=/grchive-v3-sqlclient.json &
sleep 5

exec /usr/local/bin/docker-entrypoint.sh apache2-foreground
