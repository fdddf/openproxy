package services

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net/mail"
	"strconv"
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

// APIServiceImpl implements the APIService interface
type APIServiceImpl struct {
	dbService     DatabaseService
	configService ConfigService
}

// NewAPIService creates a new APIService instance
func NewAPIService(dbService DatabaseService, configService ConfigService) APIService {
	return &APIServiceImpl{
		dbService:     dbService,
		configService: configService,
	}
}

func (a *APIServiceImpl) HandleLogin(c *fiber.Ctx) error {
	type LoginRequest struct {
		Username string `json:"username"` // todo validate required
		Password string `json:"password"`
	}

	var req LoginRequest
	if err := c.BodyParser(&req); err != nil {
		return common.HandleError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	dao := a.dbService.GetDAO()

	user, err := dao.User.Where(dao.User.Username.Eq(req.Username)).First()

	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create a new user
			hashedPassword, err := bcrypt.GenerateFromPassword([]byte(req.Password), bcrypt.DefaultCost)
			if err != nil {
				return common.HandleError(c, fiber.StatusInternalServerError, "Failed to hash password")
			}
			newUser := models.User{
				Username:     req.Username,
				PasswordHash: string(hashedPassword),
				DisplayName:  req.Username,
			}
			if err := dao.User.Create(&newUser); err != nil {
				return common.HandleError(c, fiber.StatusInternalServerError, "Failed to create user")
			}
			user = &newUser
		} else {
			return common.HandleError(c, fiber.StatusInternalServerError, "Database error")
		}
	} else {
		// Check password
		if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.Password)); err != nil {
			return common.HandleError(c, fiber.StatusUnauthorized, "Invalid credentials")
		}
	}

	// Create token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"username": user.Username,
		"user_id":  user.ID,
		"exp":      time.Now().Add(time.Hour * 72).Unix(),
	})

	signingKey := a.configService.GetConfig().Proxy.JWTSignKey
	t, err := token.SignedString([]byte(signingKey))
	if err != nil {
		return common.HandleError(c, fiber.StatusInternalServerError, "Failed to create token")
	}

	return c.JSON(fiber.Map{
		"token": t,
		"user":  dto.NewUser(user),
	})
}

func (a *APIServiceImpl) HandleResetPassword(c *fiber.Ctx) error {
	type ResetPasswordRequest struct {
		OldPassword string `json:"oldPassword"`
		NewPassword string `json:"newPassword"`
	}

	var req ResetPasswordRequest
	if err := c.BodyParser(&req); err != nil {
		return common.HandleError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if req.OldPassword == "" || req.NewPassword == "" {
		return common.HandleError(c, fiber.StatusBadRequest, "Current password and new password are required")
	}

	userID, ok := c.Locals("user_id").(float64)
	if !ok {
		return common.HandleError(c, fiber.StatusUnauthorized, "User not authenticated")
	}

	dao := a.dbService.GetDAO()
	user, err := dao.User.Where(dao.User.ID.Eq(uint(userID))).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.HandleError(c, fiber.StatusUnauthorized, "Invalid credentials")
		}
		return common.HandleError(c, fiber.StatusInternalServerError, "Database error")
	}

	if err := bcrypt.CompareHashAndPassword([]byte(user.PasswordHash), []byte(req.OldPassword)); err != nil {
		return common.HandleError(c, fiber.StatusUnauthorized, "Invalid credentials")
	}

	newHash, err := bcrypt.GenerateFromPassword([]byte(req.NewPassword), bcrypt.DefaultCost)
	if err != nil {
		return common.HandleError(c, fiber.StatusInternalServerError, "Failed to hash new password")
	}

	user.PasswordHash = string(newHash)
	if err := dao.User.Save(user); err != nil {
		return common.HandleError(c, fiber.StatusInternalServerError, "Failed to update password")
	}

	return c.JSON(fiber.Map{
		"message": "Password reset successfully",
	})
}

func (a *APIServiceImpl) HandleGetAPIKeys(c *fiber.Ctx) error {
	userID := uint(c.Locals("user_id").(float64))
	page, pageSize := common.GetPaginationParams(c)
	offset := (page - 1) * pageSize

	dao := a.dbService.GetDAO()
	apiKeys, total, err := dao.APIKey.
		Where(dao.APIKey.UserID.Eq(userID)).
		Order(dao.APIKey.CreatedAt.Desc()).
		FindByPage(offset, pageSize)
	if err != nil {
		return common.HandleError(c, fiber.StatusInternalServerError, "Failed to get API keys")
	}

	resp := make([]dto.APIKey, 0, len(apiKeys))
	for _, key := range apiKeys {
		resp = append(resp, dto.NewAPIKey(key))
	}

	return c.JSON(common.NewPaginatedResponse(resp, total, page, pageSize))
}

func (a *APIServiceImpl) HandleLogout(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"message": "Logged out successfully",
	})
}

func (a *APIServiceImpl) HandleCreateAPIKey(c *fiber.Ctx) error {
	userID := uint(c.Locals("user_id").(float64))
	dao := a.dbService.GetDAO()

	type CreateAPIKeyRequest struct {
		Description string `json:"description"`
	}

	var req CreateAPIKeyRequest
	if err := c.BodyParser(&req); err != nil {
		return common.HandleError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	apiKey, err := generateRandomKey(32)
	if err != nil {
		return common.HandleError(c, fiber.StatusInternalServerError, "Failed to generate API key")
	}

	newAPIKey := models.APIKey{
		Name:   req.Description,
		Key:    fmt.Sprintf("sk-%s", apiKey),
		UserID: userID,
	}

	if err := dao.APIKey.Create(&newAPIKey); err != nil {
		return common.HandleError(c, fiber.StatusInternalServerError, "Failed to create API key")
	}

	return c.JSON(dto.NewAPIKey(&newAPIKey))
}

func (a *APIServiceImpl) HandleDeleteAPIKey(c *fiber.Ctx) error {
	userID := uint(c.Locals("user_id").(float64))
	keyID, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return common.HandleError(c, fiber.StatusBadRequest, "Invalid API key id")
	}

	dao := a.dbService.GetDAO()
	apiKey, err := dao.APIKey.Where(
		dao.APIKey.ID.Eq(uint(keyID)),
		dao.APIKey.UserID.Eq(userID),
	).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.HandleError(c, fiber.StatusNotFound, "API key not found")
		}
		return common.HandleError(c, fiber.StatusInternalServerError, "Failed to load API key")
	}

	if _, err := dao.APIKey.Delete(apiKey); err != nil {
		return common.HandleError(c, fiber.StatusInternalServerError, "Failed to delete API key")
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func (a *APIServiceImpl) HandleGetUser(c *fiber.Ctx) error {
	user, err := a.userFromContext(c)
	if err != nil {
		return err
	}

	return c.JSON(dto.NewUser(user))
}

func (a *APIServiceImpl) HandleUpdateCurrentUser(c *fiber.Ctx) error {
	user, err := a.userFromContext(c)
	if err != nil {
		return err
	}

	var req userUpdatePayload
	if err := c.BodyParser(&req); err != nil {
		return common.HandleError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	if err := a.applyUserUpdates(c, user, req, false); err != nil {
		return err
	}

	dao := a.dbService.GetDAO()
	if err := dao.User.Save(user); err != nil {
		return common.HandleError(c, fiber.StatusInternalServerError, "Failed to update user")
	}

	return c.JSON(dto.NewUser(user))
}

func (a *APIServiceImpl) HandleListUsers(c *fiber.Ctx) error {
	admin, err := a.requireSuperUser(c)
	if err != nil {
		return err
	}

	page, pageSize := common.GetPaginationParams(c)
	offset := (page - 1) * pageSize

	dao := a.dbService.GetDAO()
	users, total, err := dao.User.Where(dao.User.ID.Neq(admin.ID)).FindByPage(offset, pageSize)
	if err != nil {
		return common.HandleError(c, fiber.StatusInternalServerError, "Failed to load users")
	}

	resp := make([]dto.User, 0, len(users))
	for _, u := range users {
		resp = append(resp, dto.NewUser(u))
	}

	return c.JSON(common.NewPaginatedResponse(resp, total, page, pageSize))
}

func (a *APIServiceImpl) HandleCreateUser(c *fiber.Ctx) error {
	_, err := a.requireSuperUser(c)
	if err != nil {
		return err
	}

	var req createUserRequest
	if err := c.BodyParser(&req); err != nil {
		return common.HandleError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	username := strings.TrimSpace(req.Username)
	email := strings.TrimSpace(req.Email)
	password := strings.TrimSpace(req.Password)

	if username == "" {
		return common.HandleError(c, fiber.StatusBadRequest, "Username is required")
	}
	if password == "" {
		return common.HandleError(c, fiber.StatusBadRequest, "Password is required")
	}
	if email != "" && !isValidEmail(email) {
		return common.HandleError(c, fiber.StatusBadRequest, "Invalid email address")
	}

	if err := a.ensureUniqueUser(c, username, email, 0); err != nil {
		return err
	}

	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return common.HandleError(c, fiber.StatusInternalServerError, "Failed to hash password")
	}

	newUser := models.User{
		Username:     username,
		PasswordHash: string(hashedPassword),
		Email:        email,
		DisplayName:  firstNonEmpty(req.DisplayName, req.Username),
		AvatarURL:    req.AvatarURL,
		Bio:          req.Bio,
		Is_Super:     req.IsSuper,
	}

	dao := a.dbService.GetDAO()
	if err := dao.User.Create(&newUser); err != nil {
		return common.HandleError(c, fiber.StatusInternalServerError, "Failed to create user")
	}

	return c.JSON(dto.NewUser(&newUser))
}

func (a *APIServiceImpl) HandleUpdateUser(c *fiber.Ctx) error {
	targetID, err := parseUserIDParam(c)
	if err != nil {
		return err
	}

	currentUser, err := a.userFromContext(c)
	if err != nil {
		return err
	}

	if !currentUser.Is_Super && currentUser.ID != targetID {
		return common.HandleError(c, fiber.StatusForbidden, "Only administrators can update other users")
	}

	dao := a.dbService.GetDAO()
	targetUser, err := dao.User.Where(dao.User.ID.Eq(targetID)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.HandleError(c, fiber.StatusNotFound, "User not found")
		}
		return common.HandleError(c, fiber.StatusInternalServerError, "Failed to load user")
	}

	var req userUpdatePayload
	if err := c.BodyParser(&req); err != nil {
		return common.HandleError(c, fiber.StatusBadRequest, "Invalid request body")
	}

	allowRoleChange := currentUser.Is_Super
	if err := a.applyUserUpdates(c, targetUser, req, allowRoleChange); err != nil {
		return err
	}

	if err := dao.User.Save(targetUser); err != nil {
		return common.HandleError(c, fiber.StatusInternalServerError, "Failed to update user")
	}

	return c.JSON(dto.NewUser(targetUser))
}

func (a *APIServiceImpl) HandleDeleteUser(c *fiber.Ctx) error {
	admin, err := a.requireSuperUser(c)
	if err != nil {
		return err
	}

	targetID, err := parseUserIDParam(c)
	if err != nil {
		return err
	}

	if admin.ID == targetID {
		return common.HandleError(c, fiber.StatusBadRequest, "You cannot delete your own account")
	}

	dao := a.dbService.GetDAO()
	user, err := dao.User.Where(dao.User.ID.Eq(targetID)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return common.HandleError(c, fiber.StatusNotFound, "User not found")
		}
		return common.HandleError(c, fiber.StatusInternalServerError, "Failed to load user")
	}

	if _, err := dao.User.Delete(user); err != nil {
		return common.HandleError(c, fiber.StatusInternalServerError, "Failed to delete user")
	}

	return c.SendStatus(fiber.StatusNoContent)
}

func generateRandomKey(length int) (string, error) {
	bytes := make([]byte, length)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func (a *APIServiceImpl) HandleGetSettings(c *fiber.Ctx) error {
	// Placeholder implementation
	return c.JSON(fiber.Map{})
}

func (a *APIServiceImpl) HandleUpdateSettings(c *fiber.Ctx) error {
	// Placeholder implementation
	return c.JSON(fiber.Map{})
}

type createUserRequest struct {
	Username    string `json:"username"`
	Email       string `json:"email"`
	Password    string `json:"password"`
	DisplayName string `json:"displayName"`
	AvatarURL   string `json:"avatarUrl"`
	Bio         string `json:"bio"`
	IsSuper     bool   `json:"isSuper"`
}

type userUpdatePayload struct {
	Username    *string `json:"username"`
	Email       *string `json:"email"`
	Password    *string `json:"password"`
	DisplayName *string `json:"displayName"`
	AvatarURL   *string `json:"avatarUrl"`
	Bio         *string `json:"bio"`
	IsSuper     *bool   `json:"isSuper"`
}

func (a *APIServiceImpl) applyUserUpdates(c *fiber.Ctx, user *models.User, payload userUpdatePayload, allowRoleChange bool) error {
	if payload.Username != nil {
		username := strings.TrimSpace(*payload.Username)
		if username == "" {
			return common.HandleError(c, fiber.StatusBadRequest, "Username cannot be empty")
		}
		if err := a.ensureUniqueUser(c, username, "", user.ID); err != nil {
			return err
		}
		user.Username = username
	}

	if payload.Email != nil {
		email := strings.TrimSpace(*payload.Email)
		if email != "" && !isValidEmail(email) {
			return common.HandleError(c, fiber.StatusBadRequest, "Invalid email address")
		}
		if err := a.ensureUniqueUser(c, "", email, user.ID); err != nil {
			return err
		}
		user.Email = email
	}

	if payload.DisplayName != nil {
		user.DisplayName = strings.TrimSpace(*payload.DisplayName)
	}
	if payload.AvatarURL != nil {
		user.AvatarURL = strings.TrimSpace(*payload.AvatarURL)
	}
	if payload.Bio != nil {
		user.Bio = strings.TrimSpace(*payload.Bio)
	}
	if payload.Password != nil {
		password := strings.TrimSpace(*payload.Password)
		if password == "" {
			return common.HandleError(c, fiber.StatusBadRequest, "Password cannot be empty")
		}
		newHash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
		if err != nil {
			return common.HandleError(c, fiber.StatusInternalServerError, "Failed to hash password")
		}
		user.PasswordHash = string(newHash)
	}

	if allowRoleChange && payload.IsSuper != nil {
		user.Is_Super = *payload.IsSuper
	}

	return nil
}

func (a *APIServiceImpl) ensureUniqueUser(c *fiber.Ctx, username, email string, excludeID uint) error {
	db := a.dbService.GetDB()
	var existing models.User

	if username != "" {
		query := db.Where("username = ?", username)
		if excludeID != 0 {
			query = query.Where("id <> ?", excludeID)
		}
		if err := query.First(&existing).Error; err == nil {
			return common.HandleError(c, fiber.StatusBadRequest, "Username already exists")
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return common.HandleError(c, fiber.StatusInternalServerError, "Failed to validate username")
		}
	}

	if email != "" {
		query := db.Where("email = ?", email)
		if excludeID != 0 {
			query = query.Where("id <> ?", excludeID)
		}
		if err := query.First(&existing).Error; err == nil {
			return common.HandleError(c, fiber.StatusBadRequest, "Email already exists")
		} else if !errors.Is(err, gorm.ErrRecordNotFound) {
			return common.HandleError(c, fiber.StatusInternalServerError, "Failed to validate email")
		}
	}

	return nil
}

func parseUserIDParam(c *fiber.Ctx) (uint, error) {
	id, err := strconv.ParseUint(c.Params("id"), 10, 64)
	if err != nil {
		return 0, common.HandleError(c, fiber.StatusBadRequest, "Invalid user id")
	}
	return uint(id), nil
}

func (a *APIServiceImpl) userFromContext(c *fiber.Ctx) (*models.User, error) {
	userID, err := a.userIDFromContext(c)
	if err != nil {
		return nil, err
	}

	dao := a.dbService.GetDAO()
	user, err := dao.User.Where(dao.User.ID.Eq(userID)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, common.HandleError(c, fiber.StatusUnauthorized, "User not found")
		}
		return nil, common.HandleError(c, fiber.StatusInternalServerError, "Failed to load user")
	}

	return user, nil
}

func (a *APIServiceImpl) userIDFromContext(c *fiber.Ctx) (uint, error) {
	rawID := c.Locals("user_id")
	if rawID == nil {
		return 0, common.HandleError(c, fiber.StatusUnauthorized, "User not authenticated")
	}

	idFloat, ok := rawID.(float64)
	if !ok {
		return 0, common.HandleError(c, fiber.StatusUnauthorized, "Invalid user context")
	}
	return uint(idFloat), nil
}

func (a *APIServiceImpl) requireSuperUser(c *fiber.Ctx) (*models.User, error) {
	user, err := a.userFromContext(c)
	if err != nil {
		return nil, err
	}
	if !user.Is_Super {
		return nil, common.HandleError(c, fiber.StatusForbidden, "Admin access required")
	}
	return user, nil
}

func isValidEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

func firstNonEmpty(values ...string) string {
	for _, v := range values {
		if strings.TrimSpace(v) != "" {
			return strings.TrimSpace(v)
		}
	}
	return ""
}

// Additional methods for API key quota management

// CheckAPIKeyQuota checks if the API key has exceeded its quota
func (a *APIServiceImpl) CheckAPIKeyQuota(ctx *fiber.Ctx, apiKey string) (bool, error) {
	db := a.dbService.GetDB()
	var apiKeyModel models.APIKey
	if err := db.Where("key = ?", apiKey).First(&apiKeyModel).Error; err != nil {
		if err == gorm.ErrRecordNotFound {
			return false, fmt.Errorf("api key not found")
		}
		return false, err
	}

	// If there's no quota set, the key is unlimited
	if apiKeyModel.Quota == nil {
		return true, nil
	}

	// Check if we need to reset the quota
	now := time.Now()
	if apiKeyModel.ResetTime != nil && now.After(*apiKeyModel.ResetTime) {
		// Reset the usage counter
		apiKeyModel.Used = 0
		// Set next reset time to next month
		nextReset := now.AddDate(0, 1, 0)
		apiKeyModel.ResetTime = &nextReset

		if err := db.Model(&apiKeyModel).Updates(map[string]interface{}{
			"used":       apiKeyModel.Used,
			"reset_time": nextReset,
		}).Error; err != nil {
			return false, err
		}
	}

	// Check if quota is exceeded
	if apiKeyModel.Used >= *apiKeyModel.Quota {
		return false, nil // Quota exceeded
	}

	return true, nil
}

// UpdateAPIKeyUsage increments the usage counter for an API key
func (a *APIServiceImpl) UpdateAPIKeyUsage(apiKey string) error {
	db := a.dbService.GetDB()
	var apiKeyModel models.APIKey
	if err := db.Where("key = ?", apiKey).First(&apiKeyModel).Error; err != nil {
		return err
	}

	// Update the usage counter
	if err := db.Model(&apiKeyModel).UpdateColumn("used", gorm.Expr("used + ?", 1)).Error; err != nil {
		return err
	}

	return nil
}

// GetAPIKeyInfo retrieves information about an API key
func (a *APIServiceImpl) GetAPIKeyInfo(apiKey string) (*models.APIKey, error) {
	db := a.dbService.GetDB()
	var apiKeyModel models.APIKey
	if err := db.Where("key = ?", apiKey).First(&apiKeyModel).Error; err != nil {
		return nil, err
	}
	return &apiKeyModel, nil
}

// UpdateAPIKeyQuota updates the quota for an API key
func (a *APIServiceImpl) UpdateAPIKeyQuota(apiKey string, quota *int64) error {
	db := a.dbService.GetDB()
	var apiKeyModel models.APIKey
	if err := db.Where("key = ?", apiKey).First(&apiKeyModel).Error; err != nil {
		return err
	}

	updates := map[string]interface{}{
		"quota": quota,
	}

	// If setting a new quota, reset the usage counter
	if quota != nil {
		updates["used"] = 0
		// Set reset time to next month
		now := time.Now()
		nextReset := now.AddDate(0, 1, 0)
		updates["reset_time"] = nextReset
	}

	if err := db.Model(&apiKeyModel).Updates(updates).Error; err != nil {
		return err
	}

	return nil
}
