#!/bin/bash 
set -xe

bazel run -c opt //src/landing_page:latest
