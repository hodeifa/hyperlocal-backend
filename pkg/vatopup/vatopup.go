package vatopup

import (
	"context"
	"time"
)

// Provider is the interface for Virtual Account payment providers.
type Provider interface {
	CreateVA(ctx context.Context, req *CreateVARequest) (*CreateVAResponse, error)
	ValidateSignature(payload []byte, signature string) bool
	ParseCallback(payload []byte) (*CallbackEvent, error)
}

// CreateVARequest represents the request to create a Virtual Account.
type CreateVARequest struct {
	// ExternalID is the internal system ID (e.g., topup-{driver_id}-{timestamp}).
	// Based on erd.md and technical-strategies.md, this field is populated FROM
	// Flip's response (bill_id), not sent TO Flip. If Flip's API actually supports
	// an external reference field in the creation payload, this should be mapped.
	// Currently ignored by the Flip provider.
	ExternalID string
	// Name is the display name for the VA (e.g., "Hyperlocal Top-Up")
	Name string
	// BankCode is the bank code (e.g., "bca", "mandiri")
	BankCode string
	// Amount in IDR
	Amount int64
	// ExpiryHours is the expiry duration in hours
	ExpiryHours int
}

// CreateVAResponse represents the response from the VA provider.
type CreateVAResponse struct {
	// ExpiresAt is the expiry time (parsed from provider's response with WIB timezone)
	ExpiresAt time.Time
	// BillID is the unique ID from provider (e.g., bill_id from Flip)
	BillID string
	// AccountNumber is the VA number to display to the mitra
	AccountNumber string
	// BankName is the bank name
	BankName string
}

// CallbackEvent represents a parsed webhook event.
type CallbackEvent struct {
	// EventID is the unique event ID from provider (for idempotency/anti-replay)
	EventID string
	// BillID is the VA ID from provider (maps to ExternalID in our DB)
	BillID string
	// Status is the raw status from provider ("SUCCESSFUL", "PENDING", "FAILED", "EXPIRED")
	Status string
	// AmountReceived is the net amount received by provider
	AmountReceived int64
	// Fee is the admin fee deducted by provider
	Fee int64
}
