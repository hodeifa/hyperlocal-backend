package vatopup

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// --- Mock Transports for 100% Branch Coverage ---

type errorTransport struct {
	err error
}

func (t *errorTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return nil, t.err
}

type failReader struct{ err error }

func (f *failReader) Read(p []byte) (int, error) { return 0, f.err }

type failingBodyTransport struct{}

func (t *failingBodyTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	return &http.Response{
		StatusCode: 200,
		Body:       io.NopCloser(&failReader{err: errors.New("simulated_read_error")}),
	}, nil
}

func TestFlipProvider_ValidateSignature(t *testing.T) {
	secret := "test-secret-key"
	provider := NewFlipProvider(FlipProviderConfig{SecretKey: secret})

	payload := []byte(`{"id":"evt_123","bill_id":"bill_456","amount":47500,"fee":2500,"status":"SUCCESSFUL"}`)

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write(payload)
	validSignature := hex.EncodeToString(mac.Sum(nil))

	tests := []struct {
		signature string
		name      string
		payload   []byte
		want      bool
	}{
		{validSignature, "valid signature", payload, true},
		{"invalid-signature", "invalid signature", payload, false},
		{validSignature, "tampered payload", []byte(`{"id":"evt_123","bill_id":"bill_456","amount":47500,"fee":2500,"status":"FAILED"}`), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := provider.ValidateSignature(tt.payload, tt.signature); got != tt.want {
				t.Errorf("ValidateSignature() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestFlipProvider_ParseCallback(t *testing.T) {
	provider := NewFlipProvider(FlipProviderConfig{})

	tests := []struct {
		check   func(*CallbackEvent) bool
		name    string
		payload []byte
		wantErr bool
	}{
		{
			name:    "successful callback",
			payload: []byte(`{"id":"evt_123","bill_id":"bill_456","amount":47500,"fee":2500,"status":"SUCCESSFUL"}`),
			wantErr: false,
			check: func(e *CallbackEvent) bool {
				return e.EventID == "evt_123" && e.BillID == "bill_456" &&
					e.AmountReceived == 47500 && e.Fee == 2500 && e.Status == "SUCCESSFUL"
			},
		},
		{
			name:    "invalid json",
			payload: []byte(`{invalid json`),
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := provider.ParseCallback(tt.payload)
			if (err != nil) != tt.wantErr {
				t.Errorf("ParseCallback() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr && tt.check != nil && !tt.check(got) {
				t.Errorf("ParseCallback() got = %+v, check failed", got)
			}
		})
	}
}

func TestFlipProvider_CreateVA(t *testing.T) {
	wib := time.FixedZone("WIB", 7*60*60)
	expectedExpiryStr := time.Now().In(wib).Add(24 * time.Hour).Format("2006-01-02 15:04:05")
	expectedExpiry, _ := time.ParseInLocation("2006-01-02 15:04:05", expectedExpiryStr, wib)

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		_, _ = w.Write([]byte(`{"id":"bill_123","va_number":"1234567890","bank_name":"BCA","expired_date":"` + expectedExpiryStr + `"}`))
	}))
	defer server.Close()

	provider := NewFlipProvider(FlipProviderConfig{
		BaseURL: server.URL,
		APIKey:  "test-api-key",
	})

	req := &CreateVARequest{
		Name:        "Hyperlocal Top-Up",
		Amount:      50000,
		BankCode:    "bca",
		ExpiryHours: 24,
		ExternalID:  "topup-123-456",
	}

	resp, err := provider.CreateVA(context.Background(), req)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if resp.BillID != "bill_123" {
		t.Errorf("expected BillID bill_123, got %s", resp.BillID)
	}
	if resp.AccountNumber != "1234567890" {
		t.Errorf("expected AccountNumber 1234567890, got %s", resp.AccountNumber)
	}
	if resp.BankName != "BCA" {
		t.Errorf("expected BankName BCA, got %s", resp.BankName)
	}
	if !resp.ExpiresAt.Equal(expectedExpiry) {
		t.Errorf("expected ExpiresAt %v, got %v", expectedExpiry, resp.ExpiresAt)
	}
}

func TestFlipProvider_CreateVA_Errors(t *testing.T) {
	tests := []struct {
		req        *CreateVARequest
		ctx        context.Context
		handler    http.HandlerFunc
		setup      func(*FlipProvider)
		name       string
		wantErrMsg string
	}{
		{
			name:       "nil request",
			req:        nil,
			ctx:        context.Background(),
			handler:    func(w http.ResponseWriter, r *http.Request) {},
			wantErrMsg: "create VA request cannot be nil",
		},
		{
			name:       "nil context triggers NewRequestWithContext error",
			req:        &CreateVARequest{Name: "Test", Amount: 50000, BankCode: "bca", ExpiryHours: 24},
			ctx:        nil,
			handler:    func(w http.ResponseWriter, r *http.Request) {},
			wantErrMsg: "failed to create http request",
		},
		{
			name: "client.Do network failure",
			req:  &CreateVARequest{Name: "Test", Amount: 50000, BankCode: "bca", ExpiryHours: 24},
			ctx:  context.Background(),
			setup: func(p *FlipProvider) {
				p.client.Transport = &errorTransport{err: errors.New("simulated_network_error")}
			},
			wantErrMsg: "failed to execute http request",
		},
		{
			name: "io.ReadAll failure",
			req:  &CreateVARequest{Name: "Test", Amount: 50000, BankCode: "bca", ExpiryHours: 24},
			ctx:  context.Background(),
			setup: func(p *FlipProvider) {
				p.client.Transport = &failingBodyTransport{}
			},
			wantErrMsg: "failed to read response body",
		},
		{
			name: "non-2xx status code",
			req:  &CreateVARequest{Name: "Test", Amount: 50000, BankCode: "bca", ExpiryHours: 24},
			ctx:  context.Background(),
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusBadRequest)
				_, _ = w.Write([]byte(`{"error": "invalid amount"}`))
			},
			wantErrMsg: "flip api returned non-2xx status code: 400",
		},
		{
			name: "invalid json response",
			req:  &CreateVARequest{Name: "Test", Amount: 50000, BankCode: "bca", ExpiryHours: 24},
			ctx:  context.Background(),
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(`{invalid json`))
			},
			wantErrMsg: "failed to unmarshal flip response",
		},
		{
			name: "invalid expired_date format in response",
			req:  &CreateVARequest{Name: "Test", Amount: 50000, BankCode: "bca", ExpiryHours: 24},
			ctx:  context.Background(),
			handler: func(w http.ResponseWriter, r *http.Request) {
				w.WriteHeader(http.StatusCreated)
				_, _ = w.Write([]byte(`{"id":"bill_123","va_number":"123","bank_name":"BCA","expired_date":"invalid-date"}`))
			},
			wantErrMsg: "failed to parse expired_date from flip response",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(tt.handler)
			defer server.Close()

			provider := NewFlipProvider(FlipProviderConfig{
				BaseURL: server.URL,
				APIKey:  "test",
			})

			if tt.setup != nil {
				tt.setup(provider)
			}

			_, err := provider.CreateVA(tt.ctx, tt.req)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if !strings.Contains(err.Error(), tt.wantErrMsg) {
				t.Errorf("expected error containing %q, got %q", tt.wantErrMsg, err.Error())
			}
		})
	}
}
