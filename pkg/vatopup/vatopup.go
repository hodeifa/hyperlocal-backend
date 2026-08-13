package vatopup

import (
	"context"
	"time"
)

// CreateVARequest merepresentasikan request untuk membuat Virtual Account.
// Field diselaraskan dengan pemanggilan di technical-strategies.md §7 (CreateTopupRequest).
type CreateVARequest struct {
	// ExternalID adalah ID unik internal (misal: topup-{driver_id}-{timestamp})
	// yang digenerate Usecase.
	//
	// ⚠️ CATATAN SINKRONISASI: Untuk saat ini, field ini TIDAK dikirim ke payload
	// API Flip. Dokumentasi publik Flip (Accept Payment/Bills) belum secara eksplisit
	// dikonfirmasi mendukung parameter referensi eksternal kustom di endpoint create bill.
	// Jika di kemudian hari ada field yang sesuai (misal: reference_id), field ini
	// bisa dipetakan ke sana.
	//
	// Usecase Sprint 11 tetap membutuhkannya di struct ini untuk tracking internal
	// SEBELUM respons Flip diterima. Setelah Flip merespons, Usecase memetakan
	// va.BillID (respons Flip) ke kolom driver_topup_requests.external_id di database —
	// BUKAN req.ExternalID ini.
	ExternalID string

	// IdempotencyKey adalah UUIDv7 dari client untuk header Idempotency-Key ke Flip.
	// Saga Pattern: mencegah dua VA tercipta akibat double-tap.
	IdempotencyKey string

	// Name adalah judul/nama VA (akan dikirim sebagai "title" ke Flip).
	// technical-strategies.md menggunakan field ini, bukan "Title".
	Name string

	Amount      int64
	ExpiryHours int
	BankCode    string
}

// CreateVAResponse merepresentasikan response setelah VA berhasil dibuat.
type CreateVAResponse struct {
	BillID        string // bill_id dari Flip — dipetakan ke external_id di DB oleh Usecase
	AccountNumber string // Nomor VA (mapped dari account_number Flip)
	BankName      string
	ExpiresAt     time.Time // Di-parse langsung dari response Flip (WIB)
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
	EventID        string // id dari payload Flip (untuk processed_flip_webhooks)
	BillID         string // bill_id dari payload Flip
	AmountReceived int64  // Nominal bersih setelah fee (amount)
	Fee            int64
	Status         string // Status MENTAH dari Flip: SUCCESSFUL, FAILED, PENDING, EXPIRED
}

// Provider mendefinisikan kontrak untuk provider Virtual Account.
type Provider interface {
	CreateVA(ctx context.Context, req *CreateVARequest) (*CreateVAResponse, error)
	ValidateSignature(payload []byte, signature string) bool
	ParseCallback(payload []byte) (*CallbackEvent, error)
}
