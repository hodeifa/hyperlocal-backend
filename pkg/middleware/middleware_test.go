package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/hodeifa/hyperlocal-backend/pkg/middleware"
	"github.com/stretchr/testify/assert"
)

func generateTestJWT(userID, role, secret string) string {
	claims := &middleware.JWTClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenStr, _ := token.SignedString([]byte(secret))
	return tokenStr
}

func TestJWTAndRoleGuard(t *testing.T) {
	// Setup router & middleware...

	// Test 1: Customer akses route Customer (Should Pass)
	custToken := generateTestJWT("cust-1", "customer", "test-secret")
	req1, _ := http.NewRequestWithContext(context.Background(), "GET", "/api/v1/customer/profile", http.NoBody)
	req1.Header.Set("Authorization", "Bearer "+custToken)
	// ... assert 200 OK ...

	// Test 2: Driver akses route Customer (Should Fail / 403 Forbidden)
	driverToken := generateTestJWT("drv-1", "driver", "test-secret")
	req2, _ := http.NewRequestWithContext(context.Background(), "GET", "/api/v1/customer/profile", http.NoBody)
	req2.Header.Set("Authorization", "Bearer "+driverToken)
	// ... assert 403 Forbidden ...
}

func TestAuthMiddlewares(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "test-secret"

	t.Run("REST: Tanpa header Authorization return 401", func(t *testing.T) {
		r := gin.New()
		r.Use(middleware.JWTAuthMiddleware(secret))
		r.GET("/test", func(c *gin.Context) { c.Status(200) })

		req, _ := http.NewRequestWithContext(context.Background(), "GET", "/test", http.NoBody)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("WS: Tanpa query param token return 401", func(t *testing.T) {
		r := gin.New()
		r.Use(middleware.WSAuthMiddleware(secret))
		r.GET("/ws", func(c *gin.Context) { c.Status(200) })

		req, _ := http.NewRequestWithContext(context.Background(), "GET", "/ws", http.NoBody)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("WS Guard: Return 403 untuk user asing", func(t *testing.T) {
		mockChecker := func(ctx context.Context, userID, role, orderID string) (bool, error) {
			return false, nil
		}

		r := gin.New()
		r.Use(func(c *gin.Context) {
			c.Set(middleware.CtxUserID, "user-asing")
			c.Set(middleware.CtxRole, "customer")
			c.Next()
		})
		r.Use(middleware.OrderParticipantGuard(mockChecker))
		r.GET("/chat/:order_id", func(c *gin.Context) { c.Status(200) })

		req, _ := http.NewRequestWithContext(context.Background(), "GET", "/chat/order-123", http.NoBody)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}
