package models

import (
	"time"

	"github.com/fdddf/openproxy/consts"
)

// Provider model (for provider configuration in database)
type Provider struct {
	CommonFields

	Name     string                  `json:"name"`
	Platform consts.ProviderPlatform `json:"platform"`
	ApiKey   string                  `json:"api_key"`
	BaseURL  string                  `json:"base_url"`
	ProxyURL string                  `json:"proxy_url"`
	IsActive bool                    `json:"is_active"`

	// OAuth2 specific fields
	ClientID     string     `json:"client_id,omitempty"`
	ClientSecret string     `json:"client_secret,omitempty"`
	AccessToken  string     `json:"access_token,omitempty"`
	RefreshToken string     `json:"refresh_token,omitempty"`
	TokenExpiry  *time.Time `json:"token_expiry,omitempty"`
	AuthURL      string     `json:"auth_url,omitempty"`
	TokenURL     string     `json:"token_url,omitempty"`
	RedirectURL  string     `json:"redirect_url,omitempty"`
	Scopes       string     `json:"scopes,omitempty"`
	AccountID    string     `json:"account_id,omitempty"`

	HealthCheckEnabled bool `json:"health_check_enabled" gorm:"default:false"`
	HealthCheckStatus  string     `json:"health_check_status" gorm:"default:unknown"`
	HealthCheckChecked *time.Time `json:"health_check_checked"`
}
