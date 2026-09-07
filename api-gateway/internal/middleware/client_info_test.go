package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
)

func TestClientInfoSetsValidValues(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var gotAppVersion string
	var gotPlatform string
	var gotOSVersion string
	var gotBuildType string
	var gotIsDebug bool

	router := gin.New()
	router.Use(ClientInfo())
	router.GET("/ping", func(c *gin.Context) {
		gotAppVersion = GetAppVersion(c)
		gotPlatform = GetPlatform(c)
		gotOSVersion = GetOSVersion(c)
		gotBuildType = GetBuildType(c)
		gotIsDebug = IsDebugBuild(c)

		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set(HeaderAppVersion, "1.2.3")
	req.Header.Set(HeaderPlatform, "ANDROID")
	req.Header.Set(HeaderOSVersion, "14")
	req.Header.Set(HeaderBuild, "debug")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	if gotAppVersion != "1.2.3" {
		t.Fatalf("expected app version 1.2.3, got %s", gotAppVersion)
	}

	if gotPlatform != "android" {
		t.Fatalf("expected platform android, got %s", gotPlatform)
	}

	if gotOSVersion != "14" {
		t.Fatalf("expected os version 14, got %s", gotOSVersion)
	}

	if gotBuildType != "debug" {
		t.Fatalf("expected build debug, got %s", gotBuildType)
	}

	if !gotIsDebug {
		t.Fatal("expected IsDebugBuild to be true")
	}
}

func TestClientInfoDefaultsWhenHeadersMissing(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var gotAppVersion string
	var gotPlatform string
	var gotOSVersion string
	var gotBuildType string
	var gotIsDebug bool

	router := gin.New()
	router.Use(ClientInfo())
	router.GET("/ping", func(c *gin.Context) {
		gotAppVersion = GetAppVersion(c)
		gotPlatform = GetPlatform(c)
		gotOSVersion = GetOSVersion(c)
		gotBuildType = GetBuildType(c)
		gotIsDebug = IsDebugBuild(c)

		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	if gotAppVersion != "" {
		t.Fatalf("expected empty app version, got %s", gotAppVersion)
	}

	if gotPlatform != "unknown" {
		t.Fatalf("expected platform unknown, got %s", gotPlatform)
	}

	if gotOSVersion != "" {
		t.Fatalf("expected empty os version, got %s", gotOSVersion)
	}

	if gotBuildType != "release" {
		t.Fatalf("expected build release, got %s", gotBuildType)
	}

	if gotIsDebug {
		t.Fatal("expected IsDebugBuild to be false")
	}
}

func TestClientInfoNormalizesInvalidValues(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var gotPlatform string
	var gotBuildType string

	router := gin.New()
	router.Use(ClientInfo())
	router.GET("/ping", func(c *gin.Context) {
		gotPlatform = GetPlatform(c)
		gotBuildType = GetBuildType(c)

		c.Status(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/ping", nil)
	req.Header.Set(HeaderPlatform, "web")
	req.Header.Set(HeaderBuild, "staging")

	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", w.Code)
	}

	if gotPlatform != "unknown" {
		t.Fatalf("expected platform unknown, got %s", gotPlatform)
	}

	if gotBuildType != "release" {
		t.Fatalf("expected build release, got %s", gotBuildType)
	}
}