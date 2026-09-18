package v1

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sony/gobreaker"

	"github.com/hodeifa/hyperlocal-backend/api-gateway/config"
	"github.com/hodeifa/hyperlocal-backend/api-gateway/internal/middleware"
	pb "github.com/hodeifa/hyperlocal-backend/proto/customer/v1"
)

// SetupRouter mendaftarkan semua route API v1.
func SetupRouter(
	r *gin.Engine,
	cfg *config.Config,
	customerClient pb.CustomerServiceClient,
	customerCB *gobreaker.CircuitBreaker,
) {
	// Inisialisasi handler dengan dependency gRPC
	authHandler := NewAuthHandler(customerClient, customerCB)

	// Inisialisasi middleware JWT (sesuaikan dengan nama middleware JWT Anda)
	jwtAuth := middleware.JWTAuthMiddleware(middleware.JWTAuthConfig{
    SecretKey: cfg.JWTSecret,
	})

	// =============================================
	// A. PUBLIC ROUTES — /api/v1/auth/customer/
	// =============================================
	authCustomer := r.Group("/api/v1/auth/customer")
	{
		authCustomer.POST("/register", authHandler.Register)
		authCustomer.POST("/verify-otp", handleVerifyOTP)
		authCustomer.POST("/login", handleLogin)
		authCustomer.POST("/refresh", handleRefresh)

		// Protected auth routes (butuh JWT)
		authCustomerProtected := authCustomer.Group("")
		authCustomerProtected.Use(jwtAuth, middleware.RoleGuard())
		authCustomerProtected.POST("/revoke_all", handleRevokeAll)
	}

	// =============================================
	// B. PROTECTED ROUTES — /api/v1/customer/
	// =============================================
	api := r.Group("/api/v1")
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