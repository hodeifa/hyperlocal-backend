package v1

import (
	"context"
	"errors"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/metadata"
	"google.golang.org/grpc/status"

	apperr "github.com/hodeifa/hyperlocal-backend/pkg/errors"
	pb "github.com/hodeifa/hyperlocal-backend/proto/customer/v1"
	usecase "github.com/hodeifa/hyperlocal-backend/services/customer/internal/usecase/v1"
)

type AuthGRPCHandler struct {
	pb.UnimplementedCustomerServiceServer
	usecase *usecase.AuthUsecase
}

func NewAuthGRPCHandler(uc *usecase.AuthUsecase) *AuthGRPCHandler {
	return &AuthGRPCHandler{usecase: uc}
}

func (h *AuthGRPCHandler) Register(ctx context.Context, req *pb.RegisterRequest) (*pb.RegisterResponse, error) {
	var ip string
	if md, ok := metadata.FromIncomingContext(ctx); ok {
		if ips := md.Get("x-client-ip"); len(ips) > 0 {
			ip = ips[0]
		}
	}

	res, err := h.usecase.Register(ctx, req.PhoneNumber, ip)
	if err != nil {
		switch {
		case errors.Is(err, apperr.ErrOTPRateLimitPhone), errors.Is(err, apperr.ErrOTPRateLimitIP):
			return nil, status.Error(codes.ResourceExhausted, err.Error())
		default:
			return nil, status.Error(codes.Internal, err.Error())
		}
	}

	return &pb.RegisterResponse{
		Message:       res.Message,
		HasDriverRole: res.HasDriverRole,
	}, nil
}