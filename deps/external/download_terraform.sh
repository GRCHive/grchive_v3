#!/bin/bash
MAJOR_VERSION=0
MINOR_VERSION=13
PATCH_VERSION=0
FULL_VERSION_NAME=${MAJOR_VERSION}.${MINOR_VERSION}.${PATCH_VERSION}

curl -o terraform.zip https://releases.hashicorp.com/terraform/${FULL_VERSION_NAME}/terraform_${FULL_VERSION_NAME}_linux_amd64.zip

mkdir -p terraform
unzip terraform.zip -d terraform
rm terraform.zip
