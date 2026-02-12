package middleware

import (
	"fmt"
	"net/http"
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/iamgideonidoko/signet/internal/config"
	"github.com/iamgideonidoko/signet/pkg/cache"
)

type RateLimiter struct {
	cache  *cache.Cache
	config *config.RateLimitConfig
}

func NewRateLimiter(cache *cache.Cache, config *config.RateLimitConfig) *RateLimiter {
	return &RateLimiter{
		cache:  cache,
		config: config,
	}
}

// LimitByIP rate limits requests by IP address
func (rl *RateLimiter) LimitByIP() fiber.Handler {
	return func(c *fiber.Ctx) error {
		ip := c.IP()
		identifier := fmt.Sprintf("ip:%s", ip)

		allowed, err := rl.cache.CheckRateLimit(
			c.Context(),
			identifier,
			rl.config.Requests,
			rl.config.Window,
		)

		if err != nil {
			// Log error but don't block on cache failure
			return c.Next()
		}

		if !allowed {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":       "Rate limit exceeded",
				"retry_after": rl.config.Window.Seconds(),
			})
		}

		return c.Next()
	}
}

// LimitByHardwareHash rate limits by hardware fingerprint hash
func (rl *RateLimiter) LimitByHardwareHash() fiber.Handler {
	return func(c *fiber.Ctx) error {
		// This middleware runs after the request body is parsed
		// We'll extract the hardware hash from the context if set by the handler
		hardwareHash := c.Locals("hardware_hash")
		if hardwareHash == nil {
			return c.Next()
		}

		identifier := fmt.Sprintf("hw:%s", hardwareHash)

		allowed, err := rl.cache.CheckRateLimit(
			c.Context(),
			identifier,
			rl.config.RequestsByHardware,
			rl.config.HardwareWindow,
		)

		if err != nil {
			return c.Next()
		}

		if !allowed {
			return c.Status(fiber.StatusTooManyRequests).JSON(fiber.Map{
				"error":       "Hardware rate limit exceeded",
				"retry_after": rl.config.HardwareWindow.Seconds(),
			})
		}

		return c.Next()
	}
}

// CORS middleware for cross-origin requests
func CORS(origins []string) fiber.Handler {
	allowedOrigins := make(map[string]bool)
	for _, origin := range origins {
		allowedOrigins[origin] = true
	}

	return func(c *fiber.Ctx) error {
		origin := c.Get("Origin")

		// Allow all origins if "*" is configured
		if allowedOrigins["*"] || allowedOrigins[origin] {
			c.Set("Access-Control-Allow-Origin", origin)
			c.Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS")
			c.Set("Access-Control-Allow-Headers", "Content-Type, Authorization")
			c.Set("Access-Control-Max-Age", "3600")
		}

		if c.Method() == "OPTIONS" {
			return c.SendStatus(http.StatusNoContent)
		}

		return c.Next()
	}
}

// Logger middleware for request logging
func Logger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := c.Context().Time()

		err := c.Next()

		duration := c.Context().Time().Sub(start)

		status := c.Response().StatusCode()
		method := c.Method()
		path := c.Path()
		ip := c.IP()

		fmt.Printf("[%s] %s %s - %d (%v) - IP: %s\n",
			start.Format("2006-01-02 15:04:05"),
			method, path, status, duration, ip,
		)

		return err
	}
}

// Recover middleware for panic recovery
func Recover() fiber.Handler {
	return func(c *fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("PANIC: %v\n", r)
				c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Internal server error",
				})
			}
		}()
		return c.Next()
	}
}

// extractRealIP extracts the real IP considering trusted proxies
func extractRealIP(c *fiber.Ctx, trustedProxies []string) string {
	// Check X-Forwarded-For header if from trusted proxy
	if len(trustedProxies) > 0 {
		xff := c.Get("X-Forwarded-For")
		if xff != "" {
			ips := strings.Split(xff, ",")
			if len(ips) > 0 {
				return strings.TrimSpace(ips[0])
			}
		}
	}

	return c.IP()
}

// AnonymizeIP removes the last octet for privacy (GDPR compliance)
func AnonymizeIP(ip string) string {
	parts := strings.Split(ip, ".")
	if len(parts) == 4 {
		// IPv4: zero out last octet
		return fmt.Sprintf("%s.%s.%s.0", parts[0], parts[1], parts[2])
	}

	// For IPv6 or invalid format, return as-is
	// In production, you'd want proper IPv6 anonymization
	return ip
}
