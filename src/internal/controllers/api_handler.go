package controllers

import (
	"github.com/fdddf/openproxy/internal/services"
	"github.com/gofiber/fiber/v2"
)

// Controller holds all the service dependencies
type Controller struct {
	apiService            services.APIService
	logService            services.LogService
	modelService          services.ModelService
	providerConfigService services.ProviderConfigService
	requestService        services.RequestService
	settingsService       services.SettingsService
	statsService          services.StatsService
}

// NewController creates a new controller instance
func NewController(
	apiService services.APIService,
	logService services.LogService,
	modelService services.ModelService,
	providerConfigService services.ProviderConfigService,
	requestService services.RequestService,
	settingsService services.SettingsService,
	statsService services.StatsService,
) *Controller {
	return &Controller{
		apiService:            apiService,
		logService:            logService,
		modelService:          modelService,
		providerConfigService: providerConfigService,
		requestService:        requestService,
		settingsService:       settingsService,
		statsService:          statsService,
	}
}

// RegisterRoutes registers all the API routes
func RegisterRoutes(app *fiber.App, controller *Controller, authMiddleware fiber.Handler) {
	api := app.Group("/api")
	api.Post("/login", controller.HandleLogin)

	apiAuthorized := api.Group("", authMiddleware)
	apiAuthorized.Post("/logout", controller.HandleLogout)
	apiAuthorized.Post("/reset-password", controller.HandleResetPassword)
	apiAuthorized.Get("/keys", controller.HandleGetAPIKeys)
	apiAuthorized.Post("/keys", controller.HandleCreateAPIKey)
	apiAuthorized.Delete("/keys/:id", controller.HandleDeleteAPIKey)
	apiAuthorized.Get("/logs", controller.HandleGetLogs)
	apiAuthorized.Get("/models", controller.HandleGetModels)
	apiAuthorized.Post("/models", controller.HandleCreateModel)
	apiAuthorized.Put("/models/:id", controller.HandleUpdateModel)
	apiAuthorized.Delete("/models/:id", controller.HandleDeleteModel)
	apiAuthorized.Get("/providers", controller.HandleGetProviders)
	apiAuthorized.Get("/platforms", controller.HandleGetProviderPlatforms)
	apiAuthorized.Post("/providers", controller.HandleCreateProvider)
	apiAuthorized.Put("/providers/:id", controller.HandleUpdateProvider)
	apiAuthorized.Delete("/providers/:id", controller.HandleDeleteProvider)
	apiAuthorized.Post("/providers/:id/refresh-token", controller.HandleRefreshToken)
	apiAuthorized.Get("/requests", controller.HandleGetRequests)
	apiAuthorized.Get("/users", controller.HandleListUsers)
	apiAuthorized.Post("/users", controller.HandleCreateUser)
	apiAuthorized.Put("/users/:id", controller.HandleUpdateUser)
	apiAuthorized.Delete("/users/:id", controller.HandleDeleteUser)
	apiAuthorized.Get("/settings", controller.HandleGetSettings)
	apiAuthorized.Put("/settings", controller.HandleUpdateSettings)
	apiAuthorized.Get("/user", controller.HandleGetUser)
	apiAuthorized.Put("/user", controller.HandleUpdateCurrentUser)
	apiAuthorized.Get("/stats", controller.HandleGetStats)

	// OAuth2 endpoints - accessible without authentication
	api.Get("/oauth2/callback", controller.HandleOAuth2Callback)  // This can still be used for redirects
	api.Post("/oauth2/callback", controller.HandleOAuth2Callback) // This is for the actual token exchange
	api.Post("/oauth2/session", controller.HandleCreateOAuthSession)
}

func (c *Controller) HandleLogin(ctx *fiber.Ctx) error {
	return c.apiService.HandleLogin(ctx)
}

func (c *Controller) HandleResetPassword(ctx *fiber.Ctx) error {
	return c.apiService.HandleResetPassword(ctx)
}

func (c *Controller) HandleLogout(ctx *fiber.Ctx) error {
	return c.apiService.HandleLogout(ctx)
}

func (c *Controller) HandleGetAPIKeys(ctx *fiber.Ctx) error {
	return c.apiService.HandleGetAPIKeys(ctx)
}

func (c *Controller) HandleCreateAPIKey(ctx *fiber.Ctx) error {
	return c.apiService.HandleCreateAPIKey(ctx)
}

func (c *Controller) HandleDeleteAPIKey(ctx *fiber.Ctx) error {
	return c.apiService.HandleDeleteAPIKey(ctx)
}

func (c *Controller) HandleGetLogs(ctx *fiber.Ctx) error {
	return c.logService.HandleGetLogs(ctx)
}

func (c *Controller) HandleGetModels(ctx *fiber.Ctx) error {
	return c.modelService.HandleGetModels(ctx)
}

func (c *Controller) HandleCreateModel(ctx *fiber.Ctx) error {
	return c.modelService.HandleCreateModel(ctx)
}

func (c *Controller) HandleUpdateModel(ctx *fiber.Ctx) error {
	return c.modelService.HandleUpdateModel(ctx)
}

func (c *Controller) HandleDeleteModel(ctx *fiber.Ctx) error {
	return c.modelService.HandleDeleteModel(ctx)
}

func (c *Controller) HandleGetProviders(ctx *fiber.Ctx) error {
	return c.providerConfigService.HandleGetProviders(ctx)
}

func (c *Controller) HandleGetProviderPlatforms(ctx *fiber.Ctx) error {
	return c.providerConfigService.HandleGetProviderPlatforms(ctx)
}

func (c *Controller) HandleCreateProvider(ctx *fiber.Ctx) error {
	return c.providerConfigService.HandleCreateProvider(ctx)
}

func (c *Controller) HandleUpdateProvider(ctx *fiber.Ctx) error {
	return c.providerConfigService.HandleUpdateProvider(ctx)
}

func (c *Controller) HandleDeleteProvider(ctx *fiber.Ctx) error {
	return c.providerConfigService.HandleDeleteProvider(ctx)
}

func (c *Controller) HandleGetRequests(ctx *fiber.Ctx) error {
	return c.requestService.HandleGetRequests(ctx)
}

func (c *Controller) HandleGetSettings(ctx *fiber.Ctx) error {
	return c.settingsService.HandleGetSettings(ctx)
}

func (c *Controller) HandleUpdateSettings(ctx *fiber.Ctx) error {
	return c.settingsService.HandleUpdateSettings(ctx)
}

func (c *Controller) HandleGetUser(ctx *fiber.Ctx) error {
	return c.apiService.HandleGetUser(ctx)
}

func (c *Controller) HandleUpdateCurrentUser(ctx *fiber.Ctx) error {
	return c.apiService.HandleUpdateCurrentUser(ctx)
}

func (c *Controller) HandleListUsers(ctx *fiber.Ctx) error {
	return c.apiService.HandleListUsers(ctx)
}

func (c *Controller) HandleCreateUser(ctx *fiber.Ctx) error {
	return c.apiService.HandleCreateUser(ctx)
}

func (c *Controller) HandleUpdateUser(ctx *fiber.Ctx) error {
	return c.apiService.HandleUpdateUser(ctx)
}

func (c *Controller) HandleDeleteUser(ctx *fiber.Ctx) error {
	return c.apiService.HandleDeleteUser(ctx)
}

func (c *Controller) HandleGetStats(ctx *fiber.Ctx) error {
	return c.statsService.HandleGetStats(ctx)
}

func (c *Controller) HandleOAuth2Callback(ctx *fiber.Ctx) error {
	return c.providerConfigService.HandleOAuth2Callback(ctx)
}

func (c *Controller) HandleCreateOAuthSession(ctx *fiber.Ctx) error {
	return c.providerConfigService.HandleCreateOAuthSession(ctx)
}

func (c *Controller) HandleRefreshToken(ctx *fiber.Ctx) error {
	return c.providerConfigService.HandleRefreshToken(ctx)
}
