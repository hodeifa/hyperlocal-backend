package v1

import (
	"context"

	pb "github.com/hodeifa/hyperlocal-backend/proto/driver/v1"
)

type DriverGRPCHandler struct {
	pb.UnimplementedDriverServiceServer
}

func NewDriverGRPCHandler() *DriverGRPCHandler {
	return &DriverGRPCHandler{}
}

func (h *DriverGRPCHandler) CheckPhoneExists(ctx context.Context, req *pb.CheckPhoneExistsRequest) (*pb.CheckPhoneExistsResponse, error) {
	// STUB: Sementara return false. 
	// Nanti di Sprint 5 akan diisi query ke DB: SELECT EXISTS(SELECT 1 FROM drivers WHERE phone_number = $1 AND is_active = TRUE)
	return &pb.CheckPhoneExistsResponse{
		Exists: false,
	}, nil
}