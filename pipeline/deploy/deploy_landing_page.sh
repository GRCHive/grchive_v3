#!/bin/bash
USERNAME=$1
PASSWORD=$2

gcloud compute ssh grchive-landing-page-central1-c --zone=us-central1-c --command="\
docker login registry.gitlab.com --username ${USERNAME} --password ${PASSWORD};
docker pull registry.gitlab.com/grchive/grchive-v3/landing_page:latest;
docker run --rm -d --name landing_page -p 80:80 -p 443:443 registry.gitlab.com/grchive/grchive-v3/landing_page:latest;
rm ~/.docker/config.json;
"
