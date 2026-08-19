package vatopup

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestFlipProvider_ValidateSignature(t *testing.T) {
	secret := "test-secret-key"
	provider := NewFlipProvider("http://localhost", "api-key", secret)

	payload := []byte(`{"id":"123","bill_id":"456","amount":50000,"fee":2500,"status":"SUCCESSFUL"}`)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	validSignature := hex.EncodeToString(mac.Sum(nil))

	// [FIX FIELALIGNMENT]
	// Urutkan field string (16 bytes) terlebih dahulu, baru slice (24 bytes), lalu bool.
	// Ini memadatkan pointer di awal struct sehingga GC hanya perlu memindai 40 byte
	// (bukan 48 byte), memuaskan linter fieldalignment.
	tests := []struct {
		name      string
		signature string
		payload   []byte
		expected  bool
	}{
		{"Valid Signature", validSignature, payload, true},
		{"Invalid Signature", "invalid-signature", payload, false},
		{"Tampered Payload", validSignature, []byte(`{"id":"123","bill_id":"456","amount":50000,"fee":2500,"status":"FAILED"}`), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := provider.ValidateSignature(tt.payload, tt.signature); got != tt.expected {
				t.Errorf("ValidateSignature() = %v, want %v", got, tt.expected)
			}
		})
	}
}

func TestFlipProvider_ParseCallback(t *testing.T) {
	provider := NewFlipProvider("http://localhost", "api-key", "secret")

	tests := []struct {
		name           string
		payload        string
		expectedStatus string // Expect RAW status — bukan PAID/EXPIRED
		expectErr      bool
	}{
		{
			name:           "Success Status",
			payload:        `{"id":"evt_1","bill_id":"bill_1","amount":47500,"fee":2500,"status":"SUCCESSFUL"}`,
			expectedStatus: "SUCCESSFUL", // RAW
		},
		{
			name:           "Failed Status",
			payload:        `{"id":"evt_2","bill_id":"bill_2","amount":0,"fee":0,"status":"FAILED"}`,
			expectedStatus: "FAILED", // RAW
		},
		{
			name:           "Pending Status",
			payload:        `{"id":"evt_3","bill_id":"bill_3","amount":0,"fee":0,"status":"PENDING"}`,
			expectedStatus: "PENDING", // RAW
		},
		{
			name:           "Expired Status",
			payload:        `{"id":"evt_4","bill_id":"bill_4","amount":0,"fee":0,"status":"EXPIRED"}`,
			expectedStatus: "EXPIRED", // RAW
		},
		{
			name:      "Invalid JSON",
			payload:   `{"invalid_json"`,
			expectErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			event, err := provider.ParseCallback([]byte(tt.payload))
			if (err != nil) != tt.expectErr {
				t.Errorf("ParseCallback() error = %v, expectErr %v", err, tt.expectErr)
				return
			}
			if !tt.expectErr && event.Status != tt.expectedStatus {
				t.Errorf("ParseCallback() status = %v, want %v", event.Status, tt.expectedStatus)
			}
		})
	}
}

func TestFlipProvider_CreateVA(t *testing.T) {
	// [FIX] Gunakan tanggal tetap yang jauh di masa depan untuk mock response.
	// Ini memastikan jika parsing expired_date GAGAL dan fallback ke time.Now()+24h,
	// assertion PASTI mendeteksi perbedaannya (tahun 2099 vs tahun sekarang).
	// Sebelumnya menggunakan time.Now() yang bisa jatuh di detik yang sama
	// dengan fallback, sehingga test selalu hijau meskipun parsing gagal.
	fixedFutureDate := "2099-12-31 23:59:59"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify Content-Type: form-urlencoded, bukan JSON
		if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
			t.Errorf("expected Content-Type application/x-www-form-urlencoded, got %s", r.Header.Get("Content-Type"))
		}

		// Verify Auth: Basic, bukan Bearer
		if !strings.HasPrefix(r.Header.Get("Authorization"), "Basic ") {
			t.Errorf("expected Basic Auth, got %s", r.Header.Get("Authorization"))
		}

		// Verify Idempotency-Key header
		if r.Header.Get("Idempotency-Key") != "uuid-v7-test" {
			t.Errorf("expected Idempotency-Key uuid-v7-test, got %s", r.Header.Get("Idempotency-Key"))
		}

		// Parse form data
		if err := r.ParseForm(); err != nil {
			t.Errorf("failed to parse form: %v", err)
		}
		if r.FormValue("title") != "Hyperlocal Top-Up" {
			t.Errorf("expected title 'Hyperlocal Top-Up', got %s", r.FormValue("title"))
		}
		if r.FormValue("bank_code") != "BCA" {
			t.Errorf("expected bank_code 'BCA', got %s", r.FormValue("bank_code"))
		}
		if r.FormValue("amount") != "50000" {
			t.Errorf("expected amount '50000', got %s", r.FormValue("amount"))
		}

		// Mock Response — Flip membalas dengan JSON
		resp := map[string]interface{}{
			"id":             "flip_bill_123",
			"account_number": "888899990000",
			"bank_code":      "BCA",
			"bank_name":      "Bank Central Asia",
			"expired_date":   fixedFutureDate, // Tanggal tetap untuk mendeteksi fallback
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	provider := NewFlipProvider(server.URL, "test-api-key", "test-secret-key")

	req := &CreateVARequest{
		ExternalID:     "topup-driver1-123456",
		IdempotencyKey: "uuid-v7-test",
		Name:           "Hyperlocal Top-Up",
		Amount:         50000,
		ExpiryHours:    24,
		BankCode:       "BCA",
	}

	resp, err := provider.CreateVA(context.Background(), req)
	if err != nil {
		t.Fatalf("CreateVA() error = %v", err)
	}

	if resp.BillID != "flip_bill_123" {
		t.Errorf("expected BillID flip_bill_123, got %s", resp.BillID)
	}
	if resp.AccountNumber != "888899990000" {
		t.Errorf("expected AccountNumber 888899990000, got %s", resp.AccountNumber)
	}
	if resp.BankName != "Bank Central Asia" {
		t.Errorf("expected BankName Bank Central Asia, got %s", resp.BankName)
	}

	// [CRITICAL ASSERTION] Verifikasi Parsing Timezone
	// Karena mock mengirim "2099-12-31 23:59:59", jika parsing GAGAL dan fallback
	// ke time.Now()+24h, nilai resp.ExpiresAt akan menjadi tahun sekarang,
	// dan assertion ini PASTI menangkap bug tersebut.
	if resp.ExpiresAt.Format("2006-01-02 15:04:05") != fixedFutureDate {
		t.Errorf("expected ExpiresAt parsed value %s, got %s (kemungkinan parsing gagal dan fallback ke waktu lokal)",
			fixedFutureDate, resp.ExpiresAt.Format("2006-01-02 15:04:05"))
	}
	if resp.ExpiresAt.Location().String() != "Asia/Jakarta" {
		t.Errorf("expected ExpiresAt timezone Asia/Jakarta, got %s", resp.ExpiresAt.Location())
	}
}
