package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

type Cache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewCache(addr, password string, db int, ttl time.Duration) (*Cache, error) {
	client := redis.NewClient(&redis.Options{
		Addr:     addr,
		Password: password,
		DB:       db,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := client.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	return &Cache{
		client: client,
		ttl:    ttl,
	}, nil
}

// GetVisitorID retrieves a cached visitorID by hardware hash.
func (c *Cache) GetVisitorID(ctx context.Context, hardwareHash string) (string, error) {
	key := fmt.Sprintf("hw:%s", hardwareHash)
	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return "", nil // Cache miss
	}
	if err != nil {
		return "", fmt.Errorf("cache get error: %w", err)
	}
	return val, nil
}

// SetVisitorID caches the hardware hash to visitorID mapping.
func (c *Cache) SetVisitorID(ctx context.Context, hardwareHash, visitorID string) error {
	key := fmt.Sprintf("hw:%s", hardwareHash)
	if err := c.client.Set(ctx, key, visitorID, c.ttl).Err(); err != nil {
		return fmt.Errorf("cache set error: %w", err)
	}
	return nil
}

// CheckRateLimit implements token bucket rate limiting.
func (c *Cache) CheckRateLimit(ctx context.Context, identifier string, limit int, window time.Duration) (bool, error) {
	key := fmt.Sprintf("rl:%s", identifier)

	pipe := c.client.Pipeline()
	incr := pipe.Incr(ctx, key)
	pipe.Expire(ctx, key, window)

	if _, err := pipe.Exec(ctx); err != nil {
		return false, fmt.Errorf("rate limit check error: %w", err)
	}

	count := incr.Val()
	return count <= int64(limit), nil
}

// IncrementMetric increments a counter metric.
func (c *Cache) IncrementMetric(ctx context.Context, metric string) error {
	key := fmt.Sprintf("metric:%s", metric)
	return c.client.Incr(ctx, key).Err()
}

// GetMetric retrieves a metric value.
func (c *Cache) GetMetric(ctx context.Context, metric string) (int64, error) {
	key := fmt.Sprintf("metric:%s", metric)
	val, err := c.client.Get(ctx, key).Result()
	if err == redis.Nil {
		return 0, nil
	}
	if err != nil {
		return 0, err
	}

	var count int64
	if _, err := fmt.Sscanf(val, "%d", &count); err != nil {
		return 0, err
	}
	return count, nil
}

// Close closes the Redis connection.
func (c *Cache) Close() error {
	return c.client.Close()
}
