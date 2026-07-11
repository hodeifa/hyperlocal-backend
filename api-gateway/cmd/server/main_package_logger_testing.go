/*
how to test
APP_ENV=development LOG_LEVEL=debug go run ./api-gateway/cmd/server/main.go
curl -i http://localhost:8080/ping \
  -H "X-App-Version: 1.2.0" \
  -H "X-Platform: android" \
  -H "X-OS-Version: 14" \
  -H "X-Build: debug" \
  -H "X-Request-ID: trace-abc-123"

curl http://localhost:8080/slow -H "X-Platform: ios" 

curl http://localhost:8080/ping
*/
package main

import (
	"net/http"
	"os"
    "time"
	"github.com/gin-gonic/gin"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/hodeifa/hyperlocal-backend/pkg/logger"
	"github.com/hodeifa/hyperlocal-backend/pkg/middleware"
)

func main() {
	// 1. Parse LOG_LEVEL dari Env
	var level zapcore.Level
	if err := level.UnmarshalText([]byte(os.Getenv("LOG_LEVEL"))); err != nil {
		level = zapcore.InfoLevel // Fallback default
	}

	// 2. Inisialisasi Logger
	isProd := os.Getenv("APP_ENV") == "production"
	cfg := logger.Config{
		ServiceName:  "api-gateway", // Hardcoded per-service
		IsProduction: isProd,
		Level:        level,
	}
	
	log := logger.NewLogger(cfg)
	defer log.Sync()

	// 3. Setup Gin
	if isProd {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}
	
	r := gin.New()

	// 4. Registrasi Middleware (URUTAN WAJIB!)
	r.Use(
		middleware.ClientInfoMiddleware(),      // 1. Ekstrak X-App-Version, X-Platform, dll
		middleware.RequestIDMiddleware(),       // 2. Generate/Extract X-Request-ID
		logger.RequestLoggerMiddleware(log),  // 3. Catat semua ke log JSON/Console
		gin.Recovery(),                         // 4. Cegah server crash saat panic
	)

	// 5. Dummy Routes untuk Testing
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message":    "pong",
			"request_id": c.GetString(middleware.CtxRequestID),
		})
	})

	r.GET("/slow", func(c *gin.Context) {
		// Simulasi endpoint lambat untuk melihat duration_ms yang lebih besar
		time.Sleep(250 * time.Millisecond) 
		c.JSON(http.StatusOK, gin.H{"message": "done"})
	})

	// 6. Run Server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	
	log.Info("Starting API Gateway", zap.String("port", port), zap.String("env", os.Getenv("APP_ENV")))
	if err := r.Run(":" + port); err != nil {
		log.Fatal("Failed to start server", zap.Error(err))
	}
}