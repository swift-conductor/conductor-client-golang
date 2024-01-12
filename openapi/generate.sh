#!/usr/bin/env bash

mkdir -p ./api
rm -rf ./api/*

java -jar openapi-generator-cli.jar generate \
  --input-spec ./api.json \
  --generator-name go \
  --output ./api \
  --config config.json
