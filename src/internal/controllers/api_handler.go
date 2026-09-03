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
	setupService          services.SetupService
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
	setupService services.SetupService,
) *Controller {
	return &Controller{
		apiService:            apiService,
		logService:            logService,
		modelService:          modelService,
		providerConfigService: providerConfigService,
		requestService:        requestService,
		settingsService:       settingsService,
		statsService:          statsService,
		setupService:          setupService,
	}
}

// RegisterRoutes registers all the API routes.
//
// Routes fall into three tiers: public (login and first-run setup), any
// authenticated user (their own profile and API keys), and super users only
// (providers, settings, and anything spanning all users).
func RegisterRoutes(app *fiber.App, controller *Controller, authMiddleware, adminMiddleware fiber.Handler) {
	api := app.Group("/api")
	api.Post("/login", controller.HandleLogin)
	api.Get("/setup/status", controller.HandleGetSetupStatus)
	api.Post("/setup", controller.HandleSetup)

	apiAuthorized := api.Group("", authMiddleware)
	apiAuthorized.Post("/logout", controller.HandleLogout)
	apiAuthorized.Post("/reset-password", controller.HandleResetPassword)
	apiAuthorized.Get("/user", controller.HandleGetUser)
	apiAuthorized.Put("/user", controller.HandleUpdateCurrentUser)
	apiAuthorized.Get("/keys", controller.HandleGetAPIKeys)
	apiAuthorized.Post("/keys", controller.HandleCreateAPIKey)
	apiAuthorized.Delete("/keys/:id", controller.HandleDeleteAPIKey)
	apiAuthorized.Get("/models", controller.HandleGetModels)

	apiAdmin := apiAuthorized.Group("", adminMiddleware)
	apiAdmin.Post("/models", controller.HandleCreateModel)
	apiAdmin.Put("/models/:id", controller.HandleUpdateModel)
	apiAdmin.Delete("/models/:id", controller.HandleDeleteModel)
	apiAdmin.Get("/providers", controller.HandleGetProviders)
	apiAdmin.Get("/platforms", controller.HandleGetProviderPlatforms)
	apiAdmin.Post("/providers", controller.HandleCreateProvider)
	apiAdmin.Put("/providers/:id", controller.HandleUpdateProvider)
	apiAdmin.Delete("/providers/:id", controller.HandleDeleteProvider)
	apiAdmin.Post("/providers/:id/refresh-token", controller.HandleRefreshToken)
	apiAdmin.Get("/requests", controller.HandleGetRequests)
	apiAdmin.Get("/logs", controller.HandleGetLogs)
	apiAdmin.Get("/stats", controller.HandleGetStats)
	apiAdmin.Get("/users", controller.HandleListUsers)
	apiAdmin.Post("/users", controller.HandleCreateUser)
	apiAdmin.Put("/users/:id", controller.HandleUpdateUser)
	apiAdmin.Delete("/users/:id", controller.HandleDeleteUser)
	apiAdmin.Get("/settings", controller.HandleGetSettings)
	apiAdmin.Put("/settings", controller.HandleUpdateSettings)

	// OAuth2 endpoints - the provider redirects a browser here, so they cannot
	// require an Authorization header. The session id issued at /oauth2/session
	// (an admin route) is what ties a callback to a provider.
	api.Get("/oauth2/callback", controller.HandleOAuth2Callback)
	api.Post("/oauth2/callback", controller.HandleOAuth2Callback)
	apiAdmin.Post("/oauth2/session", controller.HandleCreateOAuthSession)
}

func (c *Controller) HandleGetSetupStatus(ctx *fiber.Ctx) error {
	return c.setupService.HandleGetSetupStatus(ctx)
}

func (c *Controller) HandleSetup(ctx *fiber.Ctx) error {
	return c.setupService.HandleSetup(ctx)
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
