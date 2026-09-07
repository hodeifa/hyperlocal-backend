package middleware

import (
	"net/http"
	"sort"
	"strings"

	"github.com/gin-gonic/gin"
)

// CtxUserRole adalah key Gin context untuk role user.
// JWT middleware wajib mengisi nilai ini setelah token valid.
//
// Contoh:
//
//	middleware.SetUserRole(c, claims.Role)
const CtxUserRole = "user_role"

// routeRoles adalah daftar route yang dilindungi berdasarkan role.
//
// UNTUK PR INI, hanya masukkan route yang benar-benar terdaftar:
// - /api/v1/customer/* hanya untuk customer
// - /api/v1/auth/customer/revoke_all hanya untuk customer
//
// Jangan menambahkan route spekulatif yang belum ada handler/service-nya.
var routeRoles = map[string][]string{
	"/api/v1/customer/": {
		"customer",
	},
	"/api/v1/auth/customer/revoke_all": {
		"customer",
	},
}

// sortedRoutePrefixes dibuat sekali saat startup.
// Urutan prefix di-sort berdasarkan panjang terpanjang agar matching deterministik.
//
// Contoh:
// - "/api/v1/orders/:id/accept" harus dicek sebelum "/api/v1/orders/"
var sortedRoutePrefixes = sortedPrefixesFrom(routeRoles)

// RoleGuard memvalidasi apakah role user boleh mengakses route saat ini.
//
// Aturan:
// - Jika role tidak ada di context, kembalikan 401.
// - Jika route tidak terdaftar di routeRoles, kembalikan 403.
// - Jika role tidak diizinkan, kembalikan 403 dengan body:
//
//	{"error": "role_not_allowed"}
//
// Middleware ini wajib dipasang SETELAH JWTAuth.
func RoleGuard() gin.HandlerFunc {
	return func(c *gin.Context) {
		role := strings.TrimSpace(c.GetString(CtxUserRole))
		if role == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "unauthorized",
			})
			return
		}

		// c.FullPath() mengembalikan pola route Gin yang terdaftar.
		// Contoh:
		// - /api/v1/customer/profile
		// - /api/v1/customer/addresses/:id
		// - /api/v1/auth/customer/revoke_all
		path := c.FullPath()

		// Fallback defensif.
		// Dalam kondisi normal Gin route harus punya FullPath.
		if path == "" {
			path = c.Request.URL.Path
		}

		allowedRoles, matched := findAllowedRoles(path, routeRoles, sortedRoutePrefixes)
		if !matched || len(allowedRoles) == 0 {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "route_not_protected",
			})
			return
		}

		if !containsRole(allowedRoles, role) {
			c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
				"error": "role_not_allowed",
			})
			return
		}

		c.Next()
	}
}

// SetUserRole mengisi role user ke Gin context.
//
// Fungsi ini wajib dipanggil oleh JWTAuth middleware setelah token valid.
// Contoh pemakaian di JWTAuth:
//
//	claims, err := validateJWT(token)
//	if err != nil {
//		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{"error": "invalid_token"})
//		return
//	}
//
//	middleware.SetUserRole(c, claims.Role)
//	c.Next()
func SetUserRole(c *gin.Context, role string) {
	c.Set(CtxUserRole, role)
}

// sortedPrefixesFrom mengurutkan key routeRoles berdasarkan panjang descending.
// Jika panjang sama, urutkan secara lexicographic agar deterministik.
func sortedPrefixesFrom(roles map[string][]string) []string {
	keys := make([]string, 0, len(roles))

	for key := range roles {
		keys = append(keys, key)
	}

	sort.Slice(keys, func(i, j int) bool {
		if len(keys[i]) != len(keys[j]) {
			return len(keys[i]) > len(keys[j])
		}

		return keys[i] < keys[j]
	})

	return keys
}

// findAllowedRoles mencari role yang diizinkan untuk path tertentu.
//
// Matching dilakukan menggunakan prefix terpanjang yang sudah di-sort.
// Ini mencegah perilaku non-deterministik akibat iterasi map Go.
func findAllowedRoles(
	path string,
	roles map[string][]string,
	sortedKeys []string,
) ([]string, bool) {
	if path == "" {
		return nil, false
	}

	for _, key := range sortedKeys {
		if strings.HasPrefix(path, key) {
			return roles[key], true
		}
	}

	return nil, false
}

func containsRole(roles []string, role string) bool {
	for _, r := range roles {
		if r == role {
			return true
		}
	}

	return false
}