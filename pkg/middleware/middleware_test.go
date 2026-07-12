package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/stretchr/testify/assert"
	"github.com/hodeifa/hyperlocal-backend/pkg/middleware"
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

func TestAuthMiddlewares(t *testing.T) {
	gin.SetMode(gin.TestMode)
	secret := "test-secret"

	t.Run("REST: Tanpa header Authorization return 401", func(t *testing.T) {
		r := gin.New()
		r.Use(middleware.JWTAuthMiddleware(secret))
		r.GET("/test", func(c *gin.Context) { c.Status(200) })

		req, _ := http.NewRequest("GET", "/test", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusUnauthorized, w.Code)
	})

	t.Run("WS: Tanpa query param token return 401", func(t *testing.T) {
		r := gin.New()
		r.Use(middleware.WSAuthMiddleware(secret))
		r.GET("/ws", func(c *gin.Context) { c.Status(200) })

		req, _ := http.NewRequest("GET", "/ws", nil)
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

		req, _ := http.NewRequest("GET", "/chat/order-123", nil)
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		assert.Equal(t, http.StatusForbidden, w.Code)
	})
}