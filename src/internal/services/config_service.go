package services

import (
	"errors"
	"fmt"
	"log"
	"os"

	"github.com/fdddf/openproxy/common"
	"gopkg.in/yaml.v3"
)

// ConfigServiceImpl implements the ConfigService interface
type ConfigServiceImpl struct {
	config *common.Config
}

// NewConfigService creates a new ConfigService instance
func NewConfigService() ConfigService {
	return &ConfigServiceImpl{
		config: &common.Config{},
	}
}

// GetConfig returns the current configuration
func (c *ConfigServiceImpl) GetConfig() *common.Config {
	return c.config
}

// LoadConfig loads the configuration from file and environment variables
func (c *ConfigServiceImpl) LoadConfig(configPath string) error {
	if configPath == "" {
		configPath = "config.yaml"
	}

	yamlFile, err := os.ReadFile(configPath)
	switch {
	case err == nil:
		log.Printf("using config file: %s", configPath)
		if err := yaml.Unmarshal(yamlFile, c.config); err != nil {
			return fmt.Errorf("parse yaml file failed: %v", err)
		}
	case errors.Is(err, os.ErrNotExist):
		// A missing config file is not an error: the defaults are enough to run
		// a self-contained SQLite instance.
		log.Printf("no config file at %s, using defaults", configPath)
	default:
		return fmt.Errorf("read file failed: %v", err)
	}

	// Load configuration from environment variables, overriding values from file
	c.config.LoadFromEnv()
	c.config.ApplyDefaults()

	// Keep the shared config in sync for places that rely on common.Cfg directly
	common.Cfg = *c.config
	return nil
}
