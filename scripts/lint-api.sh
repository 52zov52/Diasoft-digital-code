#!/usr/bin/env bash
set -e
echo "Checking OpenAPI spec syntax..."
docker run --rm -v $(pwd)/docs/api:/spec openapitools/openapi-generator-cli validate -i /spec/openapi.yaml
echo "OpenAPI validation passed."