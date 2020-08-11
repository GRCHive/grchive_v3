#!/bin/bash 
set -xe

WORDPRESS_IMAGE=registry.gitlab.com/grchive/grchive-v3/wordpress:latest
NGINX_IMAGE=registry.gitlab.com/grchive/grchive-v3/wordpress_nginx:latest

docker tag bazel/containers/blog:latest $WORDPRESS_IMAGE
docker tag bazel/containers/blog/nginx:latest $NGINX_IMAGE

docker push $WORDPRESS_IMAGE
docker push $NGINX_IMAGE
