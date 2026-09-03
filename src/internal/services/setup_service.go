package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/mail"
	"strings"
	"time"

	"github.com/fdddf/openproxy/common"
	"github.com/fdddf/openproxy/internal/dto"
	"github.com/fdddf/openproxy/internal/models"
	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"gorm.io/gorm"
)

// settingKeyJWTSignKey stores the generated signing key when none is configured.
const settingKeyJWTSignKey = "jwt_sign_key"

// SetupService drives first-run initialization: it decides whether the instance
// still needs an administrator and creates that first account.
//
// Replacing the old seeded admin/admin123 row, which shipped identical
// credentials to every install.
type SetupService interface {
	// NeedsSetup reports whether the instance has no users yet.
	NeedsSetup() (bool, error)
	// EnsureJWTSignKey returns the configured signing key, generating and
	// persisting a random one when the config leaves it empty.
	EnsureJWTSignKey() (string, error)
	HandleGetSetupStatus(c *fiber.Ctx) error
	HandleSetup(c *fiber.Ctx) error
}

type setupService struct {
	dbService     DatabaseService
	configService ConfigService
}

// NewSetupService creates the first-run setup service.
func NewSetupService(dbService DatabaseService, configService ConfigService) SetupService {
	return &setupService{
		dbService:     dbService,
		configService: configService,
	}
}

func (s *setupService) NeedsSetup() (bool, error) {
	dao := s.dbService.GetDAO()
	count, err := dao.User.Count()
	if err != nil {
		return false, fmt.Errorf("count users: %w", err)
	}
	return count == 0, nil
}

func (s *setupService) EnsureJWTSignKey() (string, error) {
	cfg := s.configService.GetConfig()
	if key := strings.TrimSpace(cfg.Proxy.JWTSignKey); key != "" {
		return key, nil
	}

	dao := s.dbService.GetDAO()
	stored, err := dao.Setting.Where(dao.Setting.Key.Eq(settingKeyJWTSignKey)).First()
	if err == nil && stored.Value != "" {
		cfg.Proxy.JWTSignKey = stored.Value
		common.Cfg.Proxy.JWTSignKey = stored.Value
		return stored.Value, nil
	}
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return "", fmt.Errorf("load jwt sign key: %w", err)
	}

	// An empty signing key means anyone can forge a token, so generate one and
	// persist it: regenerating per start would invalidate every session.
	buf := make([]byte, 32)
	if _, err := rand.Read(buf); err != nil {
		return "", fmt.Errorf("generate jwt sign key: %w", err)
	}
	key := hex.EncodeToString(buf)

	if err := dao.Setting.Create(&models.Setting{Key: settingKeyJWTSignKey, Value: key}); err != nil {
		return "", fmt.Errorf("persist jwt sign key: %w", err)
	}

	cfg.Proxy.JWTSignKey = key
	common.Cfg.Proxy.JWTSignKey = key
	return key, nil
}

func (s *setupService) HandleGetSetupStatus(c *fiber.Ctx) error {
	needsSetup, err := s.NeedsSetup()
	if err != nil {
		return common.HandleError(c, fiber.StatusInternalServerError, "Failed to read setup status")
	}
	return c.JSON(fiber.Map{"needsSetup": needsSetup})
}

func (s *setupService) HandleSetup(c *fiber.Ctx) error {
	type setupRequest struct {
		Username    string `json:"username"`
		Password    string `json:"password"`
		Email       string `json:"email"`
		DisplayName string `json:"displayName"`
	}

	var req setupRequest
	if err := c.BodyParser(&req); err != nil {
		return common.HandleError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	username := strings.TrimSpace(req.Username)
	email := strings.TrimSpace(req.Email)

	if username == "" {
		return common.HandleError(c, fiber.StatusBadRequest, "Username is required")
	}
	if len(req.Password) < 8 {
		return common.HandleError(c, fiber.StatusBadRequest, "Password must be at least 8 characters")
	}
	if email != "" {
		if _, err := mail.ParseAddress(email); err != nil {
			return common.HandleError(c, fiber.StatusBadRequest, "Invalid email address")
		}
	}

	dao := s.dbService.GetDAO()

	// Guarded by the user count rather than by auth: this endpoint is reachable
	// without a token, so it must refuse to run once an admin exists.
	needsSetup, err := s.NeedsSetup()
	if err != nil {
		return common.HandleError(c, fiber.StatusInternalServerError, "Failed to read setup status")
	}
	if !needsSetup {
		return common.HandleError(c, fiber.StatusConflict, "This instance is already initialized")
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
	if err != nil {
		return common.HandleError(c, fiber.StatusInternalServerError, "Failed to hash password")
	}

	admin := models.User{
		Username:     username,
		PasswordHash: string(hashedPassword),
		Email:        email,
		DisplayName:  firstNonEmpty(req.DisplayName, username),
		Is_Super:     true,
	}
	if err := dao.User.Create(&admin); err != nil {
		// A concurrent setup request may have won the race and taken the name.
		return common.HandleError(c, fiber.StatusConflict, "Failed to create administrator")
	}

	signingKey, err := s.EnsureJWTSignKey()
	if err != nil {
		return common.HandleError(c, fiber.StatusInternalServerError, "Failed to load signing key")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": admin.Username,
		"user_id":  admin.ID,
		"exp":      time.Now().Add(time.Hour * 72).Unix(),
	})
	signed, err := token.SignedString([]byte(signingKey))
	if err != nil {
		return common.HandleError(c, fiber.StatusInternalServerError, "Failed to create token")
	}

	return c.JSON(fiber.Map{
		"token": signed,
		"user":  dto.NewUser(&admin),
	})
}
