#!/bin/bash
set -e

# Get the image hash using the just command
IMAGE_HASH=$(cd ../../ && just docker-image-hash)

# Output the hash in JSON format for Terraform external data source
echo "{\"hash\": \"$IMAGE_HASH\"}"
