//go:build test

package handlers

import (
	"context"
	"net/http"

	test_constants "ambient-code-backend/tests/constants"
	"ambient-code-backend/tests/logger"
	"ambient-code-backend/tests/test_utils"

	"github.com/gin-gonic/gin"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	corev1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

var _ = Describe("Models Handler", Label(test_constants.LabelUnit, test_constants.LabelHandlers), func() {
	var (
		httpUtils *test_utils.HTTPTestUtils
		k8sUtils  *test_utils.K8sTestUtils
		testToken string
	)

	BeforeEach(func() {
		logger.Log("Setting up Models Handler test")

		// Use centralized K8s test setup with fake cluster
		k8sUtils = test_utils.NewK8sTestUtils(false, "test-project")
		SetupHandlerDependencies(k8sUtils)

		httpUtils = test_utils.NewHTTPTestUtils()

		// Create namespace + role and mint a valid test token for this suite
		ctx := context.Background()
		_, err := k8sUtils.K8sClient.CoreV1().Namespaces().Create(ctx, &corev1.Namespace{
			ObjectMeta: metav1.ObjectMeta{Name: "test-project"},
		}, metav1.CreateOptions{})
		if err != nil && !errors.IsAlreadyExists(err) {
			Expect(err).NotTo(HaveOccurred())
		}
		_, err = k8sUtils.CreateTestRole(ctx, "test-project", "test-full-access-role", []string{"get", "list"}, "*", "")
		Expect(err).NotTo(HaveOccurred())

		token, _, err := httpUtils.SetValidTestToken(
			k8sUtils,
			"test-project",
			[]string{"get", "list"},
			"*",
			"",
			"test-full-access-role",
		)
		Expect(err).NotTo(HaveOccurred())
		testToken = token
	})

	AfterEach(func() {
		// Clean up created namespace (best-effort)
		if k8sUtils != nil {
			_ = k8sUtils.K8sClient.CoreV1().Namespaces().Delete(context.Background(), "test-project", metav1.DeleteOptions{})
		}
	})

	Context("Authentication", func() {
		Describe("ListAvailableModels", func() {
			It("Should require authentication", func() {
				// Arrange
				restore := WithAuthCheckEnabled()
				defer restore()

				ginCtx := httpUtils.CreateTestGinContext("GET", "/api/projects/test-project/models", nil)
				ginCtx.Params = gin.Params{
					{Key: "projectName", Value: "test-project"},
				}
				// Don't set auth header

				// Act
				ListAvailableModels(ginCtx)

				// Assert
				httpUtils.AssertHTTPStatus(http.StatusUnauthorized)
				httpUtils.AssertErrorMessage("User token required")

				logger.Log("ListAvailableModels correctly requires authentication")
			})
		})
	})

	Context("Model Listing", func() {
		Describe("ListAvailableModels", func() {
			It("Should return list of available models", func() {
				// Arrange
				ginCtx := httpUtils.CreateTestGinContext("GET", "/api/projects/test-project/models", nil)
				ginCtx.Params = gin.Params{
					{Key: "projectName", Value: "test-project"},
				}
				httpUtils.SetAuthHeader(testToken)

				// Act
				ListAvailableModels(ginCtx)

				// Assert
				httpUtils.AssertHTTPStatus(http.StatusOK)

				var response map[string]interface{}
				httpUtils.GetResponseJSON(&response)
				Expect(response).To(HaveKey("models"))

				models := response["models"].([]interface{})
				Expect(len(models)).To(BeNumerically(">", 0))

				// Verify model structure
				firstModel := models[0].(map[string]interface{})
				Expect(firstModel).To(HaveKey("value"))
				Expect(firstModel).To(HaveKey("label"))

				logger.Log("ListAvailableModels returned %d models", len(models))
			})

			It("Should include default model marker", func() {
				// Arrange
				ginCtx := httpUtils.CreateTestGinContext("GET", "/api/projects/test-project/models", nil)
				ginCtx.Params = gin.Params{
					{Key: "projectName", Value: "test-project"},
				}
				httpUtils.SetAuthHeader(testToken)

				// Act
				ListAvailableModels(ginCtx)

				// Assert
				httpUtils.AssertHTTPStatus(http.StatusOK)

				var response map[string]interface{}
				httpUtils.GetResponseJSON(&response)

				models := response["models"].([]interface{})

				// Find a model with default=true
				hasDefault := false
				for _, m := range models {
					model := m.(map[string]interface{})
					if def, ok := model["default"]; ok && def.(bool) {
						hasDefault = true
						break
					}
				}
				Expect(hasDefault).To(BeTrue(), "Expected at least one model to have default=true")

				logger.Log("ListAvailableModels includes default model marker")
			})

			It("Should return expected model identifiers", func() {
				// Arrange
				ginCtx := httpUtils.CreateTestGinContext("GET", "/api/projects/test-project/models", nil)
				ginCtx.Params = gin.Params{
					{Key: "projectName", Value: "test-project"},
				}
				httpUtils.SetAuthHeader(testToken)

				// Act
				ListAvailableModels(ginCtx)

				// Assert
				httpUtils.AssertHTTPStatus(http.StatusOK)

				var response map[string]interface{}
				httpUtils.GetResponseJSON(&response)

				models := response["models"].([]interface{})

				// Extract model values
				modelValues := make([]string, 0, len(models))
				for _, m := range models {
					model := m.(map[string]interface{})
					modelValues = append(modelValues, model["value"].(string))
				}

				// Verify expected models are present (at minimum, the default models)
				Expect(modelValues).To(ContainElement("claude-sonnet-4-5"))
				Expect(modelValues).To(ContainElement("claude-opus-4-6"))

				logger.Log("ListAvailableModels returned expected model identifiers: %v", modelValues)
			})
		})
	})

	Context("Model Definition", func() {
		Describe("masterModelList", func() {
			It("Should have at least one default model", func() {
				hasDefault := false
				for _, m := range masterModelList {
					if m.Default {
						hasDefault = true
						break
					}
				}
				Expect(hasDefault).To(BeTrue(), "masterModelList should have at least one default model")

				logger.Log("masterModelList has a default model")
			})

			It("Should have unique model values", func() {
				seen := make(map[string]bool)
				for _, m := range masterModelList {
					Expect(seen[m.Value]).To(BeFalse(), "Duplicate model value: %s", m.Value)
					seen[m.Value] = true
				}

				logger.Log("All model values in masterModelList are unique")
			})

			It("Should have non-empty labels", func() {
				for _, m := range masterModelList {
					Expect(m.Label).NotTo(BeEmpty(), "Model %s should have a non-empty label", m.Value)
				}

				logger.Log("All models have non-empty labels")
			})
		})
	})
})
