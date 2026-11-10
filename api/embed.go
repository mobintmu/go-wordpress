// Package api embeds the OpenAPI contract served at /openapi.yaml.
package api

import _ "embed"

// OpenAPI is the hand-written OpenAPI 3.0.3 contract.
//
//go:embed openapi.yaml
var OpenAPI []byte
