#!/bin/sh
HOST="${HOST-localhost}"
PORT="${PORT-4000}"
RUNFILES=$(cd ${JAVA_RUNFILES-$0.runfiles}/%{workspace_name} && pwd -P)
SOURCE_DIR="$RUNFILES/%{source_dir}"

TDIR=$(mktemp -d)
RDIR=$(mktemp -d)
trap "rm -fr $RDIR $TDIR" EXIT
(cd $RDIR && \
jekyll serve --disable-disk-cache --host "$HOST" --port "$PORT" -s "$SOURCE_DIR" -d "$TDIR")
