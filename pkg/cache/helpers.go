package cache

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

func (c *Client) Get(ctx context.Context, key string) (string, error) {
	return c.rdb.Get(ctx, key).Result()
}

func (c *Client) Set(ctx context.Context, key string, value any, expiration time.Duration) error {
	return c.rdb.Set(ctx, key, value, expiration).Err()
}

// SetEX mewajibkan expiration > 0 untuk menegakkan semantik SETEX.
func (c *Client) SetEX(ctx context.Context, key string, value any, expiration time.Duration) error {
	if expiration <= 0 {
		return errors.New("SetEX requires expiration > 0")
	}
	return c.rdb.Set(ctx, key, value, expiration).Err()
}

func (c *Client) Del(ctx context.Context, keys ...string) error {
	return c.rdb.Del(ctx, keys...).Err()
}

type GeoLocation = redis.GeoLocation

func (c *Client) GeoAdd(ctx context.Context, key string, locations ...*redis.GeoLocation) error {
	return c.rdb.GeoAdd(ctx, key, locations...).Err()
}

// GeoRadius menggunakan GEOSEARCH di balik layar untuk menghindari warning 
// staticcheck SA1019 (GEORADIUS deprecated) dan sesuai dengan erd.md §6.
func (c *Client) GeoRadius(ctx context.Context, key string, lon, lat, radius float64, unit string) ([]redis.GeoLocation, error) {
	query := &redis.GeoSearchLocationQuery{
		GeoSearchQuery: redis.GeoSearchQuery{
			Longitude:  lon,
			Latitude:   lat,
			Radius:     radius,
			RadiusUnit: unit,
			Sort:       "ASC",
		},
		WithCoord: true,
		WithDist:  true,
	}
	return c.rdb.GeoSearchLocation(ctx, key, query).Result()
}