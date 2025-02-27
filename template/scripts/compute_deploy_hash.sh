#!/bin/bash

set -euo pipefail

SCRIPT_DIR=$(dirname "$(realpath "$0")")
ROOT_DIR=$(realpath "$SCRIPT_DIR/..")
pushd "$ROOT_DIR" > /dev/null
trap 'popd > /dev/null' EXIT

# Get the image hash using the just command
IMAGE_HASH=$(docker inspect --format='{{.Id}}' "$1":latest)

# Output the hash in JSON format for Terraform external data source
echo "{\"hash\": \"$IMAGE_HASH\"}"
