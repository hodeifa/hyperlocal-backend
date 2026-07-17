// Package middleware provides HTTP middleware components for the API Gateway.
package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// Context keys — Konstanta ini yang menjadi "kontrak" dengan pkg/logger
const (
	CtxAppVersion = "client_app_version"
	CtxPlatform   = "client_platform"
	CtxOSVersion  = "client_os_version"
	CtxBuildType  = "client_build_type"
)

// ClientInfoMiddleware mengekstrak header klien dan menyimpannya ke context Gin.
// WAJIB didaftarkan paling pertama di router.
func ClientInfoMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		platform := strings.ToLower(c.GetHeader("X-Platform"))
		if platform != "android" && platform != "ios" {
			platform = "unknown"
		}

		build := strings.ToLower(c.GetHeader("X-Build"))
		if build != "debug" {
			build = "release" // default ke release untuk keamanan
		}

		c.Set(CtxAppVersion, c.GetHeader("X-App-Version"))
		c.Set(CtxPlatform, platform)
		c.Set(CtxOSVersion, c.GetHeader("X-OS-Version"))
		c.Set(CtxBuildType, build)

		c.Next()
	}
}

// IsDebugBuild returns true if the request originates from a debug build.
func IsDebugBuild(c *gin.Context) bool {
	build, _ := c.Get(CtxBuildType)
	return build == "debug"
}

// GetPlatform returns the client platform (e.g., "android", "ios", or "unknown").
func GetPlatform(c *gin.Context) string {
	if p, ok := c.Get(CtxPlatform); ok {
		if s, ok := p.(string); ok {
			return s
		}
	}
	return "unknown"
}
