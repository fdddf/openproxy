package services

import (
	"errors"
	"fmt"

	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// SecretServiceImpl provides read access to persisted secrets.
type SecretServiceImpl struct {
	dbService DatabaseService
}

// NewSecretService creates a new SecretService instance.
func NewSecretService(dbService DatabaseService) SecretService {
	return &SecretServiceImpl{
		dbService: dbService,
	}
}

// ValidateProxyAPIKey compares the provided apiKey with the stored proxy API key hash.
func (s *SecretServiceImpl) ValidateProxyAPIKey(ctx *fiber.Ctx, apiKey string) (bool, error) {
	if apiKey == "" {
		return false, nil
	}

	dao := s.dbService.GetDAO()
	// query api key and set user_id to ctx
	key, err := dao.
		APIKey.
		Where(
			dao.APIKey.DeletedAt.IsNull(),
			dao.APIKey.Key.Eq(apiKey),
		).
		First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, nil
		}
		return false, fmt.Errorf("load api keys: %w", err)
	}
	ctx.Locals("user_id", key.UserID)
	ctx.Locals("api_key_id", key.ID)

	return true, nil
}
