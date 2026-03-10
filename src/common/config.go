package common

import (
	"fmt"
	"os"
	"strings"
	"time"
)

type ProviderConfig struct {
	Type     string            `yaml:"type"`
	ApiKey   string            `yaml:"api_key"`
	BaseURL  string            `yaml:"base_url"`
	ModelMap map[string]string `yaml:"model_map"`
	ProxyURL string            `yaml:"proxy_url"`

	// OAuth2 specific fields
	ClientID     string    `yaml:"client_id,omitempty"`
	ClientSecret string    `yaml:"client_secret,omitempty"`
	AccessToken  string    `yaml:"access_token,omitempty"`
	RefreshToken string    `yaml:"refresh_token,omitempty"`
	TokenExpiry  time.Time `yaml:"token_expiry,omitempty"`
	AuthURL      string    `yaml:"auth_url,omitempty"`
	TokenURL     string    `yaml:"token_url,omitempty"`
	RedirectURL  string    `yaml:"redirect_url,omitempty"`
	Scopes       string    `yaml:"scopes,omitempty"`

	// Codex specific fields
	AccountID string `yaml:"account_id,omitempty"`
}

// SplitScopes splits the scopes string by common delimiters.
func SplitScopes(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return nil
	}

	fields := strings.FieldsFunc(raw, func(r rune) bool {
		return r == ',' || r == ' ' || r == ';'
	})

	scopes := make([]string, 0, len(fields))
	for _, scope := range fields {
		if trimmed := strings.TrimSpace(scope); trimmed != "" {
			scopes = append(scopes, trimmed)
		}
	}
	return scopes
}

type DatabaseConfig struct {
	Host           string `yaml:"host"`
	Port           int    `yaml:"port"`
	User           string `yaml:"user"`
	Password       string `yaml:"password"`
	Name           string `yaml:"name"`
	SSLMode        string `yaml:"ssl_mode"`
	MigrationsPath string `yaml:"migrations_path"`
}

// 配置结构体
type Config struct {
	Proxy struct {
		ListenAddress   string `yaml:"listen_address"`
		DefaultProvider string `yaml:"default_provider"`
		JWTSignKey      string `yaml:"jwt_sign_key"`
	} `yaml:"proxy"`

	Database DatabaseConfig `yaml:"database"`
}

// LoadFromEnv loads configuration from environment variables, with fallbacks to the original config
func (c *Config) LoadFromEnv() {
	// Load JWT Sign Key from environment or fallback to config
	if jwtKey := os.Getenv("JWT_SIGN_KEY"); jwtKey != "" {
		c.Proxy.JWTSignKey = jwtKey
	}

	// Load database config from environment variables
	if dbHost := os.Getenv("DB_HOST"); dbHost != "" {
		c.Database.Host = dbHost
	}
	if dbPort := os.Getenv("DB_PORT"); dbPort != "" {
		if port, err := parseEnvInt(dbPort); err == nil {
			c.Database.Port = port
		}
	}
	if dbUser := os.Getenv("DB_USER"); dbUser != "" {
		c.Database.User = dbUser
	}
	if dbPassword := os.Getenv("DB_PASSWORD"); dbPassword != "" {
		c.Database.Password = dbPassword
	}
	if dbName := os.Getenv("DB_NAME"); dbName != "" {
		c.Database.Name = dbName
	}
	if dbSSLMode := os.Getenv("DB_SSL_MODE"); dbSSLMode != "" {
		c.Database.SSLMode = dbSSLMode
	}
	if migrationsPath := os.Getenv("DB_MIGRATIONS_PATH"); migrationsPath != "" {
		c.Database.MigrationsPath = migrationsPath
	}
}

// parseEnvInt parses an environment variable string to int with error handling
func parseEnvInt(s string) (int, error) {
	var result int
	_, err := fmt.Sscanf(s, "%d", &result)
	return result, err
}

var Cfg Config
