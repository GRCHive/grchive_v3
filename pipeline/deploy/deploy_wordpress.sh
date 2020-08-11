#!/bin/bash

DIR=`dirname $0`
. ${DIR}/../pull_env_variables.sh ${DIR}/..

gcloud compute ssh grchive-wordpress-central1-c --zone=us-central1-c --command="\
docker login registry.gitlab.com --username ${GITLAB_REGISTRY_USERNAME} --password ${GITLAB_REGISTRY_TOKEN};
docker pull registry.gitlab.com/grchive/grchive-v3/wordpress_nginx:latest;
docker pull registry.gitlab.com/grchive/grchive-v3/wordpress:latest;
sudo mkdir -p /mnt/stateful_partition/wordpress/html;
sudo chown -R `whoami` /mnt/stateful_partition/wordpress;
"

COMPOSE_FILE=$(mktemp)
envsubst < ${DIR}/../../containers/blog/docker-compose.yml > $COMPOSE_FILE

gcloud compute scp $COMPOSE_FILE grchive-wordpress-central1-c:/mnt/stateful_partition/wordpress/docker-compose.yml --zone=us-central1-c

rm $COMPOSE_FILE

gcloud compute ssh grchive-wordpress-central1-c --zone=us-central1-c --command='\
cd /mnt/stateful_partition/wordpress;
docker run -d --name wp_compose --rm -v /var/run/docker.sock:/var/run/docker.sock -v $PWD:$PWD -w=$PWD docker/compose:1.24.0 up;
'
