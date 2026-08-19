// Package middleware provides HTTP and WebSocket middleware for the API Gateway and services.
package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// Context keys for client information extracted from headers.
// Use these constants instead of string literals when reading from the Gin context.
const (
	CtxAppVersion = "client_app_version"
	CtxPlatform   = "client_platform"
	CtxOSVersion  = "client_os_version"
	CtxBuildType  = "client_build_type"
)

// ClientInfoMiddleware extracts client info headers and stores them in the Gin context.
// It should be registered before other middleware so that the context is populated early.
// All headers are optional; the request is never aborted if they are missing.
func ClientInfoMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		platform := strings.ToLower(c.GetHeader("X-Platform"))
		if platform != "android" && platform != "ios" {
			platform = "unknown"
		}

		build := strings.ToLower(c.GetHeader("X-Build"))
		if build != "debug" {
			build = "release" // default to release for security
		}

		c.Set(CtxAppVersion, c.GetHeader("X-App-Version"))
		c.Set(CtxPlatform, platform)
		c.Set(CtxOSVersion, c.GetHeader("X-OS-Version"))
		c.Set(CtxBuildType, build)
		c.Next()
	}
}

// IsDebugBuild returns true if the request originates from a debug build.
// This can be used to skip rate limiters, return verbose errors, or skip analytics.
func IsDebugBuild(c *gin.Context) bool {
	build, _ := c.Get(CtxBuildType)
	return build == "debug"
}

// GetPlatform returns the client platform ("android", "ios", or "unknown").
func GetPlatform(c *gin.Context) string {
	if p, ok := c.Get(CtxPlatform); ok {
		if s, ok := p.(string); ok {
			return s
		}
	}
	return "unknown"
}
