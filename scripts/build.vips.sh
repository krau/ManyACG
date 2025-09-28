#!/bin/bash
set -e

builtAt="$(date +'%F %T %z')"
gitCommit=$(git log --pretty=format:"%h" -1)
version=$(git describe --abbrev=0 --tags)

versionFlags="-w -s \
-X 'github.com/krau/ManyACG/internal/common.BuildTime=$builtAt' \
-X 'github.com/krau/ManyACG/internal/common.Commit=$gitCommit' \
-X 'github.com/krau/ManyACG/internal/common.Version=$version'"

vipsFlags=$(pkg-config --static --libs vips)

# nodynamic tag is for https://github.com/gen2brain/avif
CGO_ENABLED=1 go build \
    -tags nodynamic,netgo \
    -ldflags "$versionFlags -linkmode external -extldflags \"-static $vipsFlags\"" \
    -o manyacg
