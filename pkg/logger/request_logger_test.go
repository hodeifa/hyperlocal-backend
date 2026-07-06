package logger_test

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap/zapcore"

	"github.com/hodeifa/hyperlocal-backend/pkg/logger"
	"github.com/hodeifa/hyperlocal-backend/pkg/middleware"
)

func TestRequestLoggerMiddleware_Integration(t *testing.T) {
	gin.SetMode(gin.TestMode)
	buf := &bytes.Buffer{}
	
	log := logger.NewLogger(logger.Config{
		ServiceName:  "integration-test",
		IsProduction: true,
		Level:        zapcore.InfoLevel,
		Output:       buf,
	})

	router := gin.New()
	// URUTAN WAJIB (Sesuai DoD & Technical Strategies)
	router.Use(middleware.ClientInfoMiddleware())
	router.Use(middleware.RequestIDMiddleware())
	router.Use(logger.RequestLoggerMiddleware(log))

	router.GET("/test", func(c *gin.Context) {
		c.String(http.StatusOK, "OK")
	})

	// Simulasi Request dari Flutter
	req, _ := http.NewRequest("GET", "/test", nil)
	req.Header.Set("X-App-Version", "1.2.0")
	req.Header.Set("X-Platform", "android")
	req.Header.Set("X-OS-Version", "14")
	req.Header.Set("X-Build", "debug")
	req.Header.Set("X-Request-ID", "custom-uuid-123")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	// Assert Output JSON
	var logOutput map[string]interface{}
	err := json.Unmarshal(buf.Bytes(), &logOutput)
	require.NoError(t, err)

	assert.Equal(t, "custom-uuid-123", logOutput["request_id"])
	assert.Equal(t, "1.2.0", logOutput["app_version"])
	assert.Equal(t, "android", logOutput["platform"])
	assert.Equal(t, "debug", logOutput["build"]) // Terbaca dari context, bukan string kosong!
	assert.Equal(t, float64(200), logOutput["status"])
}