package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/fdddf/openproxy/common"
	"github.com/fdddf/openproxy/internal/models"
	"github.com/fdddf/openproxy/pkg"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"gorm.io/gorm"
)

// ChatServiceImpl implements the ChatService interface
type ChatServiceImpl struct {
	configService   ConfigService
	providerService ProviderService
	dbService       DatabaseService
	requestLogger   RequestLogger
}

// NewChatService creates a new ChatService instance
func NewChatService(configService ConfigService, providerService ProviderService, dbService DatabaseService, requestLogger RequestLogger) ChatService {
	return &ChatServiceImpl{
		configService:   configService,
		providerService: providerService,
		dbService:       dbService,
		requestLogger:   requestLogger,
	}
}

// HandleResponses handles the responses request
func (c *ChatServiceImpl) HandleResponses(ctx *fiber.Ctx) error {
	log.Debugf("received responses request: %s", ctx.Body())

	return c.HandleChatCompletion(ctx)
}

// HandleChatCompletion handles the chat completion request
func (c *ChatServiceImpl) HandleChatCompletion(ctx *fiber.Ctx) error {
	log.Debugf("received chat completion request: %s", ctx.Body())
	requestStartTime := time.Now()
	rawRequestBody := ctx.Body()

	var req common.AIRequest
	if err := ctx.BodyParser(&req); err != nil {
		log.Errorf("parse body failed: %s, err: %s", ctx.Body(), err)
		return common.HandleError(ctx, fiber.StatusBadRequest, fmt.Sprintf("request body parse failed: %s", err))
	}

	provider, mapping, err := c.FindProvider(ctx, req.Model)
	if err != nil {
		log.Errorf("find provider failed for model %s: %s", req.Model, err)
		return common.HandleError(ctx, fiber.StatusInternalServerError, fmt.Sprintf("find provider failed: %s", err))
	}

	raw, err := provider.PrepareRequestBody(ctx, req)
	log.Debugf("prepared to request body: %s", raw)
	if err != nil {
		log.Errorf("prepare request body failed: %s", err)
		return common.HandleError(ctx, fiber.StatusInternalServerError, fmt.Sprintf("prepare request body failed: %s", err))
	}
	targetURL := provider.GetChatCompletionURL(ctx)

	reqBody := bytes.NewBuffer(raw)
	httpReq, err := http.NewRequest(provider.RequestMethod(), targetURL, reqBody)
	if err != nil {
		log.Errorf("create http request failed: %s", err)
		return common.HandleError(ctx, fiber.StatusInternalServerError, fmt.Sprintf("create http request failed: %s", err))
	}
	copyForwardableHeaders(ctx, httpReq)

	resp, err := provider.SendRequest(httpReq)
	if err != nil {
		responseTime := int(time.Since(requestStartTime).Milliseconds())
		log.Errorf("send http request failed to %s: %s", targetURL, err)
		c.recordRequestLog(ctx, rawRequestBody, []byte(err.Error()), nil, fiber.StatusBadGateway, &mapping.ProviderID, &mapping.ModelID, responseTime, req.Model)
		return common.HandleError(ctx, fiber.StatusBadGateway, fmt.Sprintf("send http request failed: %s", err))
	}
	log.Debugf("received response code: %d from url %s", resp.StatusCode, targetURL)

	if req.Stream {
		responseTime := int(time.Since(requestStartTime).Milliseconds())
		c.recordRequestLog(ctx, rawRequestBody, []byte("[streaming response]"), resp.Header, resp.StatusCode, &mapping.ProviderID, &mapping.ModelID, responseTime, req.Model)
		return provider.StreamResponse(ctx, resp)
	}

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		responseTime := int(time.Since(requestStartTime).Milliseconds())
		log.Errorf("read http response body failed: %s", err)
		c.recordRequestLog(ctx, rawRequestBody, []byte(err.Error()), resp.Header, fiber.StatusInternalServerError, &mapping.ProviderID, &mapping.ModelID, responseTime, req.Model)
		return common.HandleError(ctx, fiber.StatusInternalServerError, fmt.Sprintf("read http response failed: %s", err))
	}
	resp.Body.Close()
	resp.Body = io.NopCloser(bytes.NewBuffer(respBody))

	responseTime := int(time.Since(requestStartTime).Milliseconds())
	c.recordRequestLog(ctx, rawRequestBody, respBody, resp.Header, resp.StatusCode, &mapping.ProviderID, &mapping.ModelID, responseTime, req.Model)

	return provider.Response(ctx, resp)
}

// HandleModels handles the models list request
func (c *ChatServiceImpl) HandleModels(ctx *fiber.Ctx) error {
	var models []common.ModelInfo

	mappings, err := c.loadModelMappings()
	if err != nil {
		log.Errorf("load model mappings failed: %s", err)
		return common.HandleError(ctx, fiber.StatusInternalServerError, fmt.Sprintf("Failed to load models: %s", err))
	}

	for _, mapping := range mappings {
		models = append(models, common.ModelInfo{
			ID:      fmt.Sprintf("%s.%s", mapping.ProviderName, mapping.Name),
			Object:  "model",
			Created: 1672534800,
			OwnedBy: mapping.ProviderName,
		})
	}

	return ctx.JSON(common.ListResponse{
		Object: "list",
		Data:   models,
	})
}

// FindProvider finds the appropriate provider based on provider name and model
func (c *ChatServiceImpl) FindProvider(ctx *fiber.Ctx, model string) (common.AIProvider, *modelMapping, error) {
	providerName := ctx.Query("provider", c.configService.GetConfig().Proxy.DefaultProvider)

	// The proxy-key middleware stores the resolved caller; its absence means the
	// request reached here without a valid API key.
	if _, ok := ctx.Locals("user_id").(uint); !ok {
		return nil, nil, fmt.Errorf("API key invalid")
	}

	// check model contains dot
	if strings.Contains(model, ".") {
		t, m, result := strings.Cut(model, ".")
		if result {
			providerName, model = t, m
		}
	}

	mapping, err := c.getModelMapping(providerName, model)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return nil, nil, fmt.Errorf("no provider found or provider is disabled")
		}
		return nil, nil, fmt.Errorf("load model mapping: %w", err)
	}

	providerCfg := &common.ProviderConfig{
		Type:         mapping.Platform,
		ApiKey:       mapping.ApiKey,
		BaseURL:      mapping.BaseURL,
		ProxyURL:     mapping.ProxyURL,
		AccessToken:  mapping.AccessToken,
		RefreshToken: mapping.RefreshToken,
		AccountID:    mapping.AccountID,
	}

	provider, err := c.providerService.GetModelProvider(providerCfg, mapping.RealModel)
	if err != nil {
		return nil, nil, err
	}

	return provider, mapping, nil
}

type modelMapping struct {
	ModelID      uint   `gorm:"column:model_id"`
	Name         string `gorm:"column:model_name"`
	RealModel    string `gorm:"column:real_model"`
	Platform     string `gorm:"column:platform"`
	ProviderName string `gorm:"column:provider_name"`
	ApiKey       string `gorm:"column:api_key"`
	BaseURL      string `gorm:"column:base_url"`
	ProxyURL     string `gorm:"column:proxy_url"`
	AccessToken  string `gorm:"column:access_token"`
	RefreshToken string `gorm:"column:refresh_token"`
	AccountID    string `gorm:"column:account_id"`
	ProviderID   uint   `gorm:"column:provider_id"`
}

func (c *ChatServiceImpl) loadModelMappings() ([]modelMapping, error) {
	dao := c.dbService.GetDAO()
	var mappings []modelMapping
	if err := dao.Model.
		Select(
			dao.Model.ID.As("model_id"),
			dao.Model.Name.As("model_name"),
			dao.Model.RealModel,
			dao.Provider.ID.As("provider_id"),
			dao.Provider.Name.As("provider_name"),
			dao.Provider.Platform,
			dao.Provider.ApiKey,
			dao.Provider.BaseURL,
			dao.Provider.ProxyURL,
			dao.Provider.AccessToken,
			dao.Provider.RefreshToken,
			dao.Provider.AccountID,
		).
		Join(dao.Provider, dao.Provider.ID.EqCol(dao.Model.ProviderID)).
		Where(dao.Model.DeletedAt.IsNull()).
		Where(dao.Model.IsActive.Is(true)).
		Where(dao.Provider.DeletedAt.IsNull()).
		Where(dao.Provider.IsActive.Is(true)).
		Scan(&mappings); err != nil {
		return nil, err
	}

	return mappings, nil
}

func (c *ChatServiceImpl) getModelMapping(providerName, modelName string) (*modelMapping, error) {
	// Use provider name to uniquely identify providers instead of platform
	// This fixes the issue where multiple providers with the same platform were not being found correctly

	dao := c.dbService.GetDAO()
	var mapping modelMapping
	if err := dao.Model.
		Select(
			dao.Model.ID.As("model_id"),
			dao.Model.Name.As("model_name"),
			dao.Model.RealModel,
			dao.Provider.ID.As("provider_id"),
			dao.Provider.Name.As("provider_name"),
			dao.Provider.Platform,
			dao.Provider.ApiKey,
			dao.Provider.BaseURL,
			dao.Provider.ProxyURL,
			dao.Provider.AccessToken,
			dao.Provider.RefreshToken,
			dao.Provider.AccountID,
		).
		Join(dao.Provider, dao.Provider.ID.EqCol(dao.Model.ProviderID)).
		Where(dao.Model.Name.Eq(modelName)).
		Where(dao.Provider.Name.Eq(providerName)).
		Where(dao.Model.DeletedAt.IsNull()).
		Where(dao.Model.IsActive.Is(true)).
		Where(dao.Provider.DeletedAt.IsNull()).
		Where(dao.Provider.IsActive.Is(true)).
		Limit(1).
		Scan(&mapping); err != nil {
		return nil, err
	}
	if mapping.ModelID == 0 {
		return nil, gorm.ErrRecordNotFound
	}
	return &mapping, nil
}

func (c *ChatServiceImpl) recordRequestLog(ctx *fiber.Ctx, reqBody []byte, respBody []byte, respHeaders http.Header, statusCode int, providerID, modelID *uint, responseTime int, model string) {
	requestHeaders := ctx.GetReqHeaders()
	if requestHeaders == nil {
		requestHeaders = map[string][]string{}
	}
	for key := range requestHeaders {
		if strings.EqualFold(key, "Authorization") {
			requestHeaders[key] = []string{"[REDACTED]"}
		}
	}

	headersJSON, err := json.Marshal(requestHeaders)
	if err != nil {
		log.Errorf("marshal request headers failed: %v", err)
		headersJSON = []byte("{}")
	}

	var respHeadersJSON []byte
	if respHeaders != nil {
		if respHeadersJSON, err = json.Marshal(respHeaders); err != nil {
			log.Errorf("marshal response headers failed: %v", err)
			respHeadersJSON = []byte("{}")
		}
	} else {
		respHeadersJSON = []byte("{}")
	}

	// Get content encoding from response headers to properly handle compressed data
	var contentEncoding string
	if respHeaders != nil {
		contentEncoding = respHeaders.Get("Content-Encoding")
		if contentEncoding == "" {
			contentEncoding = respHeaders.Get("content-encoding")
		}
	}

	// Decompress response body if it's compressed
	decodedRespBody := respBody
	if contentEncoding != "" {
		decodedData, err := pkg.DecompressData(contentEncoding, respBody)
		if err == nil {
			decodedRespBody = decodedData
		}
	}

	// Decompress request body if it's compressed (check Content-Encoding in request headers)
	var reqContentEncoding string
	if requestHeaders != nil {
		reqContentEncoding = getHeader(requestHeaders, "Content-Encoding")
	}
	decodedReqBody := reqBody
	if reqContentEncoding != "" {
		decodedData, err := pkg.DecompressData(reqContentEncoding, reqBody)
		if err == nil {
			decodedReqBody = decodedData
		}
	}

	// Extract token usage and calculate cost from response body
	promptTokens, completionTokens, totalTokens, cost := c.extractTokenUsageAndCost(decodedRespBody, model, providerID)

	reqLog := &models.Request{
		Method:           ctx.Method(),
		Path:             ctx.OriginalURL(),
		RequestHeaders:   string(headersJSON),
		RequestBody:      string(decodedReqBody),
		ResponseHeaders:  string(respHeadersJSON),
		ResponseBody:     string(decodedRespBody),
		StatusCode:       statusCode,
		ResponseTime:     responseTime,
		PromptTokens:     promptTokens,
		CompletionTokens: completionTokens,
		TotalTokens:      totalTokens,
		Cost:             cost,
		UserID:           localsUint(ctx, "user_id"),
		APIKeyID:         localsUint(ctx, "api_key_id"),
		ProviderID:       providerID,
		ModelID:          modelID,
	}

	c.requestLogger.Record(reqLog)
}

// forwardableHeaders is the allowlist of client headers proxied upstream. The
// client's own Authorization is deliberately absent: providers set their own,
// and forwarding it would leak this proxy's API key to third parties. Hop-by-hop
// and transport headers are left to net/http.
var forwardableHeaders = map[string]bool{
	"Content-Type": true,
	"Accept":       true,
	"User-Agent":   true,
	"OpenAI-Beta":  true,
}

func copyForwardableHeaders(ctx *fiber.Ctx, httpReq *http.Request) {
	ctx.Request().Header.VisitAll(func(key, value []byte) {
		name := http.CanonicalHeaderKey(string(key))
		if forwardableHeaders[name] {
			httpReq.Header.Set(name, string(value))
		}
	})
	if httpReq.Header.Get("Content-Type") == "" {
		httpReq.Header.Set("Content-Type", "application/json")
	}
}

// Helper function to extract token usage and calculate costs from response
func (c *ChatServiceImpl) extractTokenUsageAndCost(respBody []byte, model string, providerID *uint) (int, int, int, float64) {
	// Initialize default values
	promptTokens := 0
	completionTokens := 0
	totalTokens := 0
	cost := 0.0

	// Parse response body to extract token usage
	var response map[string]interface{}
	if err := json.Unmarshal(respBody, &response); err != nil {
		log.Warnf("failed to parse response body for token usage: %v", err)
		return promptTokens, completionTokens, totalTokens, cost
	}

	// Extract usage information from response
	if usage, ok := response["usage"].(map[string]interface{}); ok {
		if pt, ok := usage["prompt_tokens"].(float64); ok {
			promptTokens = int(pt)
		}
		if ct, ok := usage["completion_tokens"].(float64); ok {
			completionTokens = int(ct)
		}
		if tt, ok := usage["total_tokens"].(float64); ok {
			totalTokens = int(tt)
		}
	}

	// If total tokens weren't found, calculate from prompt and completion
	if totalTokens == 0 {
		totalTokens = promptTokens + completionTokens
	}

	// Calculate cost based on provider and model
	cost = c.calculateCost(model, promptTokens, completionTokens, providerID)

	return promptTokens, completionTokens, totalTokens, cost
}

// Calculate cost based on provider, model, and token usage
func (c *ChatServiceImpl) calculateCost(model string, promptTokens, completionTokens int, providerID *uint) float64 {
	// Get the provider details to determine pricing
	if providerID != nil {
		dao := c.dbService.GetDAO()
		provider, err := dao.Provider.Where(dao.Provider.ID.Eq(*providerID)).First()
		if err != nil {
			log.Errorf("failed to get provider for cost calculation: %v", err)
			return 0.0
		}

		// Simple cost calculation based on provider type
		// In a real implementation, you would have a more sophisticated pricing model
		var costPerMillionInput, costPerMillionOutput float64

		switch strings.ToLower(string(provider.Platform)) {
		case "openai":
			// Example pricing for OpenAI models - in a real implementation these would be configurable
			costPerMillionInput = 10.0  // $10 per million tokens for input
			costPerMillionOutput = 30.0 // $30 per million tokens for output
		case "gemini":
			costPerMillionInput = 15.0
			costPerMillionOutput = 20.0
		case "iflow":
			costPerMillionInput = 20.0
			costPerMillionOutput = 25.0
		case "codex":
			costPerMillionInput = 12.0
			costPerMillionOutput = 28.0
		default:
			costPerMillionInput = 10.0
			costPerMillionOutput = 20.0
		}

		// Calculate cost
		inputCost := float64(promptTokens) * (costPerMillionInput / 1000000.0)
		outputCost := float64(completionTokens) * (costPerMillionOutput / 1000000.0)
		return inputCost + outputCost
	}

	return 0.0
}

// Helper function to get header value from map
func getHeader(headers map[string][]string, key string) string {
	for k, v := range headers {
		if strings.EqualFold(k, key) && len(v) > 0 {
			return v[0]
		}
	}
	return ""
}

// localsUint reads an ID the auth middleware stored on the request context.
func localsUint(ctx *fiber.Ctx, key string) *uint {
	value, ok := ctx.Locals(key).(uint)
	if !ok {
		return nil
	}
	return &value
}
