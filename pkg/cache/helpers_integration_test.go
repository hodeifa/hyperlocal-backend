//go:build integration

package cache

import (
	"context"
	"net/url"
	"testing"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go"
	tcredis "github.com/testcontainers/testcontainers-go/modules/redis"
)

func setupTestRedisIntegration(t *testing.T) *Client {
	ctx := context.Background()
	
	// [FIX] Menggunakan API v0.31.0: RunContainer (bukan Run yang baru ada di v0.32.0)
	// Image dipassing lewat testcontainers.WithImage, bukan argumen posisi kedua.
	redisContainer, err := tcredis.RunContainer(ctx,
		testcontainers.WithImage("redis:7-alpine"),
	)
	require.NoError(t, err, "Failed to start Redis container. Is Docker running?")

	t.Cleanup(func() {
		_ = redisContainer.Terminate(ctx)
	})

	// Mendapatkan connection string (format: "redis://localhost:xxxxx")
	connStr, err := redisContainer.ConnectionString(ctx)
	require.NoError(t, err)

	// Parse URL untuk mengambil Host dan Port agar sesuai dengan struct Config kita
	u, err := url.Parse(connStr)
	require.NoError(t, err)

	client, err := NewClient(Config{
		Host:         u.Hostname(),
		Port:         u.Port(),
		DialTimeout:  5 * time.Second,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		_ = client.Close()
	})

	return client
}

func TestGeoAddAndGeoRadius_Integration(t *testing.T) {
	client := setupTestRedisIntegration(t)
	ctx := context.Background()

	// GeoAdd
	err := client.GeoAdd(ctx, "drivers_online",
		&redis.GeoLocation{Name: "driver_1", Longitude: 106.8228, Latitude: -6.1930},
		&redis.GeoLocation{Name: "driver_2", Longitude: 106.8250, Latitude: -6.1950},
		&redis.GeoLocation{Name: "driver_3", Longitude: 107.0000, Latitude: -6.3000}, // Jauh
	)
	require.NoError(t, err)

	// GeoRadius (Internal menggunakan GEOSEARCH)
	locations, err := client.GeoRadius(ctx, "drivers_online", 106.8230, -6.1940, 1.0, "km")
	require.NoError(t, err)
	assert.Len(t, locations, 2) // Hanya driver_1 dan driver_2

	for _, loc := range locations {
		assert.NotEmpty(t, loc.Name)
		assert.True(t, loc.Dist >= 0)
		
		// Assertion float64 yang valid
		if loc.Name == "driver_1" {
			assert.InDelta(t, 106.8228, loc.Longitude, 0.0001)
			assert.InDelta(t, -6.1930, loc.Latitude, 0.0001)
		} else if loc.Name == "driver_2" {
			assert.InDelta(t, 106.8250, loc.Longitude, 0.0001)
			assert.InDelta(t, -6.1950, loc.Latitude, 0.0001)
		}
	}
}