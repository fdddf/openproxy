package pkg

import (
	"errors"
	"strings"

	"github.com/gofiber/fiber/v2"
)

// ExtractApiKey
func ExtractApiKey(ctx *fiber.Ctx) (string, error) {
	apiKey := ctx.Get("Authorization")
	if apiKey == "" {
		return "", errors.New("no api key provided")
	}
	apiKey = strings.TrimPrefix(apiKey, "Bearer ")
	return apiKey, nil
}
