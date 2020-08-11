#!/bin/bash

SCRIPTDIR=$(dirname $0)
TERRAFORM_FOLDER=$1
DIR=${SCRIPTDIR}/../

. ${DIR}/pull_env_variables.sh ${DIR}

VARFILE=$SCRIPTDIR/variables/deploy.tfvars
envsubst < $SCRIPTDIR/variables/deploy.tfvars.tmpl > $VARFILE
VARFILE=$(readlink -f $VARFILE)

cd $SCRIPTDIR/$TERRAFORM_FOLDER
terraform init
terraform apply -var-file $VARFILE
