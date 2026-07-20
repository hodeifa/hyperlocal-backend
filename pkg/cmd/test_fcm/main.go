// Package main is a CLI tool to verify FCM integration.
package main

import (
	"context"
	"log"
	"os"

	"github.com/hodeifa/hyperlocal-backend/pkg/fcm"
	"github.com/hodeifa/hyperlocal-backend/pkg/logger"
	"go.uber.org/zap/zapcore"
)

// mockFetcher implements fcm.TokenFetcher for CLI testing.
type mockFetcher struct {
	token string
}

func (m *mockFetcher) GetCustomerFCMToken(ctx context.Context, customerID string) (string, error) {
	return m.token, nil
}

func (m *mockFetcher) GetDriverFCMToken(ctx context.Context, driverID string) (string, error) {
	return m.token, nil
}

func main() {
	credFile := os.Getenv("FIREBASE_CREDENTIALS_FILE")
	credJSON := os.Getenv("FIREBASE_CREDENTIALS_JSON")
	testToken := os.Getenv("TEST_FCM_TOKEN")

	if (credFile == "" && credJSON == "") || testToken == "" {
		log.Fatal("FIREBASE_CREDENTIALS_FILE (atau _JSON) dan TEST_FCM_TOKEN wajib diset di .env")
	}

	cfg := logger.Config{
		ServiceName:  "test-fcm",
		IsProduction: false,
		Level:        zapcore.InfoLevel,
	}
	zapLogger := logger.NewLogger(cfg)

	fcmCfg := fcm.Config{
		CredentialsFile: credFile,
		CredentialsJSON: credJSON,
	}

	client, err := fcm.NewClient(context.Background(), fcmCfg, &mockFetcher{token: testToken}, zapLogger)
	if err != nil {
		log.Fatalf("Gagal inisialisasi FCM: %v", err)
	}

	log.Println("Mengirim test notification ke driver...")
	err = client.SendToDriver(context.Background(), "test-driver", "Hyperlocal Test", "Setup pkg/fcm berhasil!", "hmitra://home")
	if err != nil {
		log.Fatalf("Gagal kirim: %v", err)
	}
	log.Println("✅ Test notification terkirim! Cek device atau FCM Console.")
}
