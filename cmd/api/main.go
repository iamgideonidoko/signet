package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"

	"github.com/iamgideonidoko/signet/internal/config"
	"github.com/iamgideonidoko/signet/internal/handlers"
	"github.com/iamgideonidoko/signet/internal/middleware"
	"github.com/iamgideonidoko/signet/internal/repository"
	"github.com/iamgideonidoko/signet/internal/services"
	"github.com/iamgideonidoko/signet/pkg/cache"
	"github.com/iamgideonidoko/signet/pkg/logger"
)

func main() {
	_ = godotenv.Load()

	cfg, err := config.Load()
	if err != nil {
		logger.Error("Failed to load config", map[string]any{"error": err.Error()})
		os.Exit(1)
	}

	logger.SetLevel(logger.ParseLevel(cfg.Monitoring.LogLevel))
	logger.Info("Starting Signet API", map[string]any{
		"version":     "0.0.1",
		"environment": cfg.API.Environment,
	})

	var repo *repository.Repository
	err = repository.WithRetry(context.Background(), repository.DefaultRetryConfig, func() error {
		var retryErr error
		repo, retryErr = repository.NewRepository(
			cfg.Database.URL,
			cfg.Database.MaxConns,
			cfg.Database.MaxIdleConns,
		)
		return retryErr
	})
	if err != nil {
		logger.Error("Failed to connect to database", map[string]any{"error": err.Error()})
		os.Exit(1)
	}
	logger.Info("Connected to PostgreSQL", map[string]any{
		"url": cfg.Database.URL,
	})

	if err := repo.HealthCheck(context.Background()); err != nil {
		logger.Error("Database health check failed", map[string]any{"error": err.Error()})
		_ = repo.Close()
		os.Exit(1)
	}

	var redisCache *cache.Cache
	err = repository.WithRetry(context.Background(), repository.DefaultRetryConfig, func() error {
		var retryErr error
		redisCache, retryErr = cache.NewCache(
			cfg.Redis.URL,
			cfg.Redis.CacheTTL,
		)
		return retryErr
	})
	if err != nil {
		logger.Error("Failed to connect to Redis", map[string]any{"error": err.Error()})
		_ = repo.Close()
		os.Exit(1)
	}
	logger.Info("Connected to Redis", map[string]any{
		"url": cfg.Redis.URL,
	})

	// Deferred cleanup only runs after successful initialization
	defer func() {
		if err := repo.Close(); err != nil {
			logger.Error("Failed to close database connection", map[string]any{"error": err.Error()})
		}
	}()
	defer func() {
		if err := redisCache.Close(); err != nil {
			logger.Error("Failed to close Redis connection", map[string]any{"error": err.Error()})
		}
	}()

	identService := services.NewIdentificationService(repo, redisCache, &cfg.Fingerprint)
	logger.Info("Initialized identification service")

	handler := handlers.NewHandler(identService, redisCache)

	app := fiber.New(fiber.Config{
		DisableStartupMessage: false,
		ServerHeader:          "Signet",
		AppName:               "Signet API v1.0",
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			code := fiber.StatusInternalServerError
			if e, ok := err.(*fiber.Error); ok {
				code = e.Code
			}
			logger.Error("Request error", map[string]any{
				"error": err.Error(),
				"path":  c.Path(),
				"code":  code,
			})
			return c.Status(code).JSON(fiber.Map{
				"error": err.Error(),
			})
		},
	})

	app.Use(middleware.Recover())
	app.Use(middleware.Logger())
	app.Use(middleware.CORS(cfg.Security.CORSOrigins))

	rateLimiter := middleware.NewRateLimiter(redisCache, &cfg.RateLimit)

	app.Get("/health", handler.Health)
	app.Get("/metrics", handler.Metrics)
	app.Get("/dashboard", handler.Dashboard)

	v1 := app.Group("/v1")
	v1.Post("/identify",
		rateLimiter.LimitByIP(),
		handler.Identify,
	)

	api := app.Group("/api")
	api.Get("/analytics", handler.Analytics)
	api.Get("/identifications", handler.RecentIdentifications)

	app.Static("/agent.js", "./agent/dist/agent.min.js")

	sigChan := make(chan os.Signal, 1)
	signal.Notify(sigChan, os.Interrupt, syscall.SIGTERM)

	go func() {
		<-sigChan
		logger.Info("Shutting down gracefully...")

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		_ = app.ShutdownWithContext(ctx)
		logger.Info("Server shutdown complete")
		os.Exit(0)
	}()

	addr := fmt.Sprintf("%s:%s", cfg.API.Host, cfg.API.Port)
	logger.Info("Signet API started", map[string]any{
		"address":   addr,
		"dashboard": fmt.Sprintf("http://%s/dashboard", addr),
	})

	if err := app.Listen(addr); err != nil {
		logger.Error("Server error", map[string]any{"error": err.Error()})
	}
}
