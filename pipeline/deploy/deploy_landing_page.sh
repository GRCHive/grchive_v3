#!/bin/bash
DIR=`dirname $0`
. ${DIR}/../pull_env_variables.sh ${DIR}/..

gcloud compute ssh grchive-landing-page-central1-c --zone=us-central1-c --command="\
docker stop landing_page && docker rm landing_page;
docker login registry.gitlab.com --username ${GITLAB_REGISTRY_USERNAME} --password ${GITLAB_REGISTRY_TOKEN};
docker pull registry.gitlab.com/grchive/grchive-v3/landing_page:latest;
docker run --rm -d --name landing_page -p 80:80 -p 443:443 registry.gitlab.com/grchive/grchive-v3/landing_page:latest;
rm ~/.docker/config.json;
"
