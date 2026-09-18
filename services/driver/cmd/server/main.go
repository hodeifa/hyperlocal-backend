package main

import (
	"context"
	"fmt"
	"net"
	"net/url"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/joho/godotenv"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"google.golang.org/grpc"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"github.com/hodeifa/hyperlocal-backend/pkg/database"
	"github.com/hodeifa/hyperlocal-backend/pkg/grpcclient"
	"github.com/hodeifa/hyperlocal-backend/pkg/logger"

	driverv1 "github.com/hodeifa/hyperlocal-backend/proto/driver/v1"

	grpcHandler "github.com/hodeifa/hyperlocal-backend/services/driver/internal/delivery/grpc/v1"
)

// loadEnv mencoba membaca file .env dari beberapa lokasi
func loadEnv() {
	possiblePaths := []string{
		".env",
		"../.env",
		"../../.env",
		"../../../.env",
	}

	for _, path := range possiblePaths {
		if err := godotenv.Load(path); err == nil {
			fmt.Printf("✅ Loaded environment variables from: %s\n", path)
			return
		}
	}

	fmt.Println("ℹ️  No .env file found. Using system environment variables.")
}

func getEnv(key, fallback string) string {
	if v, ok := os.LookupEnv(key); ok {
		return v
	}
	return fallback
}

func buildDatabaseURL() string {
	user := getEnv("POSTGRES_USER", "hyperlocal")
	pass := getEnv("POSTGRES_PASSWORD", "hyperlocal_secret")
	name := getEnv("POSTGRES_DB", "hyperlocal")
	host := getEnv("DB_HOST", "localhost")
	port := getEnv("DB_PORT", "5432")
	sslmode := getEnv("DB_SSLMODE", "disable")

	return (&url.URL{
		Scheme:   "postgres",
		User:     url.UserPassword(user, pass),
		Host:     fmt.Sprintf("%s:%s", host, port),
		Path:     name,
		RawQuery: url.Values{"sslmode": {sslmode}}.Encode(),
	}).String()
}

func main() {
	// WAJIB: Load env vars dulu sebelum baca konfigurasi apa pun
	loadEnv()

	// 1. Initialize Logger
	log := logger.NewLogger(logger.Config{
		ServiceName:  "driver-service",
		IsProduction: os.Getenv("ENV") == "production",
		Level:        zapcore.InfoLevel,
	})
	defer log.Sync()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 2. Initialize Database
	// Catatan: Driver Service butuh DB untuk query tabel `drivers` (Sprint 5)
	db, err := database.New(ctx, database.Config{
		URL:             buildDatabaseURL(),
		MaxConns:        10,
		MinConns:        2,
		MaxConnLifetime: 30 * time.Minute,
		MaxConnIdleTime: 5 * time.Minute,
	}, log)
	if err != nil {
		log.Fatal("failed to connect to database", zap.Error(err))
	}
	defer db.Close()
	log.Info("Database connected successfully")

	// 3. Dependency Injection
	// TODO: Pass `db` ke handler saat implementasi asli CheckPhoneExists di Sprint 5
	driverGRPCHandler := grpcHandler.NewDriverGRPCHandler()

	// 4. gRPC Server Setup
	grpcServer := grpc.NewServer(grpcclient.DefaultServerOptions...)
	driverv1.RegisterDriverServiceServer(grpcServer, driverGRPCHandler)

	// Health check (wajib untuk Kubernetes probes)
	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	// Reflection (untuk debugging dengan grpcurl/Postman)
	reflection.Register(grpcServer)

	// 5. Start Listener
	grpcHost := getEnv("DRIVER_SERVICE_HOST", "localhost")
	grpcPort := getEnv("DRIVER_SERVICE_GRPC_PORT", "50052")
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%s", grpcHost, grpcPort))
	if err != nil {
		log.Fatal("failed to listen", zap.Error(err))
	}

	errChan := make(chan error, 1)
	go func() {
		log.Info("Driver Service gRPC server is running", zap.String("port", grpcPort))
		if err := grpcServer.Serve(lis); err != nil {
			errChan <- err
		}
	}()

	// 6. Graceful Shutdown
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)

	select {
	case err := <-errChan:
		log.Error("gRPC Server error", zap.Error(err))
	case sig := <-quit:
		log.Info("Received shutdown signal", zap.String("signal", sig.String()))
	}

	log.Info("Shutting down gRPC Server...")
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_NOT_SERVING)
	grpcServer.GracefulStop()
	log.Info("Driver Service stopped gracefully")
}