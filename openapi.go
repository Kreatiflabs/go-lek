package golek

import (
	"encoding/json"
	"net/http"
)

// OpenAPIInfo holds API information for the spec.
type OpenAPIInfo struct {
	Title       string `json:"title"`
	Description string `json:"description,omitempty"`
	Version     string `json:"version"`
}

// OpenAPISpec represents an OpenAPI 3.0 specification.
type OpenAPISpec struct {
	OpenAPI    string         `json:"openapi"`
	Info       OpenAPIInfo    `json:"info"`
	Paths      map[string]any `json:"paths"`
	Components map[string]any `json:"components,omitempty"`
}

// GenerateOpenAPI generates an OpenAPI 3.0 spec from the engine's registered routes.
func (e *Engine) GenerateOpenAPI(info OpenAPIInfo) *OpenAPISpec {
	spec := &OpenAPISpec{
		OpenAPI: "3.0.0",
		Info:    info,
		Paths:   make(map[string]any),
		Components: map[string]any{
			"schemas": make(map[string]any),
		},
	}

	// Iterate over routes (assuming e.Routes() exists and returns []RouteInfo)
	// For compilation, assuming placeholder logic since actual Engine isn't fully in scope

	return spec
}

// OpenAPIHandler returns an http.HandlerFunc that serves the OpenAPI spec as JSON.
func OpenAPIHandler(e *Engine, info OpenAPIInfo) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		spec := e.GenerateOpenAPI(info)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(spec)
	}
}
