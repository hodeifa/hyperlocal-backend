package middleware

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
)

func equalStringSlice(t *testing.T, got []string, want []string) bool {
	t.Helper()

	if len(got) != len(want) {
		return false
	}

	for i := range got {
		if got[i] != want[i] {
			return false
		}
	}

	return true
}

func TestSortedPrefixesFromLongestFirst(t *testing.T) {
	roles := map[string][]string{
		"/api/v1/":                         {"customer", "driver"},
		"/api/v1/customer/":                {"customer"},
		"/api/v1/auth/customer/revoke_all": {"customer"},
	}

	got := sortedPrefixesFrom(roles)

	want := []string{
		"/api/v1/auth/customer/revoke_all",
		"/api/v1/customer/",
		"/api/v1/",
	}

	if !equalStringSlice(t, got, want) {
		t.Fatalf("sortedPrefixesFrom() = %v, want %v", got, want)
	}
}

func TestFindAllowedRolesCurrentPR(t *testing.T) {
	tests := []struct {
		name      string
		path      string
		wantMatch bool
		wantRoles []string
	}{
		{
			name:      "customer profile",
			path:      "/api/v1/customer/profile",
			wantMatch: true,
			wantRoles: []string{"customer"},
		},
		{
			name:      "customer addresses",
			path:      "/api/v1/customer/addresses",
			wantMatch: true,
			wantRoles: []string{"customer"},
		},
		{
			name:      "customer address by id",
			path:      "/api/v1/customer/addresses/:id",
			wantMatch: true,
			wantRoles: []string{"customer"},
		},
		{
			name:      "customer revoke all",
			path:      "/api/v1/auth/customer/revoke_all",
			wantMatch: true,
			wantRoles: []string{"customer"},
		},
		{
			name:      "driver profile should not match",
			path:      "/api/v1/driver/profile",
			wantMatch: false,
			wantRoles: nil,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			roles, matched := findAllowedRoles(tt.path, routeRoles, sortedRoutePrefixes)

			if matched != tt.wantMatch {
				t.Fatalf("findAllowedRoles() matched = %v, want %v", matched, tt.wantMatch)
			}

			if !equalStringSlice(t, roles, tt.wantRoles) {
				t.Fatalf("findAllowedRoles() roles = %v, want %v", roles, tt.wantRoles)
			}
		})
	}
}

func TestFindAllowedRolesOverlappingRoutes(t *testing.T) {
	roles := map[string][]string{
		"/api/v1/orders/":           {"customer", "driver"},
		"/api/v1/orders/:id/accept": {"driver"},
	}

	sorted := sortedPrefixesFrom(roles)

	// Route spesifik harus menang terhadap prefix umum.
	got, matched := findAllowedRoles("/api/v1/orders/:id/accept", roles, sorted)
	if !matched {
		t.Fatal("expected /api/v1/orders/:id/accept to match")
	}

	want := []string{"driver"}
	if !equalStringSlice(t, got, want) {
		t.Fatalf("expected accept route to be driver-only, got %v", got)
	}

	// Route umum tetap memakai prefix /api/v1/orders/.
	got, matched = findAllowedRoles("/api/v1/orders/:id", roles, sorted)
	if !matched {
		t.Fatal("expected /api/v1/orders/:id to match")
	}

	want = []string{"customer", "driver"}
	if !equalStringSlice(t, got, want) {
		t.Fatalf("expected generic order route to allow customer+driver, got %v", got)
	}
}

func setupRoleGuardRouter(route string, role string, includeRole bool) *gin.Engine {
	gin.SetMode(gin.TestMode)

	router := gin.New()

	handlers := []gin.HandlerFunc{}

	if includeRole {
		handlers = append(handlers, func(c *gin.Context) {
			SetUserRole(c, role)
		})
	}

	handlers = append(handlers, RoleGuard(), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	router.GET(route, handlers...)

	return router
}

func TestRoleGuardAllowsCustomer(t *testing.T) {
	router := setupRoleGuardRouter("/api/v1/customer/profile", "customer", true)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/customer/profile", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d, body: %s", w.Code, w.Body.String())
	}
}

func TestRoleGuardRejectsDriverOnCustomerRoute(t *testing.T) {
	router := setupRoleGuardRouter("/api/v1/customer/profile", "driver", true)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/customer/profile", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d, body: %s", w.Code, w.Body.String())
	}

	if !strings.Contains(w.Body.String(), `"role_not_allowed"`) {
		t.Fatalf("expected body contains role_not_allowed, got %s", w.Body.String())
	}
}

func TestRoleGuardRejectsMissingRole(t *testing.T) {
	router := setupRoleGuardRouter("/api/v1/customer/profile", "", false)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/customer/profile", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusUnauthorized {
		t.Fatalf("expected 401, got %d, body: %s", w.Code, w.Body.String())
	}
}

func TestRoleGuardRejectsUnregisteredRoute(t *testing.T) {
	router := setupRoleGuardRouter("/api/v1/unknown", "customer", true)

	req := httptest.NewRequest(http.MethodGet, "/api/v1/unknown", nil)
	w := httptest.NewRecorder()

	router.ServeHTTP(w, req)

	if w.Code != http.StatusForbidden {
		t.Fatalf("expected 403, got %d, body: %s", w.Code, w.Body.String())
	}

	if !strings.Contains(w.Body.String(), `"route_not_protected"`) {
		t.Fatalf("expected body contains route_not_protected, got %s", w.Body.String())
	}
}