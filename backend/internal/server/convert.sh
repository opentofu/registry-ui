#!/bin/bash

# Convert OpenAPI 3.0 spec to Swagger 2.0 format
# This script converts the manually maintained OpenAPI 3.0 specification
# to Swagger 2.0 for backward compatibility.
# This should be run every time the OpenAPI spec is updated.
# It requires the `api-spec-converter` package to be installed globally.

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
OPENAPI_FILE="$SCRIPT_DIR/openapi.yml"
SWAGGER_FILE="$SCRIPT_DIR/swagger.yml"

echo "Converting OpenAPI 3.0 to Swagger 2.0..."
if [ ! -f "$OPENAPI_FILE" ]; then
    echo "Error: OpenAPI 3.0 spec not found at $OPENAPI_FILE"
    exit 1
fi
if ! command -v api-spec-converter &>/dev/null; then
    echo "Error: api-spec-converter not found. Please install it with: npm install -g api-spec-converter"
    exit 1
fi

# Convert OpenAPI 3.0 to Swagger 2.0
echo "Converting $OPENAPI_FILE to $SWAGGER_FILE..."
api-spec-converter --from=openapi_3 --to=swagger_2 --syntax=yaml "$OPENAPI_FILE" >"$SWAGGER_FILE"

if [ $? -eq 0 ]; then
    echo "✅ Successfully converted OpenAPI 3.0 to Swagger 2.0"
    echo "  - OpenAPI 3.0: openapi.yml"
    echo "  - Swagger 2.0: swagger.yml"
else
    echo "❌ Conversion failed"
    exit 1
fi
