package services

import (
	"fmt"
	"sync"

	"gorm.io/gorm"
)

// Container holds all the dependencies
type Container struct {
	dependencies map[string]interface{}
	mutex        sync.RWMutex
}

// NewContainer creates a new dependency injection container
func NewContainer() *Container {
	return &Container{
		dependencies: make(map[string]interface{}),
	}
}

// Register registers a dependency with the container
func (c *Container) Register(name string, dependency interface{}) {
	c.mutex.Lock()
	defer c.mutex.Unlock()
	c.dependencies[name] = dependency
}

// Get retrieves a dependency from the container
func (c *Container) Get(name string) (interface{}, error) {
	c.mutex.RLock()
	defer c.mutex.RUnlock()

	dep, exists := c.dependencies[name]
	if !exists {
		return nil, fmt.Errorf("dependency '%s' not found", name)
	}

	return dep, nil
}

// MustGet retrieves a dependency from the container, panics if not found
func (c *Container) MustGet(name string) interface{} {
	dep, err := c.Get(name)
	if err != nil {
		panic(err)
	}
	return dep
}

// GetDB returns the database instance
func (c *Container) GetDB() *gorm.DB {
	if db, ok := c.MustGet("db").(*gorm.DB); ok {
		return db
	}
	return nil
}

// GetConfig returns the config instance
func (c *Container) GetConfig() interface{} {
	return c.MustGet("config")
}

// Getter methods for specific services
func (c *Container) GetConfigService() ConfigService {
	return c.MustGet("config_service").(ConfigService)
}

func (c *Container) GetDatabaseService() DatabaseService {
	return c.MustGet("db_service").(DatabaseService)
}

func (c *Container) GetProviderService() ProviderService {
	return c.MustGet("provider_service").(ProviderService)
}

func (c *Container) GetChatService() ChatService {
	return c.MustGet("chat_service").(ChatService)
}

func (c *Container) GetAPIService() APIService {
	return c.MustGet("api_service").(APIService)
}

func (c *Container) GetModelService() ModelService {
	return c.MustGet("model_service").(ModelService)
}

func (c *Container) GetProviderConfigService() ProviderConfigService {
	return c.MustGet("provider_config_service").(ProviderConfigService)
}

func (c *Container) GetLogService() LogService {
	return c.MustGet("log_service").(LogService)
}

func (c *Container) GetRequestService() RequestService {
	return c.MustGet("request_service").(RequestService)
}

func (c *Container) GetStatsService() StatsService {
	return c.MustGet("stats_service").(StatsService)
}

func (c *Container) GetSettingsService() SettingsService {
	return c.MustGet("settings_service").(SettingsService)
}
