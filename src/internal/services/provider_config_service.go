package services

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/fdddf/openproxy/common"
	"github.com/fdddf/openproxy/consts"
	"github.com/fdddf/openproxy/internal/dto"
	"github.com/fdddf/openproxy/internal/models"
	"github.com/gofiber/fiber/v2"
)

// ProviderConfigServiceImpl implements the ProviderConfigService interface
type ProviderConfigServiceImpl struct {
	dbService          DatabaseService
	codexOAuth         codexOAuthService
	antigravityOAuth   antigravityOAuthService
	healthCheckService HealthCheckService
}

// NewProviderConfigService creates a new ProviderConfigService instance
func NewProviderConfigService(dbService DatabaseService, healthCheckService HealthCheckService) ProviderConfigService {
	return &ProviderConfigServiceImpl{
		dbService:          dbService,
		codexOAuth:         newCodexOAuthService(),
		antigravityOAuth:   newAntigravityOAuthService(),
		healthCheckService: healthCheckService,
	}
}

func (p *ProviderConfigServiceImpl) HandleGetProviders(c *fiber.Ctx) error {
	page, pageSize := common.GetPaginationParams(c)
	offset := (page - 1) * pageSize

	dao := p.dbService.GetDAO()
	providers, total, err := dao.Provider.
		Order(dao.Provider.CreatedAt.Desc()).
		FindByPage(offset, pageSize)
	if err != nil {
		return common.HandleError(c, fiber.StatusInternalServerError, "Failed to get providers")
	}

	var modelsList []*models.Model
	if len(providers) > 0 {
		providerIDs := make([]uint, 0, len(providers))
		for _, provider := range providers {
			providerIDs = append(providerIDs, provider.ID)
		}

		modelsList, err = dao.Model.Where(dao.Model.ProviderID.In(providerIDs...)).Find()
		if err != nil {
			return common.HandleError(c, fiber.StatusInternalServerError, "Failed to get provider models")
		}
	}

	modelsByProvider := make(map[uint][]*models.Model)
	for _, m := range modelsList {
		modelsByProvider[m.ProviderID] = append(modelsByProvider[m.ProviderID], m)
	}

	resp := make([]dto.Provider, 0, len(providers))
	for _, provider := range providers {
		resp = append(resp, dto.NewProvider(provider, modelsByProvider[provider.ID]))
	}

	return c.JSON(common.NewPaginatedResponse(resp, total, page, pageSize))
}

func (p *ProviderConfigServiceImpl) HandleGetProviderPlatforms(c *fiber.Ctx) error {
	return c.JSON(consts.ProviderPlatformNames())
}

func (p *ProviderConfigServiceImpl) HandleCreateProvider(c *fiber.Ctx) error {
	var req struct {
		Name               string `json:"name"`
		Type               string `json:"type"`
		Key                string `json:"key"`
		BaseURL            string `json:"baseUrl"`
		ProxyURL           string `json:"proxyUrl"`
		IsActive           bool   `json:"isActive"`
		ClientID           string `json:"clientId,omitempty"`
		ClientSecret       string `json:"clientSecret,omitempty"`
		AuthURL            string `json:"authUrl,omitempty"`
		TokenURL           string `json:"tokenUrl,omitempty"`
		RedirectURL        string `json:"redirectUrl,omitempty"`
		Scopes             string `json:"scopes,omitempty"`
		AccountID          string `json:"accountId,omitempty"`
		HealthCheckEnabled bool   `json:"healthCheckEnabled"`
		HealthCheckStatus  string `json:"healthCheckStatus"`
	}

	if err := c.BodyParser(&req); err != nil {
		return common.HandleError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	// Apply platform-specific defaults and sanitization
	applyPlatformDefaults(&req)

	provider := models.Provider{
		Name:               req.Name,
		Platform:           consts.ProviderPlatform(req.Type),
		ApiKey:             req.Key,
		BaseURL:            req.BaseURL,
		ProxyURL:           req.ProxyURL,
		IsActive:           req.IsActive,
		ClientID:           req.ClientID,
		ClientSecret:       req.ClientSecret,
		AuthURL:            req.AuthURL,
		TokenURL:           req.TokenURL,
		RedirectURL:        req.RedirectURL,
		Scopes:             req.Scopes,
		AccountID:          req.AccountID,
		HealthCheckEnabled: req.HealthCheckEnabled,
		HealthCheckStatus:  req.HealthCheckStatus,
	}

	if err := p.dbService.GetDAO().Provider.Create(&provider); err != nil {
		return common.HandleError(c, fiber.StatusInternalServerError, "Failed to create provider")
	}

	return c.JSON(dto.NewProvider(&provider, nil))
}

func (p *ProviderConfigServiceImpl) HandleUpdateProvider(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return common.HandleError(c, fiber.StatusBadRequest, "Invalid provider id")
	}

	dao := p.dbService.GetDAO()
	provider, err := dao.Provider.Where(dao.Provider.ID.Eq(uint(id))).First()
	if err != nil {
		return common.HandleError(c, fiber.StatusNotFound, "Provider not found")
	}

	var req struct {
		Name               string `json:"name"`
		Type               string `json:"type"`
		Key                string `json:"key"`
		BaseURL            string `json:"baseUrl"`
		ProxyURL           string `json:"proxyUrl"`
		IsActive           bool   `json:"isActive"`
		ClientID           string `json:"clientId,omitempty"`
		ClientSecret       string `json:"clientSecret,omitempty"`
		AuthURL            string `json:"authUrl,omitempty"`
		TokenURL           string `json:"tokenUrl,omitempty"`
		RedirectURL        string `json:"redirectUrl,omitempty"`
		Scopes             string `json:"scopes,omitempty"`
		AccountID          string `json:"accountId,omitempty"`
		HealthCheckEnabled bool   `json:"healthCheckEnabled"`
		HealthCheckStatus  string `json:"healthCheckStatus"`
	}

	if err := c.BodyParser(&req); err != nil {
		return common.HandleError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	// Apply platform-specific defaults and sanitization
	applyPlatformDefaults(&req)

	provider.Name = req.Name
	provider.Platform = consts.ProviderPlatform(req.Type)
	provider.ApiKey = keepUnlessChanged(req.Key, provider.ApiKey)
	provider.BaseURL = req.BaseURL
	provider.ProxyURL = req.ProxyURL
	provider.IsActive = req.IsActive
	provider.HealthCheckEnabled = req.HealthCheckEnabled
	if strings.TrimSpace(req.HealthCheckStatus) != "" {
		provider.HealthCheckStatus = strings.TrimSpace(req.HealthCheckStatus)
	}
	provider.ClientID = req.ClientID
	provider.ClientSecret = keepUnlessChanged(req.ClientSecret, provider.ClientSecret)
	provider.AuthURL = req.AuthURL
	provider.TokenURL = req.TokenURL
	provider.RedirectURL = req.RedirectURL
	provider.Scopes = req.Scopes
	if strings.TrimSpace(req.AccountID) != "" {
		provider.AccountID = strings.TrimSpace(req.AccountID)
	}

	if err := dao.Provider.Save(provider); err != nil {
		return common.HandleError(c, fiber.StatusInternalServerError, "Failed to update provider")
	}

	modelsList, err := dao.Model.Where(dao.Model.ProviderID.Eq(provider.ID)).Find()
	if err != nil {
		return common.HandleError(c, fiber.StatusInternalServerError, "Failed to load provider models")
	}

	return c.JSON(dto.NewProvider(provider, modelsList))
}

// applyPlatformDefaults normalizes provider URLs and redirect values for OAuth platforms.
func applyPlatformDefaults(req *struct {
	Name               string `json:"name"`
	Type               string `json:"type"`
	Key                string `json:"key"`
	BaseURL            string `json:"baseUrl"`
	ProxyURL           string `json:"proxyUrl"`
	IsActive           bool   `json:"isActive"`
	ClientID           string `json:"clientId,omitempty"`
	ClientSecret       string `json:"clientSecret,omitempty"`
	AuthURL            string `json:"authUrl,omitempty"`
	TokenURL           string `json:"tokenUrl,omitempty"`
	RedirectURL        string `json:"redirectUrl,omitempty"`
	Scopes             string `json:"scopes,omitempty"`
	AccountID          string `json:"accountId,omitempty"`
	HealthCheckEnabled bool   `json:"healthCheckEnabled"`
	HealthCheckStatus  string `json:"healthCheckStatus"`
}) {
	switch req.Type {
	case consts.ProviderPlatformCodex.String():
		req.RedirectURL = "http://localhost:1455/auth/callback"
		req.AuthURL = sanitizeCodexAuthURL(req.AuthURL)
		if req.TokenURL == "" {
			req.TokenURL = "https://auth.openai.com/oauth/token"
		}
		if req.BaseURL == "" {
			req.BaseURL = "https://api.openai.com"
		}
		if strings.TrimSpace(req.Scopes) == "" {
			req.Scopes = "openid profile email offline_access"
		}
	case consts.ProviderPlatformAntigravity.String():
		if req.AuthURL == "" {
			req.AuthURL = "https://accounts.google.com/o/oauth2/v2/authe"
		}
		// user info url: https://www.googleapis.com/oauth2/v2/userinfo
		if req.TokenURL == "" {
			req.TokenURL = "https://oauth2.googleapis.com/token"
		}
		if req.RedirectURL == "" {
			req.RedirectURL = "http://localhost:1455/auth/callback"
		}
		if req.BaseURL == "" {
			req.BaseURL = "https://api.antigravity.com"
		}
	}
}

func (p *ProviderConfigServiceImpl) HandleDeleteProvider(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return common.HandleError(c, fiber.StatusBadRequest, "Invalid provider id")
	}

	dao := p.dbService.GetDAO()
	if _, err := dao.Provider.Where(dao.Provider.ID.Eq(uint(id))).Delete(&models.Provider{}); err != nil {
		return common.HandleError(c, fiber.StatusInternalServerError, "Failed to delete provider")
	}
	return c.SendStatus(fiber.StatusNoContent)
}

func (p *ProviderConfigServiceImpl) HandleCreateOAuthSession(c *fiber.Ctx) error {
	var req struct {
		ProviderID          uint   `json:"providerId"`
		State               string `json:"state"`
		CodeVerifier        string `json:"codeVerifier"`
		CodeChallenge       string `json:"codeChallenge"`
		CodeChallengeMethod string `json:"codeChallengeMethod"`
	}

	if err := c.BodyParser(&req); err != nil {
		return common.HandleError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if req.ProviderID == 0 || req.State == "" {
		return common.HandleError(c, fiber.StatusBadRequest, "Provider id or state is missing")
	}

	if req.CodeVerifier == "" {
		return common.HandleError(c, fiber.StatusBadRequest, "Code verifier is required for OAuth2 session")
	}

	// Validate the provider exists
	dao := p.dbService.GetDAO()
	provider, err := dao.Provider.Where(dao.Provider.ID.Eq(req.ProviderID)).First()
	if err != nil {
		return common.HandleError(c, fiber.StatusNotFound, "Provider not found")
	}

	// Check if this is a Codex or Antigravity provider
	if provider.Platform != consts.ProviderPlatformCodex && provider.Platform != consts.ProviderPlatformAntigravity {
		return common.HandleError(c, fiber.StatusBadRequest, "OAuth2 is only supported for codex or antigravity providers")
	}

	// Derive and validate the code challenge from the verifier
	computedChallenge, err := generateCodeChallenge(req.CodeVerifier)
	if err != nil {
		return common.HandleError(c, fiber.StatusBadRequest, "Invalid code verifier")
	}

	method := req.CodeChallengeMethod
	if method == "" {
		method = "S256"
	}
	if strings.ToUpper(method) != "S256" {
		return common.HandleError(c, fiber.StatusBadRequest, "Unsupported code challenge method")
	}

	if req.CodeChallenge != "" && req.CodeChallenge != computedChallenge {
		return common.HandleError(c, fiber.StatusBadRequest, "Code challenge does not match the verifier")
	}

	// Generate a session ID and create the OAuth session record
	sessionID := generateSessionID()
	expiryTime := time.Now().Add(10 * time.Minute) // Session expires in 10 minutes

	session := models.OauthSession{
		SessionID:           sessionID,
		ProviderID:          req.ProviderID,
		State:               req.State,
		CodeVerifier:        req.CodeVerifier,
		CodeChallenge:       computedChallenge,
		CodeChallengeMethod: method,
		ExpiresAt:           expiryTime,
	}

	if err := dao.OauthSession.Create(&session); err != nil {
		return common.HandleError(c, fiber.StatusInternalServerError, "Failed to create OAuth session")
	}

	return c.JSON(fiber.Map{
		"sessionId": sessionID,
		"message":   "OAuth session created successfully",
	})
}

func (p *ProviderConfigServiceImpl) HandleOAuth2Callback(c *fiber.Ctx) error {
	// This endpoint is now primarily for frontend to make code exchange request
	// The frontend will call this with the authorization code and code_verifier

	var req struct {
		Code              string `json:"code"`
		State             string `json:"state"`
		CodeVerifier      string `json:"code_verifier"`
		CodeVerifierCamel string `json:"codeVerifier"`
	}

	if err := c.BodyParser(&req); err != nil {
		return common.HandleError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if req.Code == "" || req.State == "" {
		return common.HandleError(c, fiber.StatusBadRequest, "Authorization code or state is missing")
	}

	// Find the OAuth session by state
	dao := p.dbService.GetDAO()
	session, err := dao.OauthSession.Where(dao.OauthSession.State.Eq(req.State)).First()
	if err != nil {
		return common.HandleError(c, fiber.StatusNotFound, "OAuth session not found")
	}

	// Check if the session has expired
	if time.Now().After(session.ExpiresAt) {
		return common.HandleError(c, fiber.StatusBadRequest, "OAuth session has expired")
	}

	codeVerifier := req.CodeVerifier
	if codeVerifier == "" {
		codeVerifier = req.CodeVerifierCamel
	}

	if codeVerifier == "" {
		codeVerifier = session.CodeVerifier
	}

	if codeVerifier == "" {
		return common.HandleError(c, fiber.StatusBadRequest, "Code verifier is missing")
	}

	// Validate the code challenge (this is done by hashing the code_verifier and comparing with stored challenge)
	if err := validateCodeChallenge(codeVerifier, session.CodeChallenge); err != nil {
		return common.HandleError(c, fiber.StatusBadRequest, "Invalid code verifier")
	}

	// Get the provider from the database
	provider, err := dao.Provider.Where(dao.Provider.ID.Eq(session.ProviderID)).First()
	if err != nil {
		return common.HandleError(c, fiber.StatusNotFound, "Provider not found")
	}

	oauthSvc, err := p.getOAuthService(provider.Platform)
	if err != nil {
		return common.HandleError(c, fiber.StatusBadRequest, err.Error())
	}

	token, err := oauthSvc.ExchangeCode(context.Background(), provider, req.Code, codeVerifier)
	if err != nil {
		return common.HandleError(c, fiber.StatusInternalServerError, err.Error())
	}

	oauthSvc.ApplyToken(provider, token)

	// Save the updated provider to the database
	if err := dao.Provider.Save(provider); err != nil {
		return common.HandleError(c, fiber.StatusInternalServerError, "Failed to save provider tokens")
	}

	// Optionally delete the OAuth session after successful exchange
	_, _ = dao.OauthSession.Where(dao.OauthSession.ID.Eq(session.ID)).Delete(&models.OauthSession{})

	// Return success response
	return c.JSON(fiber.Map{
		"success": true,
		"message": "OAuth2 authentication successful, tokens saved",
	})
}

// generateCodeChallenge calculates the PKCE code challenge from a verifier using S256
func generateCodeChallenge(codeVerifier string) (string, error) {
	if codeVerifier == "" {
		return "", errors.New("code verifier cannot be empty")
	}

	hash := sha256.Sum256([]byte(codeVerifier))
	encoded := base64.RawURLEncoding.EncodeToString(hash[:])
	return encoded, nil
}

// validateCodeChallenge validates the code verifier against the stored code challenge
func validateCodeChallenge(codeVerifier, expectedChallenge string) error {
	encoded, err := generateCodeChallenge(codeVerifier)
	if err != nil {
		return err
	}

	if encoded != expectedChallenge {
		return errors.New("invalid code verifier")
	}
	return nil
}

// Helper function to generate a unique session ID
func generateSessionID() string {
	// Create a random session ID using crypto/rand
	b := make([]byte, 32)
	_, _ = rand.Read(b)
	return fmt.Sprintf("%x", b)
}

// sanitizeCodexAuthURL normalizes Codex auth URL to match the VS Code flow and
// avoid extra org consent prompts.
func sanitizeCodexAuthURL(raw string) string {
	if raw == "" {
		return "https://auth.openai.com/oauth/authorize"
	}

	parsed, err := url.Parse(raw)
	if err != nil {
		return raw
	}

	query := parsed.Query()
	query.Set("id_token_add_organizations", "false")
	if query.Get("codex_cli_simplified_flow") == "" {
		query.Set("codex_cli_simplified_flow", "true")
	}
	if query.Get("originator") == "" {
		query.Set("originator", "codex_vscode")
	}
	parsed.RawQuery = query.Encode()
	return parsed.String()
}

func sanitizeTokenURL(raw string) string {
	if raw == "" {
		return "https://auth.openai.com/oauth/token"
	}
	return raw
}

func (p *ProviderConfigServiceImpl) getOAuthService(platform consts.ProviderPlatform) (oauthService, error) {
	switch platform {
	case consts.ProviderPlatformCodex:
		return p.codexOAuth, nil
	case consts.ProviderPlatformAntigravity:
		return p.antigravityOAuth, nil
	default:
		return nil, fmt.Errorf("OAuth2 callback is only supported for codex or antigravity providers")
	}
}

func (p *ProviderConfigServiceImpl) HandleRefreshToken(c *fiber.Ctx) error {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return common.HandleError(c, fiber.StatusBadRequest, "Invalid provider id")
	}

	dao := p.dbService.GetDAO()
	provider, err := dao.Provider.Where(dao.Provider.ID.Eq(uint(id))).First()
	if err != nil {
		return common.HandleError(c, fiber.StatusNotFound, "Provider not found")
	}

	// Check if the provider has OAuth configuration (RefreshToken is not empty)
	if provider.RefreshToken == "" {
		return common.HandleError(c, fiber.StatusBadRequest, "Provider does not have OAuth configuration or refresh token")
	}

	// Check if the provider is an OAuth provider
	if provider.Platform != consts.ProviderPlatformCodex && provider.Platform != consts.ProviderPlatformAntigravity {
		return common.HandleError(c, fiber.StatusBadRequest, "Token refresh is only supported for OAuth providers (codex/antigravity)")
	}

	// Use the health check service to refresh the token
	if err := p.healthCheckService.RefreshProviderToken(provider); err != nil {
		return common.HandleError(c, fiber.StatusInternalServerError, fmt.Sprintf("Failed to refresh token: %v", err))
	}

	// Return updated provider information
	modelsList, err := dao.Model.Where(dao.Model.ProviderID.Eq(provider.ID)).Find()
	if err != nil {
		return common.HandleError(c, fiber.StatusInternalServerError, "Failed to load provider models")
	}

	return c.JSON(dto.NewProvider(provider, modelsList))
}

// keepUnlessChanged resolves a secret submitted by the admin UI. The UI renders
// the masked value returned by dto.NewProvider, so a submission that still
// matches that mask means the field was left untouched and the stored secret
// must survive the update.
func keepUnlessChanged(submitted, current string) string {
	if current != "" && submitted == dto.MaskSecret(current) {
		return current
	}
	return submitted
}
