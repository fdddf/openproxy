package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/fdddf/openproxy/common"
	"github.com/fdddf/openproxy/internal/dao"
	"github.com/fdddf/openproxy/internal/models"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/valyala/fasthttp"
	"golang.org/x/oauth2"
)

// HealthCheckService periodically validates providers and refreshes tokens.
type HealthCheckService interface {
	Start()
	Stop()
	RefreshProviderToken(provider *models.Provider) error
}

// healthCheckInterval is how often providers are re-checked.
const healthCheckInterval = time.Hour

type healthCheckService struct {
	dbService       DatabaseService
	settingsService SettingsService
	providerService ProviderService

	done      chan struct{}
	startOnce sync.Once
	stopOnce  sync.Once
	wg        sync.WaitGroup
}

// NewHealthCheckService creates a background health checker.
func NewHealthCheckService(dbService DatabaseService, settingsService SettingsService, providerService ProviderService) HealthCheckService {
	return &healthCheckService{
		dbService:       dbService,
		settingsService: settingsService,
		providerService: providerService,
		done:            make(chan struct{}),
	}
}

func (h *healthCheckService) Start() {
	h.startOnce.Do(func() {
		h.wg.Add(1)
		go h.loop()
	})
}

// Stop signals the loop to exit and waits for the in-flight pass to finish.
func (h *healthCheckService) Stop() {
	h.stopOnce.Do(func() { close(h.done) })
	h.wg.Wait()
}

func (h *healthCheckService) loop() {
	defer h.wg.Done()

	ticker := time.NewTicker(healthCheckInterval)
	defer ticker.Stop()

	h.runOnce()
	for {
		select {
		case <-ticker.C:
			h.runOnce()
		case <-h.done:
			return
		}
	}
}

func (h *healthCheckService) runOnce() {
	settings, err := h.settingsService.GetEffectiveSettings()
	if err != nil {
		log.Errorf("health check: load settings: %v", err)
		return
	}

	if !settings.HealthCheckEnabled {
		log.Debug("health check: disabled")
		return
	}

	dao := h.dbService.GetDAO()
	providers, err := dao.Provider.Where(dao.Provider.IsActive.Is(true)).Find()
	if err != nil {
		log.Errorf("health check: load providers: %v", err)
		return
	}

	for _, p := range providers {
		if err := h.checkProvider(p); err != nil {
			log.Warnf("health check: provider %d (%s) failed: %v", p.ID, p.Name, err)
			h.recordStatus(dao, p, "unhealthy")
			continue
		}
		h.recordStatus(dao, p, "healthy")
	}
}

func (h *healthCheckService) checkProvider(provider *models.Provider) error {
	dao := h.dbService.GetDAO()
	model, err := dao.Model.Where(dao.Model.ProviderID.Eq(provider.ID), dao.Model.IsActive.Is(true)).Limit(1).First()
	if err != nil {
		return fmt.Errorf("load model: %w", err)
	}

	providerCfg := &common.ProviderConfig{
		Type:         provider.Platform.String(),
		ApiKey:       provider.ApiKey,
		BaseURL:      provider.BaseURL,
		ProxyURL:     provider.ProxyURL,
		AccessToken:  provider.AccessToken,
		RefreshToken: provider.RefreshToken,
		TokenExpiry:  timePtrValue(provider.TokenExpiry),
		ClientID:     provider.ClientID,
		ClientSecret: provider.ClientSecret,
		AuthURL:      provider.AuthURL,
		TokenURL:     provider.TokenURL,
		RedirectURL:  provider.RedirectURL,
		Scopes:       provider.Scopes,
		AccountID:    provider.AccountID,
	}

	providerImpl, err := h.providerService.GetModelProvider(providerCfg, model.RealModel)
	if err != nil {
		return fmt.Errorf("build provider: %w", err)
	}

	ctx := newFiberCtx()
	defer ctx.app.ReleaseCtx(ctx.ctx)

	chatReq := common.ChatCompletionRequest{
		Model: model.RealModel,
		Messages: []common.ChatMessage{
			{Role: "user", Content: "health check"},
		},
	}
	body, _ := json.Marshal(chatReq)
	ctx.ctx.Request().SetBody(body)
	ctx.ctx.Request().SetRequestURI("/v1/chat/completions")
	ctx.ctx.Request().Header.SetContentType("application/json")

	aiReq := common.AIRequest{Model: model.RealModel, Stream: false}
	prepared, err := providerImpl.PrepareRequestBody(ctx.ctx, aiReq)
	if err != nil {
		return fmt.Errorf("prepare request: %w", err)
	}

	// Refresh OAuth tokens when expired and refresh token is available.
	if providerCfg.RefreshToken != "" && !providerCfg.TokenExpiry.IsZero() && time.Now().After(providerCfg.TokenExpiry) {
		if err := h.refreshOAuthToken(provider, providerCfg); err != nil {
			return fmt.Errorf("refresh token: %w", err)
		}
	}

	req, err := http.NewRequest(providerImpl.RequestMethod(), providerImpl.GetChatCompletionURL(ctx.ctx), bytes.NewReader(prepared))
	if err != nil {
		return fmt.Errorf("build request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := providerImpl.SendRequest(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return fmt.Errorf("unexpected status: %d", resp.StatusCode)
	}

	return nil
}

func (h *healthCheckService) recordStatus(d *dao.Query, provider *models.Provider, status string) {
	now := time.Now().UTC()
	provider.HealthCheckStatus = status
	provider.HealthCheckChecked = &now

	if err := d.Provider.Save(provider); err != nil {
		log.Warnf("health check: save status for provider %d: %v", provider.ID, err)
	}
}

func (h *healthCheckService) refreshOAuthToken(provider *models.Provider, cfg *common.ProviderConfig) error {
	if cfg.TokenURL == "" || cfg.ClientID == "" {
		return fmt.Errorf("missing oauth config")
	}

	oauthCfg := &oauth2.Config{
		ClientID:     cfg.ClientID,
		ClientSecret: cfg.ClientSecret,
		Endpoint:     oauth2.Endpoint{TokenURL: cfg.TokenURL},
		RedirectURL:  cfg.RedirectURL,
		Scopes:       common.SplitScopes(cfg.Scopes),
	}

	token := &oauth2.Token{
		AccessToken:  cfg.AccessToken,
		RefreshToken: cfg.RefreshToken,
		Expiry:       cfg.TokenExpiry,
	}

	newToken, err := oauthCfg.TokenSource(context.Background(), token).Token()
	if err != nil {
		return err
	}

	dao := h.dbService.GetDAO()
	provider.AccessToken = newToken.AccessToken
	provider.RefreshToken = newToken.RefreshToken
	provider.TokenExpiry = &newToken.Expiry

	if err := dao.Provider.Save(provider); err != nil {
		return fmt.Errorf("save provider token: %w", err)
	}

	return nil
}

func (h *healthCheckService) RefreshProviderToken(provider *models.Provider) error {
	// Build provider configuration from provider model
	providerCfg := &common.ProviderConfig{
		Type:         provider.Platform.String(),
		ApiKey:       provider.ApiKey,
		BaseURL:      provider.BaseURL,
		ProxyURL:     provider.ProxyURL,
		AccessToken:  provider.AccessToken,
		RefreshToken: provider.RefreshToken,
		TokenExpiry:  timePtrValue(provider.TokenExpiry),
		ClientID:     provider.ClientID,
		ClientSecret: provider.ClientSecret,
		AuthURL:      provider.AuthURL,
		TokenURL:     provider.TokenURL,
		RedirectURL:  provider.RedirectURL,
		Scopes:       provider.Scopes,
		AccountID:    provider.AccountID,
	}

	// Use the existing refresh logic
	return h.refreshOAuthToken(provider, providerCfg)
}

func timePtrValue(t *time.Time) time.Time {
	if t == nil {
		return time.Time{}
	}
	return *t
}

type fiberCtx struct {
	app *fiber.App
	ctx *fiber.Ctx
}

func newFiberCtx() *fiberCtx {
	app := fiber.New()
	fc := app.AcquireCtx(&fasthttp.RequestCtx{})
	return &fiberCtx{app: app, ctx: fc}
}
