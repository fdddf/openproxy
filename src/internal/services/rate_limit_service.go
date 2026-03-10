package services

import (
	"errors"
	"fmt"
	"strconv"
	"sync"
	"time"

	"github.com/fdddf/openproxy/internal/models"
	"github.com/gofiber/fiber/v2"
	"gorm.io/gorm"
)

// RateLimit represents rate limiting configuration for an API key or user
type RateLimit struct {
	ID        uint      `json:"id"`
	APIKeyID  *uint     `json:"api_key_id,omitempty"` // Optional: associate with specific API key
	UserID    *uint     `json:"user_id,omitempty"`    // Optional: associate with specific user
	Limit     int       `json:"limit"`                // Requests per time window
	Window    int       `json:"window"`               // Time window in seconds
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}

// RateLimitUsage tracks current usage for rate limiting
type RateLimitUsage struct {
	ID          uint      `json:"id"`
	RateLimitID uint      `json:"rate_limit_id"`
	Count       int       `json:"count"`
	ResetTime   time.Time `json:"reset_time"`
	CreatedAt   time.Time `json:"created_at"`
	UpdatedAt   time.Time `json:"updated_at"`
}

// RateLimitService interface defines the rate limiting service operations
type RateLimitService interface {
	CheckRateLimit(ctx *fiber.Ctx, apiKey string) (bool, error)
	GetRateLimitForAPIKey(apiKey string) (*RateLimit, error)
	SetRateLimitForAPIKey(apiKey string, limit int, window int) error
	GetRateLimitForUser(userID uint) (*RateLimit, error)
	SetRateLimitForUser(userID uint, limit int, window int) error
}

// RateLimitServiceImpl implements the RateLimitService interface
type RateLimitServiceImpl struct {
	dbService       DatabaseService
	settingsService SettingsService
	mu              sync.Mutex // for in-memory rate limiting
	// In-memory store for rate limit usage (would be replaced with Redis in production)
	inMemoryUsage map[string]*RateLimitUsage
	// Track if we have any individual rate limit settings in the system to avoid repeated database queries
	initialized   bool
	hasRateLimits bool
}

// NewRateLimitService creates a new RateLimitService instance
func NewRateLimitService(dbService DatabaseService, settingsService SettingsService) RateLimitService {
	return &RateLimitServiceImpl{
		dbService:       dbService,
		settingsService: settingsService,
		inMemoryUsage:   make(map[string]*RateLimitUsage),
	}
}

// CheckRateLimit checks if a request from the given API key is within rate limits
func (r *RateLimitServiceImpl) CheckRateLimit(ctx *fiber.Ctx, apiKey string) (bool, error) {
	// Get effective settings to check if rate limiting is enabled globally
	settings, err := r.settingsService.GetEffectiveSettings()
	if err != nil {
		// On settings error, return true (allow request) to prevent blocking due to configuration errors
		// Log the error if possible (but continue without individual checks)
		return true, nil
	}

	// If the global rate limit setting is 0 or negative, bypass rate limiting
	if settings.RateLimit <= 0 {
		return true, nil
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if we have already initialized our knowledge about whether individual rate limits exist
	// This prevents repeated database queries to check for system-wide rate limit settings
	if !r.initialized {
		dao := r.dbService.GetDAO()
		_, err := dao.Setting.Where(dao.Setting.Key.Like("rate_limit_%")).First()
		if err != nil {
			if errors.Is(err, gorm.ErrRecordNotFound) {
				// No rate limit settings found in the system
				r.hasRateLimits = false
			} else {
				// Database error - return true to not block requests
				return true, nil
			}
		} else {
			// At least one rate limit setting exists
			r.hasRateLimits = true
		}
		r.initialized = true
	}

	// If no individual rate limits are configured in the system, bypass them
	if !r.hasRateLimits {
		return true, nil
	}

	// Get DAO instance
	dao := r.dbService.GetDAO()

	// First, get the API key from the database to get user ID
	apiKeyModel, err := dao.APIKey.Where(dao.APIKey.Key.Eq(apiKey)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return false, fmt.Errorf("api key not found")
		}
		return false, err
	}

	// If we found at least one rate limit setting, check for specific limits for this API key and user
	apiKeyLimit, err := r.getRateLimit("api_key", apiKeyModel.ID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return false, err
	}

	// Check rate limit for user (associated with the API key)
	userLimit, err := r.getRateLimit("user", apiKeyModel.UserID)
	if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
		return false, err
	}

	// Use the more restrictive limit (if both exist)
	var effectiveLimit *RateLimit
	if apiKeyLimit != nil && userLimit != nil {
		// If both limits exist, use the one with the smaller limit
		if apiKeyLimit.Limit <= userLimit.Limit && apiKeyLimit.Window <= userLimit.Window {
			effectiveLimit = apiKeyLimit
		} else {
			effectiveLimit = userLimit
		}
	} else if apiKeyLimit != nil {
		effectiveLimit = apiKeyLimit
	} else if userLimit != nil {
		effectiveLimit = userLimit
	} else {
		// If no specific limits are set for this API key/user, allow the request
		return true, nil
	}

	// Generate a unique key for this rate limit instance
	usageKey := fmt.Sprintf("rate_limit_%d", effectiveLimit.ID)

	now := time.Now()
	usage, exists := r.inMemoryUsage[usageKey]

	// Check if we need to reset the counter (time window has passed)
	if !exists || now.After(usage.ResetTime) {
		// Create a new counter with the current time window
		resetTime := now.Add(time.Duration(effectiveLimit.Window) * time.Second)
		r.inMemoryUsage[usageKey] = &RateLimitUsage{
			ID:          0,
			RateLimitID: effectiveLimit.ID,
			Count:       1,
			ResetTime:   resetTime,
		}
		return true, nil
	}

	// Check if the limit has been exceeded
	if usage.Count >= effectiveLimit.Limit {
		return false, nil
	}

	// Increment the counter
	usage.Count++
	r.inMemoryUsage[usageKey] = usage
	return true, nil
}

// getRateLimit is a helper function to retrieve rate limits by type and ID
func (r *RateLimitServiceImpl) getRateLimit(entityType string, entityID uint) (*RateLimit, error) {
	// In a real implementation, we would have a RateLimit table in the database
	// For now, we'll look for a setting that might have rate limit information
	var settingKey string

	if entityType == "api_key" {
		settingKey = fmt.Sprintf("rate_limit_api_key_%d", entityID)
	} else if entityType == "user" {
		settingKey = fmt.Sprintf("rate_limit_user_%d", entityID)
	} else {
		return nil, fmt.Errorf("invalid entity type: %s", entityType)
	}

	// Get DAO instance
	dao := r.dbService.GetDAO()

	setting, err := dao.Setting.Where(dao.Setting.Key.Eq(settingKey)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, err // Return the record not found error
		}
		return nil, err
	}

	// Parse the limit and window from the setting value
	// Format: "limit:window" e.g., "100:3600" for 100 requests per hour
	parts := split(setting.Value, ":")
	if len(parts) != 2 {
		return nil, fmt.Errorf("invalid rate limit format: %s", setting.Value)
	}

	limit, err := strconv.Atoi(parts[0])
	if err != nil {
		return nil, fmt.Errorf("invalid rate limit: %s", parts[0])
	}

	window, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, fmt.Errorf("invalid time window: %s", parts[1])
	}

	return &RateLimit{
		ID:     entityID,
		Limit:  limit,
		Window: window,
	}, nil
}

// split is a helper function similar to strings.Split but with a limit
func split(s, sep string) []string {
	var parts []string
	last := 0
	for i := 0; i < len(s); i++ {
		if s[i] == sep[0] && (len(sep) == 1 || i+len(sep) <= len(s) && s[i:i+len(sep)] == sep) {
			parts = append(parts, s[last:i])
			i += len(sep) - 1
			last = i + 1
		}
	}
	parts = append(parts, s[last:])
	return parts
}

// GetRateLimitForAPIKey retrieves rate limiting configuration for a specific API key
func (r *RateLimitServiceImpl) GetRateLimitForAPIKey(apiKey string) (*RateLimit, error) {
	// Get DAO instance
	dao := r.dbService.GetDAO()

	// First get the API key model to retrieve its ID
	apiKeyModel, err := dao.APIKey.Where(dao.APIKey.Key.Eq(apiKey)).First()
	if err != nil {
		return nil, err
	}

	return r.getRateLimit("api_key", apiKeyModel.ID)
}

// SetRateLimitForAPIKey sets rate limiting configuration for a specific API key
func (r *RateLimitServiceImpl) SetRateLimitForAPIKey(apiKey string, limit int, window int) error {
	// Get DAO instance
	dao := r.dbService.GetDAO()

	// First get the API key model to retrieve its ID
	apiKeyModel, err := dao.APIKey.Where(dao.APIKey.Key.Eq(apiKey)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return fmt.Errorf("api key not found")
		}
		return err
	}

	// Store rate limit in the settings table
	settingKey := fmt.Sprintf("rate_limit_api_key_%d", apiKeyModel.ID)
	settingValue := fmt.Sprintf("%d:%d", limit, window)

	// Use FirstOrCreate to either update or create the setting
	setting, err := dao.Setting.Where(dao.Setting.Key.Eq(settingKey)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create new setting if not found
			newSetting := &models.Setting{
				Key:   settingKey,
				Value: settingValue,
			}
			return dao.Setting.Create(newSetting)
		} else {
			return err
		}
	}

	// Update existing setting
	setting.Value = settingValue
	return dao.Setting.Save(setting)
}

// GetRateLimitForUser retrieves rate limiting configuration for a specific user
func (r *RateLimitServiceImpl) GetRateLimitForUser(userID uint) (*RateLimit, error) {
	return r.getRateLimit("user", userID)
}

// SetRateLimitForUser sets rate limiting configuration for a specific user
func (r *RateLimitServiceImpl) SetRateLimitForUser(userID uint, limit int, window int) error {
	// Get DAO instance
	dao := r.dbService.GetDAO()

	settingKey := fmt.Sprintf("rate_limit_user_%d", userID)
	settingValue := fmt.Sprintf("%d:%d", limit, window)

	// Use FirstOrCreate to either update or create the setting
	setting, err := dao.Setting.Where(dao.Setting.Key.Eq(settingKey)).First()
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			// Create new setting if not found
			newSetting := &models.Setting{
				Key:   settingKey,
				Value: settingValue,
			}
			return dao.Setting.Create(newSetting)
		} else {
			return err
		}
	}

	// Update existing setting
	setting.Value = settingValue
	return dao.Setting.Save(setting)
}
