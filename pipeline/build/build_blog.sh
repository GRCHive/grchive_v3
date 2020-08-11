#!/bin/bash 
set -xe

bazel run -c opt //containers/blog:latest
bazel run -c opt //containers/blog/nginx:latest
