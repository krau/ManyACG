#!/bin/sh
# Download a pinned source tarball, verify its sha256, and extract it.
# Usage: fetch.sh <url> <sha256> <workdir> [output-name]
# Used by scripts/Dockerfile.vips so every dependency is reproducible.
set -eu

url="$1"
sha="$2"
dir="$3"
name="${4:-$(basename "$url")}"

mkdir -p "$dir"
cd "$dir"

curl -fL --retry 10 --retry-all-errors --retry-delay 3 --connect-timeout 30 "$url" -o "$name"
echo "$sha  $name" | sha256sum -c - >/dev/null
tar xf "$name"
rm -f "$name"
