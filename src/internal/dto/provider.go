package dto

import (
	"strings"
	"time"

	"github.com/fdddf/openproxy/internal/models"
)

// Provider describes the provider details exposed to the frontend.
type Provider struct {
	ID                 uint       `json:"id"`
	Name               string     `json:"name"`
	Type               string     `json:"type"`
	Key                string     `json:"key"`
	BaseURL            string     `json:"baseUrl"`
	ProxyURL           string     `json:"proxyUrl,omitempty"`
	IsActive           bool       `json:"isActive"`
	HealthCheckEnabled bool       `json:"healthCheckEnabled"`
	HealthCheckStatus  string     `json:"healthCheckStatus"`
	HealthCheckChecked *time.Time `json:"healthCheckChecked"`
	CreatedAt          time.Time  `json:"createdAt"`
	UpdatedAt          time.Time  `json:"updatedAt"`
	Models             []Model    `json:"models,omitempty"`
	ClientID           string     `json:"clientId,omitempty"`
	ClientSecret       string     `json:"clientSecret,omitempty"`
	AccessToken        string     `json:"accessToken,omitempty"`
	RefreshToken       string     `json:"refreshToken,omitempty"`
	TokenExpiry        *time.Time `json:"tokenExpiry,omitempty"`
	AuthURL            string     `json:"authUrl,omitempty"`
	TokenURL           string     `json:"tokenUrl,omitempty"`
	RedirectURL        string     `json:"redirectUrl,omitempty"`
	Scopes             string     `json:"scopes,omitempty"`
	AccountID          string     `json:"accountId,omitempty"`
}

// MaskSecret reduces a credential to a recognisable hint. The admin UI needs to
// show that a key is set and which one it is, not its value: these responses
// used to carry provider API keys, client secrets, and OAuth tokens verbatim.
//
// Update handlers compare an incoming value against this to tell "the user left
// the masked field alone" apart from "the user typed a new secret".
func MaskSecret(value string) string {
	if value == "" {
		return ""
	}
	const visible = 4
	if len(value) <= visible {
		return strings.Repeat("*", len(value))
	}
	return strings.Repeat("*", 8) + value[len(value)-visible:]
}

// NewProvider builds a Provider DTO with credentials masked. Nothing outside
// this package should need the plaintext: the proxy reads provider secrets
// straight from the database rather than through a DTO.
func NewProvider(provider *models.Provider, models []*models.Model) Provider {
	dto := newProviderDTO(provider, models)
	dto.Key = MaskSecret(dto.Key)
	dto.ClientSecret = MaskSecret(dto.ClientSecret)
	dto.AccessToken = MaskSecret(dto.AccessToken)
	dto.RefreshToken = MaskSecret(dto.RefreshToken)
	return dto
}

func newProviderDTO(provider *models.Provider, models []*models.Model) Provider {
	dto := Provider{
		ID:                 provider.ID,
		Name:               provider.Name,
		Type:               string(provider.Platform),
		Key:                provider.ApiKey,
		BaseURL:            provider.BaseURL,
		ProxyURL:           provider.ProxyURL,
		IsActive:           provider.IsActive,
		HealthCheckEnabled: provider.HealthCheckEnabled,
		HealthCheckStatus:  provider.HealthCheckStatus,
		HealthCheckChecked: provider.HealthCheckChecked,
		CreatedAt:          provider.CreatedAt,
		UpdatedAt:          provider.UpdatedAt,
		ClientID:           provider.ClientID,
		ClientSecret:       provider.ClientSecret,
		AccessToken:        provider.AccessToken,
		RefreshToken:       provider.RefreshToken,
		TokenExpiry:        provider.TokenExpiry,
		AuthURL:            provider.AuthURL,
		TokenURL:           provider.TokenURL,
		RedirectURL:        provider.RedirectURL,
		Scopes:             provider.Scopes,
		AccountID:          provider.AccountID,
	}

	for _, model := range models {
		dto.Models = append(dto.Models, NewModel(model))
	}

	return dto
}
