//go:build !integration

package cache

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func setupTestRedis(t *testing.T) (*miniredis.Miniredis, *Client) {
	mr, err := miniredis.Run()
	require.NoError(t, err)

	client, err := NewClient(Config{
		Host:         mr.Host(),
		Port:         mr.Port(),
		DialTimeout:  2 * time.Second,
		ReadTimeout:  2 * time.Second,
		WriteTimeout: 2 * time.Second,
	})
	require.NoError(t, err)

	t.Cleanup(func() {
		client.Close()
		mr.Close()
	})

	return mr, client
}

func TestGetSet(t *testing.T) {
	_, client := setupTestRedis(t)
	ctx := context.Background()

	err := client.Set(ctx, "test:key", "value123", 0)
	require.NoError(t, err)

	val, err := client.Get(ctx, "test:key")
	require.NoError(t, err)
	assert.Equal(t, "value123", val)

	_, err = client.Get(ctx, "test:nonexistent")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestSetEX(t *testing.T) {
	mr, client := setupTestRedis(t)
	ctx := context.Background()

	err := client.SetEX(ctx, "otp:phone", "123456", 5*time.Minute)
	require.NoError(t, err)

	ttl := mr.TTL("otp:phone")
	assert.True(t, ttl > 4*time.Minute && ttl <= 5*time.Minute)

	// Guard test
	err = client.SetEX(ctx, "test:invalid", "value", 0)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "expiration > 0")
}

func TestDel(t *testing.T) {
	_, client := setupTestRedis(t)
	ctx := context.Background()

	_ = client.Set(ctx, "key1", "val1", 0)
	err := client.Del(ctx, "key1")
	require.NoError(t, err)

	_, err = client.Get(ctx, "key1")
	assert.ErrorIs(t, err, ErrNotFound)
}

func TestPing(t *testing.T) {
	_, client := setupTestRedis(t)
	assert.NoError(t, client.Ping(context.Background()))
}