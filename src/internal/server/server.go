package server

import (
	"context"
	"strings"
	"sync"
	"time"

	"github.com/fdddf/openproxy/common"
	"github.com/fdddf/openproxy/internal/controllers"
	"github.com/fdddf/openproxy/internal/services"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"go.uber.org/fx"
)

type serverParams struct {
	fx.In
	Lifecycle        fx.Lifecycle
	ConfigService    services.ConfigService
	DatabaseService  services.DatabaseService
	ChatService      services.ChatService
	Controller       *controllers.Controller
	SecretService    services.SecretService
	SettingsService  services.SettingsService
	RateLimitService services.RateLimitService
	HealthCheck      services.HealthCheckService
	ConfigPath       string `name:"configPath" optional:"true"`
}

// StartServer wires the application lifecycle and HTTP routes.
func StartServer(p serverParams) {
	configPath := p.ConfigPath
	if err := p.ConfigService.LoadConfig(configPath); err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}
	if err := p.DatabaseService.InitDatabase(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	if _, err := p.SettingsService.GetEffectiveSettings(); err != nil {
		log.Fatalf("Failed to load settings: %v", err)
	}

	app := fiber.New()

	app.Use(logger.New())
	app.Use(cors.New())
	// app.Use(middleware.SecurityMiddleware())

	authMiddleware := controllers.NewAuthMiddleware(p.ConfigService.GetConfig().Proxy.JWTSignKey)
	proxyKeyMiddleware := newProxyKeyMiddleware(p.SecretService, p.RateLimitService)

	controllers.RegisterRoutes(app, p.Controller, authMiddleware)

	app.Static("/", "./ui", fiber.Static{
		Index: "index.html",
	})

	settingsMiddleware := newSettingsMiddleware(p.SettingsService)

	v1 := app.Group("/v1", proxyKeyMiddleware, settingsMiddleware)
	v1.Post("/chat/completions", p.ChatService.HandleChatCompletion)
	v1.Post("/responses", p.ChatService.HandleResponses)
	v1.Get("/models", p.ChatService.HandleModels)

	app.Use(func(c *fiber.Ctx) error {
		path := c.Path()
		if strings.HasPrefix(path, "/api/") || path == "/api" || strings.HasPrefix(path, "/v1/") || path == "/v1" {
			return c.Next()
		}
		return c.SendFile("./ui/index.html")
	})

	config := p.ConfigService.GetConfig()
	p.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if p.HealthCheck != nil {
				p.HealthCheck.Start()
			}
			go func() {
				log.Info("Server starting on ", config.Proxy.ListenAddress)
				log.Fatal(app.Listen(config.Proxy.ListenAddress))
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			return app.Shutdown()
		},
	})
}

func newProxyKeyMiddleware(secretService services.SecretService, rateLimitService services.RateLimitService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return common.HandleError(c, fiber.StatusUnauthorized, "Missing API key")
		}

		apiKey := strings.TrimPrefix(authHeader, "Bearer ")
		if apiKey == authHeader {
			return common.HandleError(c, fiber.StatusUnauthorized, "Invalid API key")
		}

		isValid, err := secretService.ValidateProxyAPIKey(c, apiKey)
		if err != nil {
			return common.HandleError(c, fiber.StatusInternalServerError, "Failed to validate API key")
		}
		if !isValid {
			return common.HandleError(c, fiber.StatusUnauthorized, "Invalid API key")
		}

		// Check rate limit for this API key
		if rateLimitService != nil {
			allowed, err := rateLimitService.CheckRateLimit(c, apiKey)
			if err != nil {
				log.Errorf("Rate limit check error: %v", err)
				// We don't fail the request on rate limit service errors to prevent DoS
			} else if !allowed {
				return common.HandleError(c, fiber.StatusTooManyRequests, "Rate limit exceeded")
			}
		}

		return c.Next()
	}
}

func newSettingsMiddleware(settingsService services.SettingsService) fiber.Handler {
	guards := newRequestGuards()

	return func(c *fiber.Ctx) error {
		settings, err := settingsService.GetEffectiveSettings()
		if err != nil {
			return common.HandleError(c, fiber.StatusInternalServerError, "Failed to load settings")
		}

		if settings.MaintenanceMode {
			return common.HandleError(c, fiber.StatusServiceUnavailable, "Service is in maintenance mode")
		}

		if !guards.allowRate(settings.RateLimit) {
			return common.HandleError(c, fiber.StatusTooManyRequests, "Rate limit exceeded")
		}

		acquired := guards.acquire(settings.ConcurrentLimit)
		if !acquired {
			return common.HandleError(c, fiber.StatusTooManyRequests, "Too many concurrent requests")
		}
		defer guards.release()

		return c.Next()
	}
}

type requestGuards struct {
	rateLimiter        *rateLimiter
	concurrencyLimiter *concurrencyLimiter
}

func newRequestGuards() *requestGuards {
	return &requestGuards{
		rateLimiter:        &rateLimiter{},
		concurrencyLimiter: &concurrencyLimiter{},
	}
}

func (g *requestGuards) allowRate(limit int) bool {
	if limit <= 0 {
		return true
	}
	return g.rateLimiter.allow(limit)
}

func (g *requestGuards) acquire(limit int) bool {
	if limit <= 0 {
		return true
	}
	return g.concurrencyLimiter.acquire(limit)
}

func (g *requestGuards) release() {
	g.concurrencyLimiter.release()
}

type rateLimiter struct {
	mu          sync.Mutex
	windowStart time.Time
	count       int
}

func (r *rateLimiter) allow(limit int) bool {
	now := time.Now()

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.windowStart.IsZero() || now.Sub(r.windowStart) >= time.Minute {
		r.windowStart = now
		r.count = 0
	}

	if r.count >= limit {
		return false
	}

	r.count++
	return true
}

type concurrencyLimiter struct {
	mu       sync.Mutex
	inFlight int
}

func (c *concurrencyLimiter) acquire(limit int) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.inFlight >= limit {
		return false
	}

	c.inFlight++
	return true
}

func (c *concurrencyLimiter) release() {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.inFlight > 0 {
		c.inFlight--
	}
}
