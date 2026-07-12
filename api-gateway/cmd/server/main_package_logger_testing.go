/* Cara menjalankan testing server ini:
how to test
APP_ENV=development LOG_LEVEL=debug JWT_SECRET="test-secret-key-for-development-only" \
  go run ./api-gateway/cmd/server/main_package_logger_testing.go

# Generate JWT token manually (install jwt-cli atau gunakan online tool)
# Atau gunakan token dari script output

# Test 1: Client Info - All headers
curl -i http://localhost:8080/ping \
  -H "X-App-Version: 1.2.0" \
  -H "X-Platform: android" \
  -H "X-OS-Version: 14" \
  -H "X-Build: debug" \
  -H "X-Request-ID: trace-abc-123"

# Test 2: REST Auth - Valid token
curl -i http://localhost:8080/api/v1/orders/history \
  -H "Authorization: Bearer <YOUR_CUSTOMER_TOKEN>"

# Test 3: REST Auth - Missing token (should return 401)
curl -i http://localhost:8080/api/v1/orders/history

# Test 4: WS Auth - Valid token
curl -i "http://localhost:8080/v1/ws/chat/order-123?token=<YOUR_CUSTOMER_TOKEN>"

# Test 5: WS Auth - Missing token (should return 401)
curl -i "http://localhost:8080/v1/ws/chat/order-123"

# Test 6: WS Role Enforcement - Driver endpoint with customer token (should return 403)
curl -i "http://localhost:8080/v1/ws/driver/location?token=<YOUR_CUSTOMER_TOKEN>"

# Test 7: WS Guard - Foreign user (should return 403 or 500 depending on mock)
curl -i "http://localhost:8080/v1/ws/chat/order-999?token=<YOUR_CUSTOMER_TOKEN>"
*/

package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/hodeifa/hyperlocal-backend/pkg/middleware"
)

func main() {
	// Set Gin mode based on environment
	if os.Getenv("APP_ENV") == "production" {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	r := gin.New()

	// 1. Global Middleware (Order matters!)
	r.Use(
		middleware.ClientInfoMiddleware(),
		middleware.RequestIDMiddleware(),
		gin.Logger(),
		gin.Recovery(),
	)

	// JWT Secret from environment
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		jwtSecret = "test-secret-key-for-development-only"
		log.Println("⚠️  JWT_SECRET not set, using default for development")
	}

	// 2. Public Routes (No Auth)
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message":      "pong",
			"timestamp":    time.Now().Format(time.RFC3339),
			"request_id":   c.GetString("request_id"),
			"app_version":  c.GetString(middleware.CtxAppVersion),
			"platform":     middleware.GetPlatform(c),
			"os_version":   c.GetString(middleware.CtxOSVersion),
			"build_type":   c.GetString(middleware.CtxBuildType),
			"is_debug":     middleware.IsDebugBuild(c),
		})
	})

	r.GET("/slow", func(c *gin.Context) {
		time.Sleep(2 * time.Second)
		c.JSON(http.StatusOK, gin.H{
			"message": "slow response completed",
			"platform": middleware.GetPlatform(c),
		})
	})

	// 3. REST API Routes (JWT Auth Required)
	apiV1 := r.Group("/api/v1")
	apiV1.Use(middleware.JWTAuthMiddleware(jwtSecret))
	{
		apiV1.GET("/orders/history", func(c *gin.Context) {
			userID := c.GetString(middleware.CtxUserID)
			role := c.GetString(middleware.CtxRole)
			
			c.JSON(http.StatusOK, gin.H{
				"message": "order history",
				"user_id": userID,
				"role":    role,
				"orders": []gin.H{
					{"id": "order-1", "status": "COMPLETED"},
					{"id": "order-2", "status": "IN_PROGRESS"},
				},
			})
		})

		apiV1.GET("/profile", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"user_id": c.GetString(middleware.CtxUserID),
				"role":    c.GetString(middleware.CtxRole),
			})
		})
	}

	// 4. WebSocket Routes (JWT via Query Param)
	wsGroup := r.Group("/v1/ws")
	{
		// Chat - Both customer and driver allowed
		wsGroup.GET("/chat/:order_id",
			middleware.WSAuthMiddleware(jwtSecret),
			middleware.OrderParticipantGuard(mockParticipantChecker),
			func(c *gin.Context) {
				orderID := c.Param("order_id")
				userID := c.GetString(middleware.CtxUserID)
				role := c.GetString(middleware.CtxRole)
				
				c.JSON(http.StatusOK, gin.H{
					"message":  "WebSocket chat handshake successful",
					"order_id": orderID,
					"user_id":  userID,
					"role":     role,
					"note":     "In production, this would upgrade to WebSocket connection",
				})
			},
		)

		// Driver Location - Driver only
		wsGroup.GET("/driver/location",
			middleware.WSAuthMiddleware(jwtSecret, "driver"),
			func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{
					"message": "Driver location WebSocket ready",
					"user_id": c.GetString(middleware.CtxUserID),
					"role":    c.GetString(middleware.CtxRole),
				})
			},
		)

		// Customer Tracking - Customer only
		wsGroup.GET("/customer/tracking/:order_id",
			middleware.WSAuthMiddleware(jwtSecret, "customer"),
			middleware.OrderParticipantGuard(mockParticipantChecker),
			func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{
					"message":  "Customer tracking WebSocket ready",
					"order_id": c.Param("order_id"),
					"user_id":  c.GetString(middleware.CtxUserID),
				})
			},
		)
	}

	// Start server
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}

	log.Printf("🚀 Server starting on port %s", port)
	log.Printf("📝 Test endpoints:")
	log.Printf("   GET  /ping - Public endpoint with client info")
	log.Printf("   GET  /api/v1/orders/history - Protected REST endpoint")
	log.Printf("   GET  /v1/ws/chat/:order_id?token=JWT - WebSocket chat")
	log.Printf("   GET  /v1/ws/driver/location?token=JWT - Driver location (driver only)")
	
	if err := r.Run(":" + port); err != nil {
		log.Fatalf("Failed to start server: %v", err)
	}
}

// Mock participant checker for testing
func mockParticipantChecker(ctx context.Context, userID, role, orderID string) (bool, error) {
	// Simulate: customer-123 and driver-456 are participants of order-123
	// All other combinations are not participants
	
	if orderID == "order-123" {
		if (userID == "customer-123" && role == "customer") ||
		   (userID == "driver-456" && role == "driver") {
			return true, nil
		}
	}
	
	// Simulate DB error for order-999
	if orderID == "order-999" {
		return false, context.DeadlineExceeded
	}
	
	// Not a participant
	return false, nil
}