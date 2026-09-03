package server

import (
	"context"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/fdddf/openproxy/common"
	"github.com/fdddf/openproxy/internal/controllers"
	"github.com/fdddf/openproxy/internal/services"
	"github.com/fdddf/openproxy/internal/web"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/log"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/filesystem"
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
	RequestLogger    services.RequestLogger
	SetupService     services.SetupService
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

	// Resolve the signing key before any route is wired: an unset key would
	// otherwise sign and accept tokens with an empty secret.
	signingKey, err := p.SetupService.EnsureJWTSignKey()
	if err != nil {
		log.Fatalf("Failed to resolve JWT signing key: %v", err)
	}
	if needsSetup, err := p.SetupService.NeedsSetup(); err == nil && needsSetup {
		log.Info("No administrator exists yet; open the UI to run first-time setup")
	}

	authMiddleware := controllers.NewAuthMiddleware(signingKey)
	adminMiddleware := controllers.NewAdminMiddleware(p.DatabaseService)
	proxyKeyMiddleware := newProxyKeyMiddleware(p.SecretService, p.RateLimitService)

	controllers.RegisterRoutes(app, p.Controller, authMiddleware, adminMiddleware)

	uiAvailable := web.Available()
	if uiAvailable {
		app.Use("/", filesystem.New(filesystem.Config{
			Root:  http.FS(web.Assets()),
			Index: "index.html",
		}))
	} else {
		log.Warn("admin UI is not embedded in this binary; build it with `make ui`")
	}

	settingsMiddleware := newSettingsMiddleware(p.SettingsService)

	v1 := app.Group("/v1", proxyKeyMiddleware, settingsMiddleware)
	v1.Post("/chat/completions", p.ChatService.HandleChatCompletion)
	v1.Post("/responses", p.ChatService.HandleResponses)
	v1.Get("/models", p.ChatService.HandleModels)

	// SPA fallback: anything that is not an API route and did not match a built
	// asset renders the Vue entrypoint so client-side routing works on reload.
	app.Use(func(c *fiber.Ctx) error {
		path := c.Path()
		if strings.HasPrefix(path, "/api/") || path == "/api" || strings.HasPrefix(path, "/v1/") || path == "/v1" {
			return c.Next()
		}
		if !uiAvailable {
			return c.Status(fiber.StatusNotFound).
				Type("txt").
				SendString("admin UI is not embedded in this binary; build it with `make ui`")
		}
		index, err := web.Index()
		if err != nil {
			return common.HandleError(c, fiber.StatusInternalServerError, "Failed to load UI")
		}
		// The static middleware sets 404 before falling through, so the status
		// has to be reset for the SPA entrypoint to load normally.
		return c.Status(fiber.StatusOK).Type("html").Send(index)
	})

	config := p.ConfigService.GetConfig()
	p.Lifecycle.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			p.RequestLogger.Start()
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
			if err := app.Shutdown(); err != nil {
				return err
			}
			if p.HealthCheck != nil {
				p.HealthCheck.Stop()
			}
			// Flush queued request logs before the process exits.
			return p.RequestLogger.Stop(ctx)
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
