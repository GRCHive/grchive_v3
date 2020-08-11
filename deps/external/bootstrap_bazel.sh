#!/bin/bash

BAZEL_MAJOR_VERSION=3
BAZEL_MINOR_VERSION=3
BAZEL_PATCH_VERSION=0
BAZEL_VERSION_NAME=${BAZEL_MAJOR_VERSION}.${BAZEL_MINOR_VERSION}.${BAZEL_PATCH_VERSION}

if [ ! -d bazel ] && [ ! -f bazel-${BAZEL_VERSION_NAME}-dist.zip ]; then
    curl -L -O https://github.com/bazelbuild/bazel/releases/download/${BAZEL_VERSION_NAME}/bazel-${BAZEL_VERSION_NAME}-dist.zip
fi

unzip bazel-${BAZEL_VERSION_NAME}-dist.zip -d bazel
cd bazel

env EXTRA_BAZEL_ARGS="--host_javabase=@local_jdk//:jdk" bash ./compile.sh

cd ../
rm bazel-${BAZEL_VERSION_NAME}-dist.zip
