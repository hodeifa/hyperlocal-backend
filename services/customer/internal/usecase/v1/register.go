package v1

import (
	"context"
	"crypto/rand"
	"fmt"
	"math/big"
	"time"

	"github.com/sony/gobreaker"
	"go.uber.org/zap"

	apperr "github.com/hodeifa/hyperlocal-backend/pkg/errors"
	"github.com/hodeifa/hyperlocal-backend/pkg/cache"
	"github.com/hodeifa/hyperlocal-backend/pkg/grpcclient"
	pbDriver "github.com/hodeifa/hyperlocal-backend/proto/driver/v1"
)

type RegisterResponse struct {
	Message       string
	HasDriverRole bool
}

type AuthUsecase struct {
	cache        cache.Cache // Menggunakan interface yang sudah ada di pkg/cache
	driverClient pbDriver.DriverServiceClient
	driverCB     *gobreaker.CircuitBreaker
	logger       *zap.Logger
}

func NewAuthUsecase(c cache.Cache, drv pbDriver.DriverServiceClient, cb *gobreaker.CircuitBreaker, log *zap.Logger) *AuthUsecase {
	return &AuthUsecase{cache: c, driverClient: drv, driverCB: cb, logger: log}
}

func (u *AuthUsecase) Register(ctx context.Context, phone, ip string) (*RegisterResponse, error) {
	// 1. Dual Rate Limiting (via pkg/cache)
	if err := u.enforceRateLimit(ctx, phone, ip); err != nil {
		return nil, err
	}

	// 2. Generate & Store OTP (Stateless: TIDAK insert ke DB customers)
	otp, err := generateSecureOTP()
	if err != nil {
		return nil, apperr.ErrInternal
	}
	otpKey := fmt.Sprintf("otp:code:%s", phone)
	// Menggunakan SetEX dari pkg/cache (sudah ada di repo Anda)
	if err := u.cache.SetEX(ctx, otpKey, otp, 5*time.Minute); err != nil {
		return nil, apperr.ErrInternal
	}

	// 3. Check Dual-Role via gRPC + Circuit Breaker
	hasDriverRole := false
	_, err = u.driverCB.Execute(func() (interface{}, error) {
		grpcCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
		defer cancel()
		
		resp, callErr := u.driverClient.CheckPhoneExists(grpcCtx, &pbDriver.CheckPhoneExistsRequest{
			PhoneNumber: phone,
		})
		if callErr != nil {
			return nil, callErr
		}
		hasDriverRole = resp.Exists
		return nil, nil
	})

	maskedPhone := maskPhone(phone)
	if err != nil {
		// Menggunakan helper IsCircuitOpen dari pkg/grpcclient (sudah ada di repo Anda)
		if grpcclient.IsCircuitOpen(err) {
			u.logger.Warn("Circuit breaker open for Driver Service", zap.String("phone_masked", maskedPhone))
		} else {
			u.logger.Error("gRPC to Driver Service failed", zap.String("error", err.Error()), zap.String("phone_masked", maskedPhone))
		}
		// Fail-safe: fallback ke false
	}

	// 4. Simulate SMS (Sprint 4) - UNMASKED untuk kemudahan QA
	u.logger.Info("📱 SIMULATED SMS OTP", zap.String("phone", phone), zap.String("otp", otp))

	return &RegisterResponse{
		Message:       "OTP berhasil dikirim",
		HasDriverRole: hasDriverRole,
	}, nil
}

func (u *AuthUsecase) enforceRateLimit(ctx context.Context, phone, ip string) error {
	// Asumsi: Anda sudah menambahkan method Incr() dan Expire() ke interface cache.Cache di pkg/cache
	phoneKey := fmt.Sprintf("otp:phone:%s", phone)
	phoneCount, err := u.cache.Incr(ctx, phoneKey)
	if err == nil {
		if phoneCount == 1 {
			u.cache.Expire(ctx, phoneKey, 10*time.Minute)
		}
		if phoneCount > 3 {
			return apperr.ErrOTPRateLimitPhone
		}
	}

	ipKey := fmt.Sprintf("otp:ip:%s", ip)
	ipCount, err := u.cache.Incr(ctx, ipKey)
	if err == nil {
		if ipCount == 1 {
			u.cache.Expire(ctx, ipKey, time.Hour)
		}
		if ipCount > 10 {
			return apperr.ErrOTPRateLimitIP
		}
	}
	return nil
}

func generateSecureOTP() (string, error) {
	const digits = "0123456789"
	ret := make([]byte, 6)
	for i := 0; i < 6; i++ {
		num, err := rand.Int(rand.Reader, big.NewInt(int64(len(digits))))
		if err != nil {
			return "", err
		}
		ret[i] = digits[num.Int64()]
	}
	return string(ret), nil
}

func maskPhone(phone string) string {
	if len(phone) < 7 {
		return phone
	}
	return phone[:4] + "****" + phone[len(phone)-3:]
}