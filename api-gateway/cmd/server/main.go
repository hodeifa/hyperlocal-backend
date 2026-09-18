package main

import (
	"log"
	"os"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	"github.com/hodeifa/hyperlocal-backend/api-gateway/config"
	v1 "github.com/hodeifa/hyperlocal-backend/api-gateway/internal/delivery/http/v1"
	"github.com/hodeifa/hyperlocal-backend/api-gateway/internal/middleware"
	"github.com/hodeifa/hyperlocal-backend/pkg/grpcclient" // Menggunakan pkg/grpcclient yang sudah ada
	pb "github.com/hodeifa/hyperlocal-backend/proto/customer/v1"
)

func main() {
	if err := godotenv.Load("../.env"); err == nil {
		log.Println("loaded env from ../.env")
	}
	if err := godotenv.Load(".env"); err == nil {
		log.Println("loaded env from api-gateway/.env")
	}

	cfg := config.Load()
	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET is required")
	}
	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	// --- TAMBAHAN BARU: Inisialisasi gRPC Client & Circuit Breaker ---
	customerGRPCAddr := os.Getenv("CUSTOMER_SERVICE_GRPC_ADDR")
	if customerGRPCAddr == "" {
		customerGRPCAddr = "localhost:50051"
	}

	opts := append(grpcclient.DefaultDialOptions, grpc.WithTransportCredentials(insecure.NewCredentials()))
	customerConn, err := grpc.Dial(customerGRPCAddr, opts...)

	if err != nil {
		log.Fatalf("failed to dial customer service: %v", err)
	}
	defer customerConn.Close()

	customerClient := pb.NewCustomerServiceClient(customerConn)
	customerCB := grpcclient.NewCircuitBreaker("customer-service") // Menggunakan NewCircuitBreaker dari pkg/grpcclient
	// --------------------------------------------------------------------

	r := gin.New()
	r.Use(gin.Recovery())
	r.Use(middleware.ClientInfo())

	// Pass dependency baru ke SetupRouter
	v1.SetupRouter(r, &cfg, customerClient, customerCB)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok", "service": "api-gateway"})
	})

	addr := ":" + cfg.Port
	log.Printf("API Gateway starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}