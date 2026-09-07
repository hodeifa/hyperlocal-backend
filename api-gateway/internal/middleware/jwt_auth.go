package middleware

import (
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

// CustomClaims adalah struktur JWT claims yang digunakan di seluruh sistem.
// Sesuai technical-strategies.md §12:
// - Access token 24 jam
// - Role selalu "customer" untuk Customer App, "driver" untuk Mitra App
type CustomClaims struct {
	UserID string `json:"user_id"`
	Role   string `json:"role"`
	jwt.RegisteredClaims
}

// JWTAuthConfig menyimpan konfigurasi untuk JWT middleware.
type JWTAuthConfig struct {
	SecretKey string
}

// JWTAuthMiddleware mengembalikan Gin middleware yang:
// 1. Extract token dari header "Authorization: Bearer <token>"
// 2. Validasi signature HMAC-SHA256
// 3. Validasi expiry (exp)
// 4. Validasi role ada di claims
// 5. Mengisi user_id dan user_role ke Gin context
//
// Jika validasi gagal → abort 401 dengan body {"error": "..."}
// Jika sukses → panggil SetUserRole(c, role) agar RoleGuard bisa bekerja.
func JWTAuthMiddleware(cfg JWTAuthConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Extract token dari header Authorization
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "missing_authorization_header",
			})
			return
		}

		// Format wajib: "Bearer <token>"
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || !strings.EqualFold(parts[0], "Bearer") {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid_authorization_format",
			})
			return
		}

		tokenString := strings.TrimSpace(parts[1])
		if tokenString == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "empty_token",
			})
			return
		}

		// 2. Parse dan validasi token
		token, err := jwt.ParseWithClaims(tokenString, &CustomClaims{}, func(token *jwt.Token) (interface{}, error) {
			// Validasi signing method — hanya terima HMAC
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, jwt.ErrSignatureInvalid
			}
			return []byte(cfg.SecretKey), nil
		})

		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid_token",
			})
			return
		}

		if !token.Valid {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid_token",
			})
			return
		}

		// 3. Extract claims
		claims, ok := token.Claims.(*CustomClaims)
		if !ok {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid_claims",
			})
			return
		}

		// 4. Validasi field wajib
		if claims.UserID == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "missing_user_id_in_token",
			})
			return
		}

		if claims.Role == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "missing_role_in_token",
			})
			return
		}

		// 5. Validasi role hanya boleh "customer" atau "driver"
		if claims.Role != "customer" && claims.Role != "driver" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "invalid_role_in_token",
			})
			return
		}

		// 6. Isi context Gin — ini yang dibaca oleh RoleGuard
		c.Set("user_id", claims.UserID)
		SetUserRole(c, claims.Role)

		c.Next()
	}
}

// GenerateAccessToken membuat JWT access token baru.
// Dipakai oleh Customer Service dan Driver Service saat verify-otp / login / refresh.
//
// Parameter:
//   - userID: UUID user
//   - role: "customer" atau "driver"
//   - secretKey: dari env JWT_SECRET
//   - expiry: 24 jam (time.Duration)
//
// Return: token string, error
func GenerateAccessToken(userID string, role string, secretKey string, expiry time.Duration) (string, error) {
	now := time.Now()

	claims := &CustomClaims{
		UserID: userID,
		Role:   role,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(now.Add(expiry)),
			IssuedAt:  jwt.NewNumericDate(now),
			NotBefore: jwt.NewNumericDate(now),
			Issuer:    "hyperlocal-backend",
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString([]byte(secretKey))
}