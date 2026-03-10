package dto

import (
	"time"

	"github.com/fdddf/openproxy/internal/models"
)

// User represents the user profile details returned to the frontend.
type User struct {
	ID          uint      `json:"id"`
	Username    string    `json:"username"`
	Email       string    `json:"email"`
	DisplayName string    `json:"displayName"`
	AvatarURL   string    `json:"avatarUrl"`
	Bio         string    `json:"bio"`
	IsSuper     bool      `json:"isSuper"`
	CreatedAt   time.Time `json:"createdAt"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// NewUser builds a DTO from the database model.
func NewUser(user *models.User) User {
	return User{
		ID:          user.ID,
		Username:    user.Username,
		Email:       user.Email,
		DisplayName: user.DisplayName,
		AvatarURL:   user.AvatarURL,
		Bio:         user.Bio,
		IsSuper:     user.Is_Super,
		CreatedAt:   user.CreatedAt,
		UpdatedAt:   user.UpdatedAt,
	}
}
