// Package types defines common type definitions.
package types

// Model represents an available LLM model for agentic sessions.
type Model struct {
	ID          string `json:"id"`
	Label       string `json:"label"`
	VertexID    string `json:"vertexId"`
	Enabled     bool   `json:"enabled"`
	Default     bool   `json:"default"`
	Tier        string `json:"tier"` // "standard" | "premium"
	Description string `json:"description"`
}

// ListModelsResponse is the response structure for the models endpoint.
type ListModelsResponse struct {
	Models []Model `json:"models"`
}
