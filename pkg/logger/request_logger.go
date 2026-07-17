package logger

import (
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/hodeifa/hyperlocal-backend/pkg/middleware"
)

// RequestLoggerMiddleware mencatat HTTP request dengan field standar.
// WAJIB didaftarkan SETELAH ClientInfoMiddleware dan RequestIDMiddleware di main.go.
func RequestLoggerMiddleware(logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		// Eksekusi handler
		c.Next()

		// c.GetString() aman dari panic (mengembalikan "" jika key tidak ada/beda tipe)
		logger.Info("http_request",
			zap.String("request_id", c.GetString(middleware.CtxRequestID)),
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.Int("status", c.Writer.Status()),
			zap.Int64("duration_ms", time.Since(start).Milliseconds()),

			// Client Info (Sinkron dengan KB §15)
			zap.String("app_version", c.GetString(middleware.CtxAppVersion)),
			zap.String("platform", c.GetString(middleware.CtxPlatform)),
			zap.String("os_version", c.GetString(middleware.CtxOSVersion)),
			zap.String("build", c.GetString(middleware.CtxBuildType)),
		)
	}
}
