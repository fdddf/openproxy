package controllers

import (
	"errors"

	"github.com/fdddf/openproxy/common"
	"github.com/fdddf/openproxy/internal/services"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// NewAdminMiddleware restricts a route to super users.
//
// Only the user-management handlers used to check this, so any authenticated
// account could read provider credentials, rewrite global settings, and browse
// every user's request log. It runs after NewAuthMiddleware, which puts the
// caller's id on the context.
func NewAdminMiddleware(dbService services.DatabaseService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		userID, ok := c.Locals("user_id").(float64)
		if !ok {
			return common.HandleError(c, fiber.StatusUnauthorized, "User not authenticated")
		}

		dao := dbService.GetDAO()
		user, err := dao.User.Where(dao.User.ID.Eq(uint(userID))).First()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				return common.HandleError(c, fiber.StatusUnauthorized, "User not found")
			}
			return common.HandleError(c, fiber.StatusInternalServerError, "Database error")
		}

		if !user.Is_Super {
			return common.HandleError(c, fiber.StatusForbidden, "Administrator privileges required")
		}

		return c.Next()
	}
}
