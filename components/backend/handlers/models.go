package handlers

import (
	"ambient-code-backend/types"
	"net/http"

	"github.com/gin-gonic/gin"
)

// AvailableModels is the single source of truth for model definitions.
// This list is returned by the /api/models endpoint and used for validation.
var AvailableModels = []types.Model{
	{
		ID:          "claude-sonnet-4-5",
		Label:       "Claude Sonnet 4.5",
		VertexID:    "claude-sonnet-4-5@20250929",
		Enabled:     true,
		Default:     true,
		Tier:        "standard",
		Description: "Fast and intelligent, best for most tasks",
	},
	{
		ID:          "claude-opus-4-6",
		Label:       "Claude Opus 4.6",
		VertexID:    "claude-opus-4-6@default",
		Enabled:     true,
		Default:     false,
		Tier:        "premium",
		Description: "Most capable model for complex reasoning",
	},
	{
		ID:          "claude-opus-4-5",
		Label:       "Claude Opus 4.5",
		VertexID:    "claude-opus-4-5@20251101",
		Enabled:     true,
		Default:     false,
		Tier:        "premium",
		Description: "Advanced reasoning and analysis",
	},
	{
		ID:          "claude-haiku-4-5",
		Label:       "Claude Haiku 4.5",
		VertexID:    "claude-haiku-4-5@20251001",
		Enabled:     true,
		Default:     false,
		Tier:        "standard",
		Description: "Fast and cost-effective for simple tasks",
	},
}

// ListModels handles GET /api/models
// Returns the list of enabled models available for agentic sessions.
// This is a public endpoint that does not require authentication.
func ListModels(c *gin.Context) {
	// Filter to only enabled models
	var enabledModels []types.Model
	for _, model := range AvailableModels {
		if model.Enabled {
			enabledModels = append(enabledModels, model)
		}
	}

	c.JSON(http.StatusOK, types.ListModelsResponse{
		Models: enabledModels,
	})
}

// GetVertexModelID returns the Vertex AI model ID for a given model ID.
// Returns the original model ID if no mapping is found.
func GetVertexModelID(modelID string) string {
	for _, model := range AvailableModels {
		if model.ID == modelID {
			return model.VertexID
		}
	}
	return modelID
}

// IsValidModel checks if a model ID is valid and enabled.
func IsValidModel(modelID string) bool {
	for _, model := range AvailableModels {
		if model.ID == modelID && model.Enabled {
			return true
		}
	}
	return false
}

// GetDefaultModel returns the default model ID.
func GetDefaultModel() string {
	for _, model := range AvailableModels {
		if model.Default && model.Enabled {
			return model.ID
		}
	}
	// Fallback to first enabled model
	for _, model := range AvailableModels {
		if model.Enabled {
			return model.ID
		}
	}
	return ""
}
