package cache

import (
	"context"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
)

// ErrNotFound adalah alias untuk redis.Nil (cache miss).
var ErrNotFound = redis.Nil

// Cache mendefinisikan interface untuk operasi bisnis.
// Dirancang agar mudah di-mock untuk unit test usecase.
type Cache interface {
	Get(ctx context.Context, key string) (string, error)
	Set(ctx context.Context, key string, value any, expiration time.Duration) error
	SetEX(ctx context.Context, key string, value any, expiration time.Duration) error
	Del(ctx context.Context, keys ...string) error
	GeoAdd(ctx context.Context, key string, locations ...*redis.GeoLocation) error
	GeoRadius(ctx context.Context, key string, lon, lat, radius float64, unit string) ([]redis.GeoLocation, error)
	Ping(ctx context.Context) error
}

// Client adalah implementasi konkret dari Cache.
type Client struct {
	rdb *redis.Client
}

// Compile-time check
var _ Cache = (*Client)(nil)

// NewClient membuat koneksi Redis dan memverifikasinya (fail-fast).
func NewClient(cfg Config) (*Client, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:         cfg.Addr(),
		Password:     cfg.Password,
		DB:           cfg.DB,
		PoolSize:     cfg.PoolSize,
		MinIdleConns: cfg.MinIdleConns,
		DialTimeout:  cfg.DialTimeout,
		ReadTimeout:  cfg.ReadTimeout,
		WriteTimeout: cfg.WriteTimeout,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to redis: %w", err)
	}

	return &Client{rdb: rdb}, nil
}

// Close menutup koneksi (dipanggil di main.go, bukan di usecase).
func (c *Client) Close() error {
	return c.rdb.Close()
}

// Ping mengecek apakah Redis server reachable.
func (c *Client) Ping(ctx context.Context) error {
	return c.rdb.Ping(ctx).Err()
}

// GetRedisClient adalah escape hatch untuk command Redis yang belum di-wrap 
// (misal: Incr, SetNX, ZRem, Pub/Sub). Usecase yang butuh ini harus depend 
// ke *cache.Client secara langsung, bukan ke interface Cache.
func (c *Client) GetRedisClient() *redis.Client {
	return c.rdb
}