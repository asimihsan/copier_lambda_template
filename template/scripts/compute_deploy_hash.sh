#!/bin/bash

set -euo pipefail

SCRIPT_DIR=$(dirname "$(realpath "$0")")
ROOT_DIR=$(realpath "$SCRIPT_DIR/..")
pushd "$ROOT_DIR" > /dev/null
trap 'popd > /dev/null' EXIT

# Get the image hash using the docker inspect command
FULL_IMAGE_HASH=$(docker inspect --format='{{.Id}}' "$1":latest)
# Remove the "sha256:" prefix if it exists
IMAGE_HASH=${FULL_IMAGE_HASH#sha256:}

# Compute the hash of the sam/template.yaml file
TEMPLATE_HASH=$(sha256sum "sam/template.yaml" | awk '{print $1}')

# Combine both hashes to compute a final hash
COMBINED_HASH=$(echo "$IMAGE_HASH $TEMPLATE_HASH" | sha256sum | awk '{print $1}')

# Output the combined hash in JSON format for Terraform external data source
echo "{\"hash\": \"$COMBINED_HASH\"}"
