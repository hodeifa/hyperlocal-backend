package errors

import "errors"

var (
	// ==========================================
	// General / System Errors
	// ==========================================
	ErrInternal = errors.New("internal server error")

	// ==========================================
	// Auth & OTP Errors (Issue #46)
	// ==========================================
	ErrOTPRateLimitPhone  = errors.New("too many OTP requests for this phone number")
	ErrOTPRateLimitIP     = errors.New("too many OTP requests from this IP address")
	ErrInvalidPhoneFormat = errors.New("invalid phone number format, must be E.164 with +62 prefix")
	ErrInvalidOTP         = errors.New("invalid or expired OTP")
	
	// ==========================================
	// Customer / User Errors
	// ==========================================
	ErrCustomerAlreadyExists = errors.New("customer with this phone number already exists")
	ErrCustomerNotFound      = errors.New("customer not found")
)