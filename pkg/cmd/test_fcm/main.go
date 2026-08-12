// Package main provides a CLI tool to verify FCM setup locally without a database.
package main

import (
	"context"
	"fmt"
	"os"

	"github.com/hodeifa/hyperlocal-backend/pkg/fcm"
	"go.uber.org/zap"
)

// mockFetcher implements fcm.TokenFetcher for CLI testing.
type mockFetcher struct{}

// GetCustomerFCMToken returns an empty string for CLI mock.
func (m *mockFetcher) GetCustomerFCMToken(ctx context.Context, id string) (string, error) {
	return "", nil
}

// GetDriverFCMToken returns an empty string for CLI mock.
func (m *mockFetcher) GetDriverFCMToken(ctx context.Context, id string) (string, error) {
	return "", nil
}

func main() {
	logger, _ := zap.NewProduction()

	token := os.Getenv("TEST_FCM_TOKEN")
	if token == "" {
		logger.Fatal("TEST_FCM_TOKEN environment variable is required")
	}

	cfg := fcm.Config{
		CredentialsFile: os.Getenv("FIREBASE_CREDENTIALS_FILE"),
		CredentialsJSON: os.Getenv("FIREBASE_CREDENTIALS_JSON"),
	}

	ctx := context.Background()
	client, err := fcm.NewClient(ctx, cfg, &mockFetcher{}, logger)
	if err != nil {
		logger.Fatal("Failed to init FCM client", zap.Error(err))
	}

	fmt.Println("🚀 Mengirim pesan test ke FCM...")
	err = client.Send(ctx, token, "🔔 Sprint 3 Test", "Setup pkg/fcm berhasil!", map[string]string{
		"deep_link": "hmitra://home",
		"type":      "test",
	})

	if err != nil {
		logger.Fatal("Gagal kirim", zap.Error(err))
	}

	fmt.Println("✅ SUKSES! Cek FCM Console atau Device Anda.")
}
