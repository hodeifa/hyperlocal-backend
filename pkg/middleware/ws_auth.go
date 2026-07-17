package middleware

import (
	"context"
	"net/http"

	"github.com/gin-gonic/gin"
)

// ParticipantChecker memvalidasi partisipasi user dalam order.
// Mengembalikan (bool, error) untuk membedakan "bukan partisipan" vs "error infra".
type ParticipantChecker func(ctx context.Context, userID, role, orderID string) (bool, error)

// WSAuthMiddleware memvalidasi JWT dari query param ?token= untuk WebSocket.
func WSAuthMiddleware(jwtSecret string, allowedRoles ...string) gin.HandlerFunc {
	roleMap := make(map[string]bool)
	for _, r := range allowedRoles {
		roleMap[r] = true
	}
	enforceRole := len(allowedRoles) > 0

	return func(c *gin.Context) {
		tokenStr := c.Query("token")
		if tokenStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "missing token"})
			return
		}

		claims, err := validateJWT(tokenStr, jwtSecret)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid or expired token"})
			return
		}

		// Validasi Role (Sesuai blueprint.md)
		if enforceRole && !roleMap[claims.Role] {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "role not allowed for this channel"})
			return
		}

		c.Set(CtxUserID, claims.UserID)
		c.Set(CtxRole, claims.Role)
		c.Next()
	}
}

// OrderParticipantGuard memvalidasi bahwa user adalah partisipan order.
func OrderParticipantGuard(checker ParticipantChecker) gin.HandlerFunc {
	return func(c *gin.Context) {
		orderID := c.Param("order_id")
		if orderID == "" {
			c.Next()
			return
		}

		userID := c.GetString(CtxUserID)
		role := c.GetString(CtxRole)

		isParticipant, err := checker(c.Request.Context(), userID, role, orderID)
		if err != nil {
			// Error infrastruktur (DB down / gRPC timeout) -> 500
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": "failed to verify participation"})
			return
		}
		if !isParticipant {
			// User asing / bukan partisipan -> 403
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{"error": "not a participant"})
			return
		}

		c.Next()
	}
}
