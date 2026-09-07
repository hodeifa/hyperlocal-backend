package v1

import (
	"log"
	"net/http"

	"github.com/gin-gonic/gin"

	// [FIX] Gunakan prefix github.com/hodeifa/ sesuai go.mod
	"github.com/hodeifa/hyperlocal-backend/api-gateway/config"
	"github.com/hodeifa/hyperlocal-backend/api-gateway/internal/middleware"
)

// SetupRouter mendaftarkan semua route API v1.
func SetupRouter(r *gin.Engine, cfg config.Config) {
	if cfg.JWTSecret == "" {
		log.Println("⚠️  WARNING: JWT_SECRET is empty. JWT middleware will reject all tokens.")
	}

	jwtAuth := middleware.JWTAuthMiddleware(middleware.JWTAuthConfig{
		SecretKey: cfg.JWTSecret,
	})

	api := r.Group("/api/v1")

	// A. PUBLIC ROUTES — /api/v1/auth/customer/
	authCustomer := api.Group("/auth/customer")
	{
		authCustomer.POST("/register", handleRegister)
		authCustomer.POST("/verify-otp", handleVerifyOTP)
		authCustomer.POST("/login", handleLogin)
		authCustomer.POST("/refresh", handleRefresh)

		authCustomerProtected := authCustomer.Group("")
		authCustomerProtected.Use(jwtAuth, middleware.RoleGuard())
		authCustomerProtected.POST("/revoke_all", handleRevokeAll)
	}

	// B. PROTECTED ROUTES — /api/v1/customer/
	customer := api.Group("/customer")
	customer.Use(jwtAuth, middleware.RoleGuard())
	{
		customer.GET("/profile", handleGetProfile)
		customer.PUT("/profile", handleUpdateProfile)
		customer.GET("/addresses", handleGetAddresses)
		customer.POST("/addresses", handleCreateAddress)
		customer.DELETE("/addresses/:id", handleDeleteAddress)
	}
}
// =============================================
// STUB HANDLERS — akan diganti dengan implementasi nyata
// di sprint berikutnya. Untuk sekarang return 501.
// =============================================

func handleRegister(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error":   "not_implemented",
		"message": "Handler register belum diimplementasikan",
	})
}

func handleVerifyOTP(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error":   "not_implemented",
		"message": "Handler verify-otp belum diimplementasikan",
	})
}

func handleLogin(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error":   "not_implemented",
		"message": "Handler login belum diimplementasikan",
	})
}

func handleRefresh(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error":   "not_implemented",
		"message": "Handler refresh belum diimplementasikan",
	})
}

func handleRevokeAll(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error":   "not_implemented",
		"message": "Handler revoke_all belum diimplementasikan",
	})
}

func handleGetProfile(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error":   "not_implemented",
		"message": "Handler get profile belum diimplementasikan",
	})
}

func handleUpdateProfile(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error":   "not_implemented",
		"message": "Handler update profile belum diimplementasikan",
	})
}

func handleGetAddresses(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error":   "not_implemented",
		"message": "Handler get addresses belum diimplementasikan",
	})
}

func handleCreateAddress(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error":   "not_implemented",
		"message": "Handler create address belum diimplementasikan",
	})
}

func handleDeleteAddress(c *gin.Context) {
	c.JSON(http.StatusNotImplemented, gin.H{
		"error":   "not_implemented",
		"message": "Handler delete address belum diimplementasikan",
	})
}