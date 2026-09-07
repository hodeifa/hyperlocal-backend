package main

import (
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"

	"github.com/hodeifa/hyperlocal-backend/api-gateway/internal/middleware"
)

func main() {
	// Jika dijalankan dari folder api-gateway:
	// cd api-gateway && go run cmd/jwtgen/main.go
	_ = godotenv.Load("../.env")
	_ = godotenv.Load(".env")

	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		log.Fatal("JWT_SECRET is empty")
	}

	customerToken, err := middleware.GenerateAccessToken(
		"uuid-customer-123",
		"customer",
		secret,
		24*time.Hour,
	)
	if err != nil {
		log.Fatalf("failed to generate customer token: %v", err)
	}

	driverToken, err := middleware.GenerateAccessToken(
		"uuid-driver-456",
		"driver",
		secret,
		24*time.Hour,
	)
	if err != nil {
		log.Fatalf("failed to generate driver token: %v", err)
	}

	expiredToken, err := middleware.GenerateAccessToken(
		"uuid-customer-expired",
		"customer",
		secret,
		-1*time.Hour,
	)
	if err != nil {
		log.Fatalf("failed to generate expired token: %v", err)
	}

	fmt.Println("CUSTOMER_TOKEN=" + customerToken)
	fmt.Println("DRIVER_TOKEN=" + driverToken)
	fmt.Println("EXPIRED_TOKEN=" + expiredToken)
}