package services

import (
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

	log.Printf("using config file: %s\n", configPath)
	yamlFile, err := os.ReadFile(configPath)
	if err != nil {
		return fmt.Errorf("read file failed: %v", err)
	}

	if err := yaml.Unmarshal(yamlFile, c.config); err != nil {
		return fmt.Errorf("parse yaml file failed: %v", err)
	}

	// Load configuration from environment variables, overriding values from file
	c.config.LoadFromEnv()

	// Keep the shared config in sync for places that rely on common.Cfg directly
	common.Cfg = *c.config
	return nil
}
