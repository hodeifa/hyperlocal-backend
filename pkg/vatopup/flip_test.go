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

	// Urutan field sengaja: string dulu (16 byte), baru bool, agar linter
	// govet/fieldalignment (aktif di .golangci.yml, TIDAK dikecualikan untuk _test.go) tidak komplain padding.
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

	// Selain Status, kita juga assert EventID/BillID/AmountReceived/Fee supaya
	// regresi mapping field (bukan cuma status) ikut tertangkap oleh test ini.
	//
	// vatopup.go mewajibkan Status TETAP MENTAH ("SUCCESSFUL"/"FAILED"/"PENDING"/"EXPIRED") —
	// mapping ke status internal (PAID/EXPIRED) adalah tanggung jawab Usecase, BUKAN provider ini,
	// supaya guard `if event.Status != "SUCCESSFUL"` di webhook handler tetap benar.
	tests := []struct {
		name                   string
		payload                string
		expectedStatus         string
		expectedEventID        string
		expectedBillID         string
		expectedAmountReceived int64
		expectedFee            int64
		expectErr              bool
	}{
		{
			name:                   "Success Status",
			payload:                `{"id":"evt_1","bill_id":"bill_1","amount":47500,"fee":2500,"status":"SUCCESSFUL"}`,
			expectedStatus:         "SUCCESSFUL", // RAW, bukan "PAID"
			expectedEventID:        "evt_1",
			expectedBillID:         "bill_1",
			expectedAmountReceived: 47500,
			expectedFee:            2500,
		},
		{
			name:            "Failed Status",
			payload:         `{"id":"evt_2","bill_id":"bill_2","amount":0,"fee":0,"status":"FAILED"}`,
			expectedStatus:  "FAILED", // RAW
			expectedEventID: "evt_2",
			expectedBillID:  "bill_2",
		},
		{
			name:            "Pending Status",
			payload:         `{"id":"evt_3","bill_id":"bill_3","amount":0,"fee":0,"status":"PENDING"}`,
			expectedStatus:  "PENDING", // RAW
			expectedEventID: "evt_3",
			expectedBillID:  "bill_3",
		},
		{
			name:            "Expired Status",
			payload:         `{"id":"evt_4","bill_id":"bill_4","amount":0,"fee":0,"status":"EXPIRED"}`,
			expectedStatus:  "EXPIRED", // RAW
			expectedEventID: "evt_4",
			expectedBillID:  "bill_4",
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
			if tt.expectErr {
				return
			}
			if event.Status != tt.expectedStatus {
				t.Errorf("ParseCallback() status = %v, want %v", event.Status, tt.expectedStatus)
			}
			if event.EventID != tt.expectedEventID {
				t.Errorf("ParseCallback() EventID = %v, want %v", event.EventID, tt.expectedEventID)
			}
			if event.BillID != tt.expectedBillID {
				t.Errorf("ParseCallback() BillID = %v, want %v", event.BillID, tt.expectedBillID)
			}
			if event.AmountReceived != tt.expectedAmountReceived {
				t.Errorf("ParseCallback() AmountReceived = %v, want %v", event.AmountReceived, tt.expectedAmountReceived)
			}
			if event.Fee != tt.expectedFee {
				t.Errorf("ParseCallback() Fee = %v, want %v", event.Fee, tt.expectedFee)
			}
		})
	}
}

func TestFlipProvider_CreateVA_Success(t *testing.T) {
	// Pakai tanggal tetap yang jauh di masa depan untuk mock response.
	// Ini memastikan jika parsing expired_date GAGAL dan diam-diam fallback ke
	// time.Now()+ExpiryHours, assertion di bawah PASTI mendeteksi selisihnya
	// (tahun 2099 vs tahun berjalan) — bukan cuma kebetulan lolos karena timing.
	fixedFutureDate := "2099-12-31 23:59:59"

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("Content-Type") != "application/x-www-form-urlencoded" {
			t.Errorf("expected Content-Type application/x-www-form-urlencoded, got %s", r.Header.Get("Content-Type"))
		}
		if !strings.HasPrefix(r.Header.Get("Authorization"), "Basic ") {
			t.Errorf("expected Basic Auth, got %s", r.Header.Get("Authorization"))
		}
		if r.Header.Get("Idempotency-Key") != "uuid-v7-test" {
			t.Errorf("expected Idempotency-Key uuid-v7-test, got %s", r.Header.Get("Idempotency-Key"))
		}

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

		resp := map[string]interface{}{
			"id":             "flip_bill_123",
			"account_number": "888899990000",
			"bank_code":      "BCA",
			"bank_name":      "Bank Central Asia",
			"expired_date":   fixedFutureDate,
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(resp) //nolint:errcheck // helper test server
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

	// Verifikasi parsing timezone WIB — lihat komentar fixedFutureDate di atas.
	if resp.ExpiresAt.Format("2006-01-02 15:04:05") != fixedFutureDate {
		t.Errorf("expected ExpiresAt parsed value %s, got %s (kemungkinan parsing gagal dan diam-diam fallback ke waktu lokal)",
			fixedFutureDate, resp.ExpiresAt.Format("2006-01-02 15:04:05"))
	}
	if resp.ExpiresAt.Location().String() != "Asia/Jakarta" {
		t.Errorf("expected ExpiresAt timezone Asia/Jakarta, got %s", resp.ExpiresAt.Location())
	}
}

func TestFlipProvider_CreateVA_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(`{"error": "internal server error"}`)) //nolint:errcheck
	}))
	defer server.Close()

	provider := NewFlipProvider(server.URL, "test-api-key", "test-secret-key")

	req := &CreateVARequest{Name: "Test", Amount: 50000, ExpiryHours: 24, BankCode: "BCA"}
	_, err := provider.CreateVA(context.Background(), req)

	if err == nil {
		t.Fatal("expected error for HTTP 500, got nil")
	}
	if !strings.Contains(err.Error(), "Flip API returned status 500") {
		t.Errorf("expected error to contain 'Flip API returned status 500', got %v", err)
	}
}

func TestFlipProvider_CreateVA_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		w.Write([]byte(`{"invalid_json`)) //nolint:errcheck
	}))
	defer server.Close()

	provider := NewFlipProvider(server.URL, "test-api-key", "test-secret-key")

	req := &CreateVARequest{Name: "Test", Amount: 50000, ExpiryHours: 24, BankCode: "BCA"}
	_, err := provider.CreateVA(context.Background(), req)

	if err == nil {
		t.Fatal("expected error for invalid JSON, got nil")
	}
	if !strings.Contains(err.Error(), "failed to unmarshal Flip response") {
		t.Errorf("expected unmarshal error, got %v", err)
	}
}

func TestFlipProvider_CreateVA_NetworkError(t *testing.T) {
	// Port 1 biasanya tidak dipakai dan langsung memicu connection refused.
	provider := NewFlipProvider("http://127.0.0.1:1", "test-api-key", "test-secret-key")

	req := &CreateVARequest{Name: "Test", Amount: 50000, ExpiryHours: 24, BankCode: "BCA"}
	_, err := provider.CreateVA(context.Background(), req)

	if err == nil {
		t.Fatal("expected network error, got nil")
	}
	if !strings.Contains(err.Error(), "HTTP request to Flip failed") {
		t.Errorf("expected network error, got %v", err)
	}
}