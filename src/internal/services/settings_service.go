package services

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/fdddf/openproxy/common"
	"github.com/fdddf/openproxy/internal/dto"
	"github.com/fdddf/openproxy/internal/models"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"gorm.io/gorm"
)

const (
	settingKeyDefaultProvider = "default_provider"
	settingKeyRateLimit       = "rate_limit"
	settingKeyConcurrentLimit = "concurrent_limit"
	settingKeyLogLevel        = "log_level"
	settingKeyCacheTTL        = "cache_ttl"
	settingKeyMaintenanceMode = "maintenance_mode"
	settingKeyHealthCheck     = "health_check_enabled"
)

var allowedLogLevels = map[string]struct{}{
	"debug": {},
	"info":  {},
	"warn":  {},
	"error": {},
}

// SettingsServiceImpl implements the SettingsService interface
type SettingsServiceImpl struct {
	dbService     DatabaseService
	configService ConfigService

	mu             sync.RWMutex
	cachedSettings dto.SystemSettings
	loaded         bool
}

// NewSettingsService creates a new SettingsService instance
func NewSettingsService(dbService DatabaseService, configService ConfigService) SettingsService {
	return &SettingsServiceImpl{
		dbService:     dbService,
		configService: configService,
	}
}

func (s *SettingsServiceImpl) HandleGetSettings(c *fiber.Ctx) error {
	settings, err := s.GetEffectiveSettings()
	if err != nil {
		return common.HandleError(c, fiber.StatusInternalServerError, "Failed to load settings")
	}

	return c.JSON(settings)
}

func (s *SettingsServiceImpl) HandleUpdateSettings(c *fiber.Ctx) error {
	var req dto.SystemSettings
	if err := c.BodyParser(&req); err != nil {
		return common.HandleError(c, fiber.StatusBadRequest, "Invalid request body")
	}
	req.LogLevel = strings.ToLower(req.LogLevel)

	if err := s.validateSettings(req); err != nil {
		return common.HandleError(c, fiber.StatusBadRequest, err.Error())
	}

	if err := s.saveSettings(req); err != nil {
		return common.HandleError(c, fiber.StatusInternalServerError, "Failed to save settings")
	}

	s.setCached(req)

	return c.JSON(req)
}

// GetEffectiveSettings returns the cached settings, loading them if needed.
func (s *SettingsServiceImpl) GetEffectiveSettings() (dto.SystemSettings, error) {
	s.mu.RLock()
	if s.loaded {
		defer s.mu.RUnlock()
		return s.cachedSettings, nil
	}
	s.mu.RUnlock()

	settings, err := s.loadSettings()
	if err != nil {
		return dto.SystemSettings{}, err
	}
	s.setCached(settings)
	return settings, nil
}

func (s *SettingsServiceImpl) loadSettings() (dto.SystemSettings, error) {
	settings := dto.DefaultSystemSettings(s.defaultProviderFromConfig())

	dao := s.dbService.GetDAO()
	stored, err := dao.Setting.Find()
	if err != nil {
		return dto.SystemSettings{}, fmt.Errorf("load stored settings: %w", err)
	}

	for _, setting := range stored {
		switch setting.Key {
		case settingKeyDefaultProvider:
			settings.DefaultProvider = setting.Value
		case settingKeyRateLimit:
			if v, err := strconv.Atoi(setting.Value); err == nil {
				settings.RateLimit = v
			} else {
				log.Warnf("invalid rate limit setting %q: %v", setting.Value, err)
			}
		case settingKeyConcurrentLimit:
			if v, err := strconv.Atoi(setting.Value); err == nil {
				settings.ConcurrentLimit = v
			} else {
				log.Warnf("invalid concurrent limit setting %q: %v", setting.Value, err)
			}
		case settingKeyLogLevel:
			level := strings.ToLower(setting.Value)
			if _, ok := allowedLogLevels[level]; ok {
				settings.LogLevel = level
			}
		case settingKeyCacheTTL:
			if v, err := strconv.Atoi(setting.Value); err == nil {
				settings.CacheTTL = v
			} else {
				log.Warnf("invalid cache ttl setting %q: %v", setting.Value, err)
			}
		case settingKeyMaintenanceMode:
			if v, err := strconv.ParseBool(setting.Value); err == nil {
				settings.MaintenanceMode = v
			} else {
				log.Warnf("invalid maintenance mode setting %q: %v", setting.Value, err)
			}
		case settingKeyHealthCheck:
			if v, err := strconv.ParseBool(setting.Value); err == nil {
				settings.HealthCheckEnabled = v
			} else {
				log.Warnf("invalid health check setting %q: %v", setting.Value, err)
			}
		}
	}

	if err := s.applyDefaultProviderToConfig(settings.DefaultProvider); err != nil {
		log.Warnf("apply default provider to config: %v", err)
	}
	return settings, nil
}

func (s *SettingsServiceImpl) saveSettings(settings dto.SystemSettings) error {
	if settings.DefaultProvider != "" {
		if err := s.applyDefaultProviderToConfig(settings.DefaultProvider); err != nil {
			return err
		}
	}

	logLevel := strings.ToLower(settings.LogLevel)

	values := map[string]string{
		settingKeyDefaultProvider: settings.DefaultProvider,
		settingKeyRateLimit:       strconv.Itoa(settings.RateLimit),
		settingKeyConcurrentLimit: strconv.Itoa(settings.ConcurrentLimit),
		settingKeyLogLevel:        logLevel,
		settingKeyCacheTTL:        strconv.Itoa(settings.CacheTTL),
		settingKeyMaintenanceMode: strconv.FormatBool(settings.MaintenanceMode),
		settingKeyHealthCheck:     strconv.FormatBool(settings.HealthCheckEnabled),
	}

	dao := s.dbService.GetDAO()

	for key, value := range values {
		setting, err := dao.Setting.Where(dao.Setting.Key.Eq(key)).First()
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				// Create new setting if not found
				newSetting := &models.Setting{
					Key:   key,
					Value: value,
				}
				if err := dao.Setting.Create(newSetting); err != nil {
					return fmt.Errorf("create setting %s: %w", key, err)
				}
			} else {
				return fmt.Errorf("find setting %s: %w", key, err)
			}
		} else {
			// Update existing setting
			setting.Value = value
			if err := dao.Setting.Save(setting); err != nil {
				return fmt.Errorf("update setting %s: %w", key, err)
			}
		}
	}

	return nil
}

func (s *SettingsServiceImpl) validateSettings(settings dto.SystemSettings) error {
	if settings.RateLimit <= 0 {
		return fmt.Errorf("rateLimit must be greater than 0")
	}
	if settings.ConcurrentLimit <= 0 {
		return fmt.Errorf("concurrentLimit must be greater than 0")
	}
	if settings.CacheTTL < 0 {
		return fmt.Errorf("cacheTTL must be 0 or greater")
	}
	logLevel := strings.ToLower(settings.LogLevel)
	if _, ok := allowedLogLevels[logLevel]; !ok {
		return fmt.Errorf("logLevel must be one of debug, info, warn, error")
	}
	if settings.DefaultProvider != "" {
		if err := s.validateDefaultProvider(settings.DefaultProvider); err != nil {
			return err
		}
	}
	return nil
}

func (s *SettingsServiceImpl) validateDefaultProvider(providerID string) error {
	id, err := strconv.ParseUint(providerID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid default provider")
	}

	dao := s.dbService.GetDAO()
	if _, err := dao.Provider.Where(dao.Provider.ID.Eq(uint(id))).First(); err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("default provider not found")
		}
		return fmt.Errorf("validate default provider: %w", err)
	}

	return nil
}

func (s *SettingsServiceImpl) defaultProviderFromConfig() string {
	platform := s.configService.GetConfig().Proxy.DefaultProvider
	if platform == "" {
		return ""
	}

	dao := s.dbService.GetDAO()
	provider, err := dao.Provider.Where(dao.Provider.Platform.Eq(platform)).First()
	if err != nil {
		if err != gorm.ErrRecordNotFound {
			log.Warnf("load default provider from config: %v", err)
		}
		return ""
	}

	return strconv.FormatUint(uint64(provider.ID), 10)
}

func (s *SettingsServiceImpl) applyDefaultProviderToConfig(providerID string) error {
	if providerID == "" {
		return nil
	}

	id, err := strconv.ParseUint(providerID, 10, 64)
	if err != nil {
		return fmt.Errorf("invalid default provider")
	}

	dao := s.dbService.GetDAO()
	provider, err := dao.Provider.Where(dao.Provider.ID.Eq(uint(id))).First()
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return fmt.Errorf("default provider not found")
		}
		return fmt.Errorf("load default provider: %w", err)
	}

	s.configService.GetConfig().Proxy.DefaultProvider = string(provider.Platform)
	return nil
}

func (s *SettingsServiceImpl) setCached(settings dto.SystemSettings) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cachedSettings = settings
	s.loaded = true
}
