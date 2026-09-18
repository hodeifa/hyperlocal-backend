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
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"
	"google.golang.org/grpc/reflection"

	"github.com/hodeifa/hyperlocal-backend/pkg/cache"
	"github.com/hodeifa/hyperlocal-backend/pkg/database"
	"github.com/hodeifa/hyperlocal-backend/pkg/grpcclient"
	"github.com/hodeifa/hyperlocal-backend/pkg/logger"

	customerv1 "github.com/hodeifa/hyperlocal-backend/proto/customer/v1"
	driverv1 "github.com/hodeifa/hyperlocal-backend/proto/driver/v1"

	grpcHandler "github.com/hodeifa/hyperlocal-backend/services/customer/internal/delivery/grpc/v1"
	usecase "github.com/hodeifa/hyperlocal-backend/services/customer/internal/usecase/v1"
)

func loadEnv() {
	// Daftar path yang mungkin, dari yang paling spesifik ke paling umum
	possiblePaths := []string{
		".env",                    // Current directory
		"../.env",                 // Parent (jika run dari cmd/)
		"../../.env",              // Root monorepo (jika run dari services/customer/cmd/server/)
		"../../../.env",           // Alternative jika struktur berbeda
	}

	for _, path := range possiblePaths {
		if err := godotenv.Load(path); err == nil {
			fmt.Printf("✅ Loaded environment variables from: %s\n", path)
			return
		}
	}

	// Jika tidak ada file .env yang ditemukan, tidak error
	// Karena di production biasanya env vars di-inject via Docker/Kubernetes
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
	// ⚠️ WAJIB: Load env vars dulu sebelum baca konfigurasi apa pun
	loadEnv()
	// 1. Initialize Logger
	log := logger.NewLogger(logger.Config{
		ServiceName:  "customer-service",
		IsProduction: os.Getenv("ENV") == "production",
		Level:        zapcore.InfoLevel,
	})
	defer log.Sync()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	// 2. Initialize Database
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

	// 3. Initialize Cache
	cacheClient, err := cache.NewClient(cache.Config{
		Host:     getEnv("REDIS_HOST", "localhost"),
		Port:     getEnv("REDIS_PORT", "6379"),
		Password: getEnv("REDIS_PASSWORD", ""),
		DB:       0,
	})
	if err != nil {
		log.Fatal("failed to connect to Redis", zap.Error(err))
	}
	defer cacheClient.Close()
	log.Info("Redis connected successfully")

	// 4. Initialize gRPC Client to Driver Service
	driverAddr := fmt.Sprintf("%s:%s",
		getEnv("DRIVER_SERVICE_HOST", "localhost"),
		getEnv("DRIVER_SERVICE_GRPC_PORT", "50052"),
	)

	// [FIX] Gabungkan DefaultDialOptions dengan insecure credentials menjadi satu slice
	opts := append(grpcclient.DefaultDialOptions, grpc.WithTransportCredentials(insecure.NewCredentials()))
	
	driverConn, err := grpc.NewClient(driverAddr, opts...)
	if err != nil {
		log.Fatal("failed to dial Driver Service", zap.Error(err))
	}
	defer driverConn.Close()

	driverClient := driverv1.NewDriverServiceClient(driverConn)
	driverCB := grpcclient.NewCircuitBreaker("driver-service")
	log.Info("Driver Service gRPC client initialized", zap.String("address", driverAddr))

	// 5. Dependency Injection
	authUsecase := usecase.NewAuthUsecase(cacheClient, driverClient, driverCB, log)
	authGRPCHandler := grpcHandler.NewAuthGRPCHandler(authUsecase)

	// 6. gRPC Server Setup
	grpcServer := grpc.NewServer(grpcclient.DefaultServerOptions...)
	customerv1.RegisterCustomerServiceServer(grpcServer, authGRPCHandler)

	healthServer := health.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, healthServer)
	healthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	reflection.Register(grpcServer)

	// 7. Start Listener
	grpcHost := getEnv("CUSTOMER_SERVICE_HOST", "localhost") // ganti sesuai service
	grpcPort := getEnv("CUSTOMER_SERVICE_GRPC_PORT", "50051")
	lis, err := net.Listen("tcp", fmt.Sprintf("%s:%s", grpcHost, grpcPort))
	if err != nil {
		log.Fatal("failed to listen", zap.Error(err))
	}

	errChan := make(chan error, 1)
	go func() {
		log.Info("gRPC Server is running", zap.String("address", lis.Addr().String()))
		if err := grpcServer.Serve(lis); err != nil {
			errChan <- err
		}
	}()

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
	log.Info("Customer Service stopped gracefully")
}