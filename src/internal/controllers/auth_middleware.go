package controllers

import (
	"strings"

	"github.com/fdddf/openproxy/common"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

// NewAuthMiddleware validates JWT tokens issued by the API service.
func NewAuthMiddleware(signingKey string) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return common.HandleError(c, fiber.StatusUnauthorized, "Missing authorization header")
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		if tokenString == authHeader {
			return common.HandleError(c, fiber.StatusUnauthorized, "Invalid authorization header")
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fiber.NewError(fiber.StatusUnauthorized, "Unexpected signing method")
			}
			return []byte(signingKey), nil
		})

		if err != nil {
			return common.HandleError(c, fiber.StatusUnauthorized, "Invalid token")
		}

		if claims, ok := token.Claims.(jwt.MapClaims); ok && token.Valid {
			c.Locals("user_id", claims["user_id"])
			return c.Next()
		}

		return common.HandleError(c, fiber.StatusUnauthorized, "Invalid token")
	}
}
