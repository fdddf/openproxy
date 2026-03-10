package dto

// SystemSettings represents the settings exposed to the frontend.
type SystemSettings struct {
	DefaultProvider    string `json:"defaultProvider"`
	RateLimit          int    `json:"rateLimit"`
	ConcurrentLimit    int    `json:"concurrentLimit"`
	LogLevel           string `json:"logLevel"`
	CacheTTL           int    `json:"cacheTTL"`
	MaintenanceMode    bool   `json:"maintenanceMode"`
	HealthCheckEnabled bool   `json:"healthCheckEnabled"`
}

// DefaultSystemSettings returns the baseline settings shown when nothing is stored.
func DefaultSystemSettings(defaultProvider string) SystemSettings {
	return SystemSettings{
		DefaultProvider:    defaultProvider,
		RateLimit:          100,
		ConcurrentLimit:    10,
		LogLevel:           "info",
		CacheTTL:           3600,
		MaintenanceMode:    false,
		HealthCheckEnabled: false,
	}
}
