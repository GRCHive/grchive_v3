# Getting Started

## Prerequisites

This list may not be complete.

- G++ v?
- JDK8
- Docker
- Docker Compose
- curl
- tar
- Ruby (Jekyll)

## Setup

The root directory of the `grchive-v3` repository will be referenced to as `$GRCHIVE` in this document.
This document will walk you through setting up build environment and the necessary infrastructure to run the GRCHive app on your local machine.

1. Download the binaries needed for development.

    ```
    cd $GRCHIVE/deps/external
    ./download_all.sh
    ```
1. Add the following paths to your `$PATH`:

    ```
    $GRCHIVE/deps/external/bazel/output
    $GRCHIVE/deps/external/terraform
    ```
