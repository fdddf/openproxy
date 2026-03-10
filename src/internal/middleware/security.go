package middleware

import (
	"github.com/gofiber/fiber/v2"
)

// SecurityMiddleware applies security-related HTTP headers
func SecurityMiddleware() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Additional security headers not covered by helmet
		c.Set("X-Content-Type-Options", "nosniff")
		c.Set("X-Frame-Options", "DENY") // Or "SAMEORIGIN" depending on your needs
		c.Set("X-XSS-Protection", "1; mode=block")
		c.Set("Referrer-Policy", "strict-origin-when-cross-origin")
		c.Set("Permissions-Policy", "geolocation=(), microphone=(), camera=()")

		// Content Security Policy - basic policy
		csp := "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; img-src 'self' data: https:; font-src 'self' data:; frame-ancestors 'none';"
		c.Set("Content-Security-Policy", csp)

		// Strict Transport Security (HSTS) - only if using HTTPS
		c.Set("Strict-Transport-Security", "max-age=31536000; includeSubDomains")

		return c.Next()
	}
}
