package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
)

const testJWTSecret = "test-secret-key-for-unit-test-only"

func setupJWTTestRouter(secret string) *gin.Engine {
	gin.SetMode(gin.TestMode)

	router := gin.New()

	// Route protected — butuh JWT valid
	router.GET("/api/v1/customer/profile",
		JWTAuthMiddleware(JWTAuthConfig{SecretKey: secret}),
		RoleGuard(),
		func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{
				"user_id": c.GetString("user_id"),
				"role":    c.GetString(CtxUserRole),
			})
		},
	)

	return router
}

func TestJWTAuthMiddleware_ValidCustomerToken(t *testing.T) {
	router := setupJWTTestRouter(testJWTSecret)

	// Generate token customer valid (24 jam)
	token, err := GenerateAccessToken("uuid-customer-123", "customer", testJWTSecret, 24*time.Hour)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/customer/profile", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body: %s", w.Code, w.Body.String())
	}
}

func TestJWTAuthMiddleware_ValidDriverToken_CustomerRoute_Should403(t *testing.T) {
	router := setupJWTTestRouter(testJWTSecret)

	// Generate token DRIVER — coba akses route CUSTOMER
	token, err := GenerateAccessToken("uuid-driver-456", "driver", testJWTSecret, 24*time.Hour)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/customer/profile", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	// RoleGuard harus menolak dengan 403
	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d, body: %s", w.Code, w.Body.String())
	}
}

func TestJWTAuthMiddleware_MissingAuthorizationHeader(t *testing.T) {
	router := setupJWTTestRouter(testJWTSecret)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/customer/profile", nil)
	// Tidak set header Authorization
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d, body: %s", w.Code, w.Body.String())
	}
}

func TestJWTAuthMiddleware_InvalidTokenFormat(t *testing.T) {
	router := setupJWTTestRouter(testJWTSecret)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/customer/profile", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d, body: %s", w.Code, w.Body.String())
	}
}

func TestJWTAuthMiddleware_ExpiredToken(t *testing.T) {
	router := setupJWTTestRouter(testJWTSecret)

	// Generate token yang sudah expired (expiry = -1 jam)
	token, err := GenerateAccessToken("uuid-customer-123", "customer", testJWTSecret, -1*time.Hour)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/customer/profile", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for expired token, got %d, body: %s", w.Code, w.Body.String())
	}
}

func TestJWTAuthMiddleware_WrongSecret(t *testing.T) {
	router := setupJWTTestRouter(testJWTSecret)

	// Generate token dengan secret BERBEDA
	token, err := GenerateAccessToken("uuid-customer-123", "customer", "wrong-secret", 24*time.Hour)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/customer/profile", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for wrong secret, got %d, body: %s", w.Code, w.Body.String())
	}
}

func TestJWTAuthMiddleware_NoBearerPrefix(t *testing.T) {
	router := setupJWTTestRouter(testJWTSecret)

	token, err := GenerateAccessToken("uuid-customer-123", "customer", testJWTSecret, 24*time.Hour)
	if err != nil {
		t.Fatalf("failed to generate token: %v", err)
	}

	req := httptest.NewRequest(http.MethodGet, "/api/v1/customer/profile", nil)
	// Kirim token tanpa prefix "Bearer"
	req.Header.Set("Authorization", token)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401 for missing Bearer prefix, got %d, body: %s", w.Code, w.Body.String())
	}
}