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

// Database drivers supported by the server.
const (
	DriverSQLite   = "sqlite"
	DriverPostgres = "postgres"
)

// DefaultSQLitePath is where the database file lands when nothing is configured.
const DefaultSQLitePath = "./openproxy.db"

type DatabaseConfig struct {
	// Driver selects the backend: "sqlite" (default) or "postgres".
	Driver string `yaml:"driver"`

	// Path is the SQLite database file. Only used when Driver is "sqlite".
	Path string `yaml:"path"`

	// The remaining fields only apply to Postgres.
	Host     string `yaml:"host"`
	Port     int    `yaml:"port"`
	User     string `yaml:"user"`
	Password string `yaml:"password"`
	Name     string `yaml:"name"`
	SSLMode  string `yaml:"ssl_mode"`

	// MigrationsPath overrides the embedded migration set with a directory on
	// disk. Leave empty to use the migrations compiled into the binary.
	MigrationsPath string `yaml:"migrations_path"`
}

// RequestLogConfig controls how much of each proxied request is persisted.
// The defaults matter most on SQLite, where unbounded bodies grow the file fast.
type RequestLogConfig struct {
	// Enabled turns request logging on. Default true.
	Enabled *bool `yaml:"enabled"`
	// MaxBodyBytes truncates stored request/response bodies. 0 disables the cap.
	MaxBodyBytes int `yaml:"max_body_bytes"`
	// RetentionDays deletes records older than this. 0 keeps everything.
	RetentionDays int `yaml:"retention_days"`
}

// LoggingEnabled reports whether request logging is on, defaulting to true.
func (r RequestLogConfig) LoggingEnabled() bool {
	return r.Enabled == nil || *r.Enabled
}

// 配置结构体
type Config struct {
	Proxy struct {
		ListenAddress   string `yaml:"listen_address"`
		DefaultProvider string `yaml:"default_provider"`
		JWTSignKey      string `yaml:"jwt_sign_key"`
	} `yaml:"proxy"`

	Database   DatabaseConfig   `yaml:"database"`
	RequestLog RequestLogConfig `yaml:"request_log"`
}

// ApplyDefaults fills in the values that let the binary start with no config
// file at all: SQLite in the working directory, listening on :8081.
func (c *Config) ApplyDefaults() {
	if c.Proxy.ListenAddress == "" {
		c.Proxy.ListenAddress = ":8081"
	}
	if c.Database.Driver == "" {
		// An existing Postgres config without an explicit driver keeps working.
		if c.Database.Host != "" || c.Database.Name != "" {
			c.Database.Driver = DriverPostgres
		} else {
			c.Database.Driver = DriverSQLite
		}
	}
	c.Database.Driver = strings.ToLower(strings.TrimSpace(c.Database.Driver))
	if c.Database.Driver == DriverSQLite && c.Database.Path == "" {
		c.Database.Path = DefaultSQLitePath
	}
	if c.RequestLog.MaxBodyBytes == 0 {
		c.RequestLog.MaxBodyBytes = 64 * 1024
	}
	if c.RequestLog.RetentionDays == 0 {
		c.RequestLog.RetentionDays = 30
	}
}

// LoadFromEnv loads configuration from environment variables, with fallbacks to the original config
func (c *Config) LoadFromEnv() {
	// Load JWT Sign Key from environment or fallback to config
	if jwtKey := os.Getenv("JWT_SIGN_KEY"); jwtKey != "" {
		c.Proxy.JWTSignKey = jwtKey
	}

	// Load database config from environment variables
	if driver := os.Getenv("DB_DRIVER"); driver != "" {
		c.Database.Driver = driver
	}
	if dbPath := os.Getenv("DB_PATH"); dbPath != "" {
		c.Database.Path = dbPath
	}
	if addr := os.Getenv("LISTEN_ADDRESS"); addr != "" {
		c.Proxy.ListenAddress = addr
	}
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
