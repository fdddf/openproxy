package services

import (
	"github.com/fdddf/openproxy/common"
	"github.com/fdddf/openproxy/internal/dao"
	"github.com/fdddf/openproxy/internal/dto"
	"github.com/fdddf/openproxy/internal/models"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// ConfigService defines the interface for configuration management
type ConfigService interface {
	GetConfig() *common.Config
	LoadConfig(configPath string) error
}

// DatabaseService defines the interface for database operations
type DatabaseService interface {
	GetDB() *gorm.DB
	GetDAO() *dao.Query
	InitDatabase() error
}

// ProviderService defines the interface for AI provider operations
type ProviderService interface {
	GetModelProvider(providerConfig *common.ProviderConfig, model string) (common.AIProvider, error)
}

// ChatService defines the interface for chat completion operations
type ChatService interface {
	HandleChatCompletion(c *fiber.Ctx) error
	HandleResponses(c *fiber.Ctx) error
	HandleModels(c *fiber.Ctx) error
	FindProvider(c *fiber.Ctx, model string) (common.AIProvider, *modelMapping, error)
}

// APIService defines the interface for API key management
type APIService interface {
	// Existing methods for user authentication and management
	HandleLogin(c *fiber.Ctx) error
	HandleLogout(c *fiber.Ctx) error
	HandleGetAPIKeys(c *fiber.Ctx) error
	HandleCreateAPIKey(c *fiber.Ctx) error
	HandleDeleteAPIKey(c *fiber.Ctx) error
	HandleGetUser(c *fiber.Ctx) error
	HandleUpdateCurrentUser(c *fiber.Ctx) error
	HandleListUsers(c *fiber.Ctx) error
	HandleCreateUser(c *fiber.Ctx) error
	HandleUpdateUser(c *fiber.Ctx) error
	HandleDeleteUser(c *fiber.Ctx) error
	HandleResetPassword(c *fiber.Ctx) error
	HandleGetSettings(c *fiber.Ctx) error
	HandleUpdateSettings(c *fiber.Ctx) error

	// Additional methods for API key quota management
	CheckAPIKeyQuota(ctx *fiber.Ctx, apiKey string) (bool, error)
	UpdateAPIKeyUsage(apiKey string) error
	GetAPIKeyInfo(apiKey string) (*models.APIKey, error)
	UpdateAPIKeyQuota(apiKey string, quota *int64) error
}

// SettingsService defines the interface for settings management
type SettingsService interface {
	HandleGetSettings(c *fiber.Ctx) error
	HandleUpdateSettings(c *fiber.Ctx) error
	GetEffectiveSettings() (dto.SystemSettings, error)
}

// SecretService defines the interface for secret management
type SecretService interface {
	ValidateProxyAPIKey(c *fiber.Ctx, apiKey string) (bool, error)
}

// ModelService defines the interface for model management
type ModelService interface {
	HandleGetModels(c *fiber.Ctx) error
	HandleCreateModel(c *fiber.Ctx) error
	HandleUpdateModel(c *fiber.Ctx) error
	HandleDeleteModel(c *fiber.Ctx) error
}

// ProviderConfigService defines the interface for provider configuration management
type ProviderConfigService interface {
	HandleGetProviders(c *fiber.Ctx) error
	HandleGetProviderPlatforms(c *fiber.Ctx) error
	HandleCreateProvider(c *fiber.Ctx) error
	HandleUpdateProvider(c *fiber.Ctx) error
	HandleDeleteProvider(c *fiber.Ctx) error
	HandleOAuth2Callback(c *fiber.Ctx) error
	HandleCreateOAuthSession(c *fiber.Ctx) error
	HandleRefreshToken(c *fiber.Ctx) error
}

// LogService defines the interface for log management
type LogService interface {
	HandleGetLogs(c *fiber.Ctx) error
}

// RequestService defines the interface for request management
type RequestService interface {
	HandleGetRequests(c *fiber.Ctx) error
}

// StatsService defines the interface for statistics operations
type StatsService interface {
	HandleGetStats(c *fiber.Ctx) error
}
