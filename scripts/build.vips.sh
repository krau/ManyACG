#!/usr/bin/env bash
# Build the static manyacg binary (libvips + webp/avif, musl) inside Docker.
#   ./scripts/build.vips.sh [out-dir]            # fast: pulls prebuilt vips image
#   ./scripts/build.vips.sh --from-source [out]  # full build from source
# Output: <out-dir>/manyacg. VIPS_BASE_IMAGE overrides the prebuilt image.
set -euo pipefail

cd "$(dirname "$0")/.."

out="${1:-dist}"
if [ "$1" = "--from-source" ]; then
    dockerfile="scripts/Dockerfile.vips"
    out="${2:-dist}"
    mode="from source"
else
    dockerfile="scripts/Dockerfile.build"
    mode="prebuilt image (${VIPS_BASE_IMAGE:-acherkrau/libvips-static:8.17.2})"
fi
mkdir -p "$out"

echo "Building ($mode)..."
docker buildx build \
    --platform linux/amd64 \
    --build-arg "BUILT_AT=$(date +'%F %T %z')" \
    --build-arg "GIT_COMMIT=$(git log --pretty=format:%h -1 2>/dev/null || echo unknown)" \
    --build-arg "VERSION=$(git describe --abbrev=0 --tags 2>/dev/null || echo dev)" \
    --target exporter \
    --output "type=local,dest=${out}" \
    -f "$dockerfile" \
    .

echo "Built: ${out}/manyacg"
