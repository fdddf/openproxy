package dto

import (
	"time"

	"github.com/fdddf/openproxy/internal/models"
)

// Provider describes the provider details exposed to the frontend.
type Provider struct {
	ID           uint       `json:"id"`
	Name         string     `json:"name"`
	Type         string     `json:"type"`
	Key          string     `json:"key"`
	BaseURL      string     `json:"baseUrl"`
	ProxyURL     string     `json:"proxyUrl,omitempty"`
	IsActive     bool       `json:"isActive"`
	HealthCheckEnabled bool   `json:"healthCheckEnabled"`
	HealthCheckStatus  string `json:"healthCheckStatus"`
	HealthCheckChecked *time.Time `json:"healthCheckChecked"`
	CreatedAt    time.Time  `json:"createdAt"`
	UpdatedAt    time.Time  `json:"updatedAt"`
	Models       []Model    `json:"models,omitempty"`
	ClientID     string     `json:"clientId,omitempty"`
	ClientSecret string     `json:"clientSecret,omitempty"`
	AccessToken  string     `json:"accessToken,omitempty"`
	RefreshToken string     `json:"refreshToken,omitempty"`
	TokenExpiry  *time.Time `json:"tokenExpiry,omitempty"`
	AuthURL      string     `json:"authUrl,omitempty"`
	TokenURL     string     `json:"tokenUrl,omitempty"`
	RedirectURL  string     `json:"redirectUrl,omitempty"`
	Scopes       string     `json:"scopes,omitempty"`
	AccountID    string     `json:"accountId,omitempty"`
}

// NewProvider builds a Provider DTO and optionally attaches its models.
func NewProvider(provider *models.Provider, models []*models.Model) Provider {
	dto := Provider{
		ID:           provider.ID,
		Name:         provider.Name,
		Type:         string(provider.Platform),
		Key:          provider.ApiKey,
		BaseURL:      provider.BaseURL,
		ProxyURL:     provider.ProxyURL,
		IsActive:     provider.IsActive,
		HealthCheckEnabled: provider.HealthCheckEnabled,
		HealthCheckStatus:  provider.HealthCheckStatus,
		HealthCheckChecked: provider.HealthCheckChecked,
		CreatedAt:    provider.CreatedAt,
		UpdatedAt:    provider.UpdatedAt,
		ClientID:     provider.ClientID,
		ClientSecret: provider.ClientSecret,
		AccessToken:  provider.AccessToken,
		RefreshToken: provider.RefreshToken,
		TokenExpiry:  provider.TokenExpiry,
		AuthURL:      provider.AuthURL,
		TokenURL:     provider.TokenURL,
		RedirectURL:  provider.RedirectURL,
		Scopes:       provider.Scopes,
		AccountID:    provider.AccountID,
	}

	for _, model := range models {
		dto.Models = append(dto.Models, NewModel(model))
	}

	return dto
}
