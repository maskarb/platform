package handlers

import (
	"ambient-code-backend/types"
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"k8s.io/apimachinery/pkg/api/errors"
	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// DefaultModels is the fallback when the ambient-models ConfigMap is missing or invalid.
var DefaultModels = []types.Model{
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

// loadModelsFromConfigMap reads the ambient-models ConfigMap.
// Returns DefaultModels if the ConfigMap is missing or unparseable.
func loadModelsFromConfigMap() []types.Model {
	if K8sClient == nil {
		return DefaultModels
	}
	cm, err := K8sClient.CoreV1().ConfigMaps(Namespace).Get(
		context.Background(), "ambient-models", v1.GetOptions{})
	if err != nil {
		if !errors.IsNotFound(err) {
			log.Printf("Failed to read ambient-models ConfigMap: %v", err)
		}
		return DefaultModels
	}
	raw, ok := cm.Data["models.json"]
	if !ok {
		log.Printf("ambient-models ConfigMap missing 'models.json' key")
		return DefaultModels
	}
	var models []types.Model
	if err := json.Unmarshal([]byte(raw), &models); err != nil {
		log.Printf("Failed to parse models.json from ConfigMap: %v", err)
		return DefaultModels
	}
	if len(models) == 0 {
		return DefaultModels
	}
	return models
}

// ListModels handles GET /api/models
// Returns the list of enabled models available for agentic sessions.
// This is a public endpoint that does not require authentication.
func ListModels(c *gin.Context) {
	models := loadModelsFromConfigMap()
	// Filter to only enabled models
	var enabledModels []types.Model
	for _, model := range models {
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
	for _, model := range loadModelsFromConfigMap() {
		if model.ID == modelID {
			return model.VertexID
		}
	}
	return modelID
}

// IsValidModel checks if a model ID is valid and enabled.
func IsValidModel(modelID string) bool {
	for _, model := range loadModelsFromConfigMap() {
		if model.ID == modelID && model.Enabled {
			return true
		}
	}
	return false
}

// GetDefaultModel returns the default model ID.
func GetDefaultModel() string {
	models := loadModelsFromConfigMap()
	for _, model := range models {
		if model.Default && model.Enabled {
			return model.ID
		}
	}
	// Fallback to first enabled model
	for _, model := range models {
		if model.Enabled {
			return model.ID
		}
	}
	return ""
}
