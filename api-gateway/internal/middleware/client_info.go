package middleware

import (
	"strings"

	"github.com/gin-gonic/gin"
)

const (
	HeaderAppVersion = "X-App-Version"
	HeaderPlatform   = "X-Platform"
	HeaderOSVersion  = "X-OS-Version"
	HeaderBuild      = "X-Build"

	CtxAppVersion = "client_app_version"
	CtxPlatform   = "client_platform"
	CtxOSVersion  = "client_os_version"
	CtxBuildType  = "client_build_type"
)

// ClientInfo mengekstrak header client info dari Flutter dan menyimpannya ke Gin context.
//
// Aturan penting:
// - Semua header optional.
// - Jangan pernah abort request karena header tidak ada.
// - X-Platform hanya boleh bernilai "android", "ios", atau "unknown".
// - X-Build default harus "release" untuk keamanan.
//
// Middleware ini wajib didaftarkan paling pertama setelah gin.Recovery(),
// karena middleware lain seperti logger dan rate limiter bisa membaca nilai ini.
func ClientInfo() gin.HandlerFunc {
	return func(c *gin.Context) {
		appVersion := strings.TrimSpace(c.GetHeader(HeaderAppVersion))
		osVersion := strings.TrimSpace(c.GetHeader(HeaderOSVersion))

		platform := strings.ToLower(strings.TrimSpace(c.GetHeader(HeaderPlatform)))
		switch platform {
		case "android", "ios":
			// valid, biarkan
		default:
			platform = "unknown"
		}

		build := strings.ToLower(strings.TrimSpace(c.GetHeader(HeaderBuild)))
		if build != "debug" {
			build = "release"
		}

		c.Set(CtxAppVersion, appVersion)
		c.Set(CtxPlatform, platform)
		c.Set(CtxOSVersion, osVersion)
		c.Set(CtxBuildType, build)

		c.Next()
	}
}

// IsDebugBuild mengembalikan true jika request berasal dari build debug.
// Gunakan untuk:
// - skip rate limiter saat local development,
// - menampilkan error yang lebih verbose di environment non-production,
// - skip analytics untuk request debug.
func IsDebugBuild(c *gin.Context) bool {
	build, exists := c.Get(CtxBuildType)
	if !exists {
		return false
	}

	buildStr, ok := build.(string)
	if !ok {
		return false
	}

	return buildStr == "debug"
}

// GetPlatform mengembalikan platform client: "android", "ios", atau "unknown".
func GetPlatform(c *gin.Context) string {
	return getStringFromContext(c, CtxPlatform, "unknown")
}

// GetAppVersion mengembalikan versi aplikasi client.
// Contoh: "1.0.0".
func GetAppVersion(c *gin.Context) string {
	return getStringFromContext(c, CtxAppVersion, "")
}

// GetOSVersion mengembalikan versi OS client.
// Contoh Android: "14".
// Contoh iOS: "17.2".
func GetOSVersion(c *gin.Context) string {
	return getStringFromContext(c, CtxOSVersion, "")
}

// GetBuildType mengembalikan build type: "debug" atau "release".
func GetBuildType(c *gin.Context) string {
	return getStringFromContext(c, CtxBuildType, "release")
}

func getStringFromContext(c *gin.Context, key string, fallback string) string {
	value, exists := c.Get(key)
	if !exists {
		return fallback
	}

	str, ok := value.(string)
	if !ok {
		return fallback
	}

	return str
}