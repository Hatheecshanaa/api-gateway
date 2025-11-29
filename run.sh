#!/usr/bin/env bash
set -euo pipefail

CONFIG=${CONFIG:-config.yaml}
JWT_SECRET=${JWT_SECRET:-"VerySecretJwtKeyForLocalDevelopmentChangeThis"}

export JWT_SECRET

# Build
cd "$(dirname "$0")" || exit 1

if [ ! -f ./apigateway ]; then
  echo "Building gateway..."
  go build -o apigateway .
fi

# Run
./apigateway -config "$CONFIG"
