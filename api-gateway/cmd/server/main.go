package main

import (
	"log"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"

	"github.com/hodeifa/hyperlocal-backend/api-gateway/config"
	v1 "github.com/hodeifa/hyperlocal-backend/api-gateway/internal/delivery/http/v1"
	"github.com/hodeifa/hyperlocal-backend/api-gateway/internal/middleware"
)

func main() {
	// Jika menjalankan dari folder api-gateway:
	// cd api-gateway && go run cmd/server/main.go
	// maka ../.env adalah .env di root monorepo.
	if err := godotenv.Load("../.env"); err == nil {
		log.Println("loaded env from ../.env")
	}

	// Optional fallback jika Anda punya api-gateway/.env
	if err := godotenv.Load(".env"); err == nil {
		log.Println("loaded env from api-gateway/.env")
	}

	cfg := config.Load()

	log.Printf("JWT_SECRET length: %d", len(cfg.JWTSecret))	
	
	if cfg.JWTSecret == "" {
		log.Fatal("JWT_SECRET is required")
	}

	if cfg.Env == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	r := gin.New()

	r.Use(gin.Recovery())
	r.Use(middleware.ClientInfo())

	v1.SetupRouter(r, cfg)

	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"service": "api-gateway",
			"env":     cfg.Env,
		})
	})

	addr := ":" + cfg.Port
	log.Printf("API Gateway starting on %s env=%s", addr, cfg.Env)

	if err := r.Run(addr); err != nil {
		log.Fatalf("failed to start server: %v", err)
	}
}