// Package handlers: model listing with feature flag gating.
// Returns available LLM models filtered by feature flags.

package handlers

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// Model represents an available LLM model
type Model struct {
	Value       string `json:"value"`                 // Model identifier (e.g., "claude-sonnet-4-5")
	Label       string `json:"label"`                 // Display name (e.g., "Claude Sonnet 4.5")
	Description string `json:"description,omitempty"` // Optional description
	Default     bool   `json:"default,omitempty"`     // Is this the default model?
}

// modelDefinition includes feature flag gating info (not exposed to API)
type modelDefinition struct {
	Model
	FeatureFlag string // Optional flag name; if set, model only shown when flag is enabled
}

// masterModelList is the source of truth for available models.
// Add new models here. Set FeatureFlag to gate availability.
var masterModelList = []modelDefinition{
	{Model: Model{Value: "claude-sonnet-4-5", Label: "Claude Sonnet 4.5", Default: true}},
	{Model: Model{Value: "claude-opus-4-6", Label: "Claude Opus 4.6"}},
	{Model: Model{Value: "claude-opus-4-5", Label: "Claude Opus 4.5"}},
	{Model: Model{Value: "claude-haiku-4-5", Label: "Claude Haiku 4.5"}},
	// Example of a gated model:
	// {Model: Model{Value: "claude-next", Label: "Claude Next (Beta)"}, FeatureFlag: "model.claude-next.enabled"},
}

// ListAvailableModels handles GET /api/projects/:projectName/models
// Returns the list of models available to the user, filtered by feature flags.
func ListAvailableModels(c *gin.Context) {
	// Verify user has project access
	reqK8s, _ := GetK8sClientsForRequest(c)
	if reqK8s == nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "User token required"})
		c.Abort()
		return
	}

	// Filter models based on feature flags
	availableModels := make([]Model, 0, len(masterModelList))
	for _, m := range masterModelList {
		// If no feature flag required, include the model
		if m.FeatureFlag == "" {
			availableModels = append(availableModels, m.Model)
			continue
		}
		// Check feature flag using the Unleash Client SDK
		if FeatureEnabledForRequest(c, m.FeatureFlag) {
			availableModels = append(availableModels, m.Model)
		}
	}

	c.JSON(http.StatusOK, gin.H{"models": availableModels})
}
