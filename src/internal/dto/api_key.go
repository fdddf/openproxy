package dto

import (
	"time"

	"github.com/fdddf/openproxy/internal/models"
)

// APIKey represents the fields the frontend needs for API key management.
type APIKey struct {
	ID          uint       `json:"id"`
	Key         string     `json:"key"`
	Description string     `json:"description"`
	IsActive    bool       `json:"isActive"`
	UsageCount  int        `json:"usageCount"`
	MaxUsage    *int       `json:"maxUsage,omitempty"`
	ExpiresAt   *time.Time `json:"expiresAt,omitempty"`
	CreatedAt   time.Time  `json:"createdAt"`
	UpdatedAt   time.Time  `json:"updatedAt"`
}

// NewAPIKey builds an APIKey DTO from the database model.
func NewAPIKey(apiKey *models.APIKey) APIKey {
	return APIKey{
		ID:          apiKey.ID,
		Key:         apiKey.Key,
		Description: apiKey.Name,
		IsActive:    true,
		UsageCount:  0,
		MaxUsage:    nil,
		ExpiresAt:   nil,
		CreatedAt:   apiKey.CreatedAt,
		UpdatedAt:   apiKey.UpdatedAt,
	}
}
