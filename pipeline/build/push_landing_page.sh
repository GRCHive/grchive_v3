#!/bin/bash 
set -xe

FULL_IMAGE=registry.gitlab.com/grchive/grchive-v3/landing_page:latest
docker tag bazel/src/landing_page:latest $FULL_IMAGE
docker push $FULL_IMAGE
