package v1

import (
	"context"
	"errors"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/sony/gobreaker"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	"github.com/hodeifa/hyperlocal-backend/api-gateway/internal/delivery/http/v1/dto"
	"github.com/hodeifa/hyperlocal-backend/pkg/grpcclient"
	"github.com/hodeifa/hyperlocal-backend/pkg/response"
	pb "github.com/hodeifa/hyperlocal-backend/proto/customer/v1"
)

type AuthHandler struct {
	customerClient pb.CustomerServiceClient
	customerCB     *gobreaker.CircuitBreaker
}

func NewAuthHandler(cc pb.CustomerServiceClient, cb *gobreaker.CircuitBreaker) *AuthHandler {
	return &AuthHandler{customerClient: cc, customerCB: cb}
}

// formatValidationError memetakan error validator Gin ke pesan user-friendly.
// WAJIB: jangan pernah return err.Error() mentah ke client — membocorkan
// nama struct, field, dan tag internal (information leak).
func formatValidationError(err error) string {
	var ve validator.ValidationErrors
	if errors.As(err, &ve) {
		for _, fe := range ve {
			switch fe.Tag() {
			case "required":
				return "field wajib diisi"
			case "e164":
				return "nomor HP harus format internasional (contoh: +628123456789)"
			default:
				return "format input tidak valid"
			}
		}
	}
	return "format request tidak valid"
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req dto.RegisterRequestDTO
	if err := c.ShouldBindJSON(&req); err != nil {
		// [FIX] Sanitasi error — jangan bocorkan detail internal
		response.Error(c, http.StatusBadRequest, formatValidationError(err), "invalid_input")
		return
	}

	// Validasi business rule: prefix +62 (Product Requirement)
	if !strings.HasPrefix(req.PhoneNumber, "+62") {
		response.Error(c, http.StatusBadRequest, "nomor HP wajib diawali +62", "invalid_prefix")
		return
	}

	// Inject IP ke gRPC metadata
	md := metadata.Pairs("x-client-ip", c.ClientIP())
	ctx := metadata.NewOutgoingContext(c.Request.Context(), md)
	grpcCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
	defer cancel()

	var resp *pb.RegisterResponse

	// Bungkus dengan Circuit Breaker
	_, err := h.customerCB.Execute(func() (interface{}, error) {
		var callErr error
		resp, callErr = h.customerClient.Register(grpcCtx, &pb.RegisterRequest{
			PhoneNumber: req.PhoneNumber,
		})
		return nil, callErr
	})

	if err != nil {
		if grpcclient.IsCircuitOpen(err) {
			response.Error(c, http.StatusServiceUnavailable, "layanan auth sedang tidak tersedia", "service_unavailable")
			return
		}

		st, ok := status.FromError(err)
		if ok {
			switch st.Code() {
			case codes.ResourceExhausted:
				response.Error(c, http.StatusTooManyRequests, "terlalu banyak permintaan OTP", "rate_limited")
			case codes.InvalidArgument:
				response.Error(c, http.StatusBadRequest, st.Message(), "invalid_input")
			default:
				// [FIX] Jangan bocorkan st.Message() mentah — bisa berisi stack trace/detail internal
				response.Error(c, http.StatusInternalServerError, "terjadi kesalahan internal", "internal_error")
			}
			return
		}
		response.Error(c, http.StatusInternalServerError, "gagal menghubungi service auth", "internal_error")
		return
	}

	response.Success(c, http.StatusOK, resp.Message, map[string]bool{
		"has_driver_role": resp.HasDriverRole,
	})
}