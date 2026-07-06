package middleware

import (
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

const CtxRequestID = "request_id"
const HeaderRequestID = "X-Request-ID"

// RequestIDMiddleware menghasilkan UUID unik per request jika tidak ada di header.
// WAJIB didaftarkan setelah ClientInfoMiddleware, sebelum Logger.
func RequestIDMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		requestID := c.GetHeader(HeaderRequestID)
		if requestID == "" {
			requestID = uuid.New().String()
		}
		
		// Set ke context Gin (untuk dibaca oleh pkg/logger)
		c.Set(CtxRequestID, requestID)
		
		// Set ke response header (agar client/frontend bisa mereferensikannya saat komplain)
		c.Header(HeaderRequestID, requestID)
		
		c.Next()
	}
}