package v1_test

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"go.uber.org/zap"
	"google.golang.org/grpc"

	"github.com/hodeifa/hyperlocal-backend/pkg/cache"
	"github.com/hodeifa/hyperlocal-backend/pkg/grpcclient"
	apperr "github.com/hodeifa/hyperlocal-backend/pkg/errors"
	pbDriver "github.com/hodeifa/hyperlocal-backend/proto/driver/v1"
	usecase "github.com/hodeifa/hyperlocal-backend/services/customer/internal/usecase/v1"
)

// MockDriverServiceClient adalah mock untuk gRPC client Driver Service
type MockDriverServiceClient struct {
	mock.Mock
}

// [FIX 1] Signature variadic parameter WAJIB ...grpc.CallOption agar sesuai interface gRPC
func (m *MockDriverServiceClient) CheckPhoneExists(ctx context.Context, req *pbDriver.CheckPhoneExistsRequest, opts ...grpc.CallOption) (*pbDriver.CheckPhoneExistsResponse, error) {
	args := m.Called(ctx, req)
	if args.Get(0) == nil {
		return nil, args.Error(1)
	}
	return args.Get(0).(*pbDriver.CheckPhoneExistsResponse), args.Error(1)
}

// setupTestEnvironment membuat environment test dengan miniredis
func setupTestEnvironment(t *testing.T) (*miniredis.Miniredis, *cache.Client, *zap.Logger) {
	mr, err := miniredis.Run()
	require.NoError(t, err, "failed to start miniredis")

	// [FIX 3] Parse host dan port dari miniredis address menggunakan net.SplitHostPort
	host, portStr, err := net.SplitHostPort(mr.Addr())
	require.NoError(t, err)

	cacheClient, err := cache.NewClient(cache.Config{
		Host: host,
		Port: portStr,
		DB:   0,
	})
	require.NoError(t, err, "failed to create cache client")

	logger := zap.NewNop() // Silent logger untuk testing

	return mr, cacheClient, logger
}

// TestRegister_Success menguji skenario happy path
func TestRegister_Success(t *testing.T) {
	mr, cacheClient, logger := setupTestEnvironment(t)
	defer mr.Close()
	defer cacheClient.Close()

	mockDriverClient := new(MockDriverServiceClient)
	mockDriverClient.On("CheckPhoneExists", mock.Anything, mock.Anything).
		Return(&pbDriver.CheckPhoneExistsResponse{Exists: false}, nil)

	// [FIX 2] Gunakan CircuitBreaker asli dari pkg/grpcclient, tidak perlu mock struct
	cb := grpcclient.NewCircuitBreaker("test-driver-success")

	uc := usecase.NewAuthUsecase(cacheClient, mockDriverClient, cb, logger)

	ctx := context.Background()
	phone := "+628123456789"
	ip := "127.0.0.1"

	result, err := uc.Register(ctx, phone, ip)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.Equal(t, "OTP berhasil dikirim", result.Message)
	assert.False(t, result.HasDriverRole, "should not have driver role")

	// Verifikasi OTP tersimpan di Redis
	otpKey := fmt.Sprintf("otp:code:%s", phone)
	otpValue, err := mr.Get(otpKey)
	assert.NoError(t, err)
	assert.NotEmpty(t, otpValue, "OTP should be stored in Redis")
	assert.Len(t, otpValue, 6, "OTP should be 6 digits")

	mockDriverClient.AssertExpectations(t)
}

// TestRegister_DualRoleDetected menguji skenario dual-role detection
func TestRegister_DualRoleDetected(t *testing.T) {
	mr, cacheClient, logger := setupTestEnvironment(t)
	defer mr.Close()
	defer cacheClient.Close()

	mockDriverClient := new(MockDriverServiceClient)
	mockDriverClient.On("CheckPhoneExists", mock.Anything, mock.Anything).
		Return(&pbDriver.CheckPhoneExistsResponse{Exists: true}, nil)

	cb := grpcclient.NewCircuitBreaker("test-driver-dual-role")
	uc := usecase.NewAuthUsecase(cacheClient, mockDriverClient, cb, logger)

	ctx := context.Background()
	phone := "+628123456789"
	ip := "127.0.0.1"

	result, err := uc.Register(ctx, phone, ip)

	assert.NoError(t, err)
	assert.NotNil(t, result)
	assert.True(t, result.HasDriverRole, "should have driver role")

	mockDriverClient.AssertExpectations(t)
}

// TestRegister_RateLimitPhone menguji rate limiting per nomor HP
func TestRegister_RateLimitPhone(t *testing.T) {
	mr, cacheClient, logger := setupTestEnvironment(t)
	defer mr.Close()
	defer cacheClient.Close()

	mockDriverClient := new(MockDriverServiceClient)
	mockDriverClient.On("CheckPhoneExists", mock.Anything, mock.Anything).
		Return(&pbDriver.CheckPhoneExistsResponse{Exists: false}, nil)

	cb := grpcclient.NewCircuitBreaker("test-driver-rate-phone")
	uc := usecase.NewAuthUsecase(cacheClient, mockDriverClient, cb, logger)

	ctx := context.Background()
	phone := "+628129999999"
	ip := "127.0.0.1"

	// Request 1-3: Sukses
	for i := 1; i <= 3; i++ {
		result, err := uc.Register(ctx, phone, ip)
		assert.NoError(t, err, "request %d should succeed", i)
		assert.NotNil(t, result)
	}

	// Request 4: Rate limit
	result, err := uc.Register(ctx, phone, ip)
	assert.Error(t, err, "4th request should fail with rate limit")
	assert.ErrorIs(t, err, apperr.ErrOTPRateLimitPhone)
	assert.Nil(t, result)
}

// TestRegister_RateLimitIP menguji rate limiting per IP address
func TestRegister_RateLimitIP(t *testing.T) {
	mr, cacheClient, logger := setupTestEnvironment(t)
	defer mr.Close()
	defer cacheClient.Close()

	mockDriverClient := new(MockDriverServiceClient)
	mockDriverClient.On("CheckPhoneExists", mock.Anything, mock.Anything).
		Return(&pbDriver.CheckPhoneExistsResponse{Exists: false}, nil)

	cb := grpcclient.NewCircuitBreaker("test-driver-rate-ip")
	uc := usecase.NewAuthUsecase(cacheClient, mockDriverClient, cb, logger)

	ctx := context.Background()
	ip := "192.168.1.100"

	// Request 1-10: Sukses (dengan nomor HP berbeda-beda)
	for i := 1; i <= 10; i++ {
		phone := fmt.Sprintf("+628120000%02d", i)
		result, err := uc.Register(ctx, phone, ip)
		assert.NoError(t, err, "request %d should succeed", i)
		assert.NotNil(t, result)
	}

	// Request 11: Rate limit IP
	phone := "+628120000011"
	result, err := uc.Register(ctx, phone, ip)
	assert.Error(t, err, "11th request should fail with IP rate limit")
	assert.ErrorIs(t, err, apperr.ErrOTPRateLimitIP)
	assert.Nil(t, result)
}

// TestRegister_DriverServiceDown menguji fallback saat Driver Service tidak tersedia
func TestRegister_DriverServiceDown(t *testing.T) {
	mr, cacheClient, logger := setupTestEnvironment(t)
	defer mr.Close()
	defer cacheClient.Close()

	mockDriverClient := new(MockDriverServiceClient)
	mockDriverClient.On("CheckPhoneExists", mock.Anything, mock.Anything).
		Return(nil, fmt.Errorf("connection refused"))

	cb := grpcclient.NewCircuitBreaker("test-driver-down")
	uc := usecase.NewAuthUsecase(cacheClient, mockDriverClient, cb, logger)

	ctx := context.Background()
	phone := "+628123456789"
	ip := "127.0.0.1"

	result, err := uc.Register(ctx, phone, ip)

	// Seharusnya tetap sukses dengan fallback has_driver_role=false
	assert.NoError(t, err, "should succeed despite Driver Service being down")
	assert.NotNil(t, result)
	assert.False(t, result.HasDriverRole, "should fallback to false when Driver Service fails")
	assert.Equal(t, "OTP berhasil dikirim", result.Message)

	// Verifikasi OTP tetap tersimpan
	otpKey := fmt.Sprintf("otp:code:%s", phone)
	otpValue, err := mr.Get(otpKey)
	assert.NoError(t, err)
	assert.NotEmpty(t, otpValue, "OTP should still be stored")
}

// TestRegister_OTPExpiry menguji bahwa OTP expire setelah 5 menit
func TestRegister_OTPExpiry(t *testing.T) {
	mr, cacheClient, logger := setupTestEnvironment(t)
	defer mr.Close()
	defer cacheClient.Close()

	mockDriverClient := new(MockDriverServiceClient)
	mockDriverClient.On("CheckPhoneExists", mock.Anything, mock.Anything).
		Return(&pbDriver.CheckPhoneExistsResponse{Exists: false}, nil)

	cb := grpcclient.NewCircuitBreaker("test-driver-expiry")
	uc := usecase.NewAuthUsecase(cacheClient, mockDriverClient, cb, logger)

	ctx := context.Background()
	phone := "+628123456789"
	ip := "127.0.0.1"

	// Register pertama kali
	result, err := uc.Register(ctx, phone, ip)
	assert.NoError(t, err)
	assert.NotNil(t, result)

	// Verifikasi OTP ada
	otpKey := fmt.Sprintf("otp:code:%s", phone)
	otpValue1, err := mr.Get(otpKey)
	assert.NoError(t, err)
	assert.NotEmpty(t, otpValue1)

	// Fast-forward waktu 5 menit + 1 detik (fitur bawaan miniredis)
	mr.FastForward(5*time.Minute + 1*time.Second)

	// OTP seharusnya sudah expire dan dihapus dari Redis
	_, err = mr.Get(otpKey)
	assert.Error(t, err, "OTP should be expired and deleted from miniredis")
}