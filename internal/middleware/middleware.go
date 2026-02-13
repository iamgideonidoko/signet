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

// LimitByIP rate limits requests by IP address.
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

// LimitByHardwareHash rate limits by hardware fingerprint hash.
func (rl *RateLimiter) LimitByHardwareHash() fiber.Handler {
	return func(c *fiber.Ctx) error {
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

func CORS(origins []string) fiber.Handler {
	allowedOrigins := make(map[string]bool)
	for _, origin := range origins {
		allowedOrigins[origin] = true
	}

	return func(c *fiber.Ctx) error {
		origin := c.Get("Origin")

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

func Logger() fiber.Handler {
	return func(c *fiber.Ctx) error {
		start := c.Context().Time()
		err := c.Next()
		duration := c.Context().Time().Sub(start)

		fmt.Printf("[%s] %s %s - %d (%v) - IP: %s\n",
			start.Format("2006-01-02 15:04:05"),
			c.Method(), c.Path(), c.Response().StatusCode(), duration, c.IP(),
		)

		return err
	}
}

func Recover() fiber.Handler {
	return func(c *fiber.Ctx) error {
		defer func() {
			if r := recover(); r != nil {
				fmt.Printf("PANIC: %v\n", r)
				_ = c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
					"error": "Internal server error",
				})
			}
		}()
		return c.Next()
	}
}

// AnonymizeIP removes the last octet for GDPR compliance.
func AnonymizeIP(ip string) string {
	parts := strings.Split(ip, ".")
	if len(parts) == 4 {
		return fmt.Sprintf("%s.%s.%s.0", parts[0], parts[1], parts[2])
	}
	return ip
}
