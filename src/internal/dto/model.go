package dto

import (
	"time"

	"github.com/fdddf/openproxy/internal/models"
)

// Model represents a model mapping visible to the frontend.
type Model struct {
	ID         uint      `json:"id"`
	ProviderID uint      `json:"providerId"`
	Name       string    `json:"name"`
	MappedName string    `json:"mappedName"`
	IsActive   bool      `json:"isActive"`
	CreatedAt  time.Time `json:"createdAt"`
	UpdatedAt  time.Time `json:"updatedAt"`
	UsageCount int       `json:"usageCount"`
}

// NewModel builds a Model DTO from the database model.
func NewModel(model *models.Model) Model {
	return Model{
		ID:         model.ID,
		ProviderID: model.ProviderID,
		// The provider's real model name is considered the "original" name,
		// while the stored Name is the alias used by clients.
		Name:       model.RealModel,
		MappedName: model.Name,
		IsActive:   model.IsActive,
		CreatedAt:  model.CreatedAt,
		UpdatedAt:  model.UpdatedAt,
	}
}
