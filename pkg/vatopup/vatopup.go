package vatopup

import (
	"context"
	"time"
)

// CreateVARequest merepresentasikan request untuk membuat Virtual Account.
// Field diselaraskan dengan pemanggilan di technical-strategies.md §7 (CreateTopupRequest).
type CreateVARequest struct {
	ExternalID     string
	IdempotencyKey string
	Name           string
	BankCode       string
	Amount         int64
	ExpiryHours    int
}

// CreateVAResponse merepresentasikan response setelah VA berhasil dibuat.
type CreateVAResponse struct {
	ExpiresAt     time.Time
	BillID        string
	AccountNumber string
	BankName      string
}

// CallbackEvent merepresentasikan event webhook yang sudah di-parse.
//
// ⚠️ PENTING: Field Status berisi status MENTAH dari Flip
// (SUCCESSFUL, FAILED, PENDING, EXPIRED).
// JANGAN mapping ke status internal DB (PAID, EXPIRED, dll) di layer ini.
// Mapping ke status internal (topup_status ENUM: INIT/PENDING/PAID/EXPIRED/
// CANCELLED/FAILED/UNDERPAID/OVERPAID) adalah tanggung jawab Usecase/Handler
// agar guard `if event.Status != "SUCCESSFUL"` di technical-strategies.md
// bekerja dengan benar.
type CallbackEvent struct {
	EventID        string
	BillID         string
	Status         string
	AmountReceived int64
	Fee            int64
}

// Provider mendefinisikan kontrak untuk provider Virtual Account.
type Provider interface {
	CreateVA(ctx context.Context, req *CreateVARequest) (*CreateVAResponse, error)
	ValidateSignature(payload []byte, signature string) bool
	ParseCallback(payload []byte) (*CallbackEvent, error)
}
