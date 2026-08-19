// Package vatopup provides for Flip Business Virtual Accounts.
package vatopup

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// Compile-time check to ensure FlipProvider implements Provider
var _ Provider = (*FlipProvider)(nil)

// FlipProvider implements the Provider interface for Flip Business Virtual Accounts.
type FlipProvider struct {
	client    *http.Client
	baseURL   string
	apiKey    string
	secretKey string // Untuk HMAC-SHA256 signature webhook & Basic Auth
}

// NewFlipProvider menginisialisasi client Flip dengan timeout ketat 5 detik.
func NewFlipProvider(baseURL, apiKey, secretKey string) *FlipProvider {
	return &FlipProvider{
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		baseURL:   baseURL,
		apiKey:    apiKey,
		secretKey: secretKey,
	}
}

// ValidateSignature memvalidasi HMAC-SHA256 dari payload webhook.
func (f *FlipProvider) ValidateSignature(payload []byte, signature string) bool {
	mac := hmac.New(sha256.New, []byte(f.secretKey))
	mac.Write(payload)
	expectedMAC := mac.Sum(nil)
	expectedSignature := hex.EncodeToString(expectedMAC)

	// hmac.Equal mencegah timing attack
	return hmac.Equal([]byte(expectedSignature), []byte(signature))
}

// ParseCallback mem-parse payload webhook dan meneruskan status MENTAH dari Flip.
func (f *FlipProvider) ParseCallback(payload []byte) (*CallbackEvent, error) {
	var flipCallback struct {
		ID     string `json:"id"`
		BillID string `json:"bill_id"`
		Status string `json:"status"`
		Amount int64  `json:"amount"`
		Fee    int64  `json:"fee"`
	}

	if err := json.Unmarshal(payload, &flipCallback); err != nil {
		return nil, fmt.Errorf("failed to unmarshal flip callback: %w", err)
	}

	return &CallbackEvent{
		EventID:        flipCallback.ID,
		BillID:         flipCallback.BillID,
		AmountReceived: flipCallback.Amount,
		Fee:            flipCallback.Fee,
		Status:         flipCallback.Status, // MENTAH — mapping ke PAID/EXPIRED dilakukan di Usecase
	}, nil
}

// CreateVA memanggil API Flip untuk membuat Virtual Account.
func (f *FlipProvider) CreateVA(ctx context.Context, req *CreateVARequest) (*CreateVAResponse, error) {
	wib, err := time.LoadLocation("Asia/Jakarta")
	if err != nil {
		return nil, fmt.Errorf("failed to load WIB timezone: %w", err)
	}

	// Hitung expired_date untuk dikirim ke Flip (WIB) sebagai fallback
	expiredAtReq := time.Now().In(wib).Add(time.Duration(req.ExpiryHours) * time.Hour)

	// Gunakan url.Values (form-urlencoded), BUKAN JSON payload
	formData := url.Values{}
	formData.Set("title", req.Name)
	formData.Set("amount", strconv.FormatInt(req.Amount, 10))
	formData.Set("expired_date", expiredAtReq.Format("2006-01-02 15:04:05"))
	formData.Set("bank_code", req.BankCode)

	endpoint := f.baseURL + "/v2/pwf/bill"

	// strings.NewReader digunakan untuk form-urlencoded, sehingga import "bytes" tidak lagi dibutuhkan
	httpReq, err := http.NewRequestWithContext(ctx, "POST", endpoint, strings.NewReader(formData.Encode()))
	if err != nil {
		return nil, fmt.Errorf("failed to create HTTP request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/x-www-form-urlencoded")

	// Basic Auth: Base64(SecretKey + ":")
	authString := base64.StdEncoding.EncodeToString([]byte(f.secretKey + ":"))
	httpReq.Header.Set("Authorization", "Basic "+authString)

	if req.IdempotencyKey != "" {
		httpReq.Header.Set("Idempotency-Key", req.IdempotencyKey)
	}

	resp, err := f.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("HTTP request to Flip failed: %w", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode >= 400 {
		return nil, fmt.Errorf("Flip API returned status %d: %s", resp.StatusCode, string(body))
	}

	var flipResp struct {
		ID            string `json:"id"`
		AccountNumber string `json:"account_number"`
		BankCode      string `json:"bank_code"`
		BankName      string `json:"bank_name"`
		ExpiredDate   string `json:"expired_date"`
	}

	if err = json.Unmarshal(body, &flipResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal Flip response: %w", err)
	}

	// Parse expired_date yang dikembalikan Flip dengan timezone WIB
	parsedExpiresAt, err := time.ParseInLocation("2006-01-02 15:04:05", flipResp.ExpiredDate, wib)
	if err != nil {
		// Fallback jika Flip tidak mengembalikan format yang diharapkan atau error parsing
		parsedExpiresAt = expiredAtReq
	}

	bankName := flipResp.BankName
	if bankName == "" {
		bankName = flipResp.BankCode
	}

	return &CreateVAResponse{
		BillID:        flipResp.ID,
		AccountNumber: flipResp.AccountNumber,
		BankName:      bankName,
		ExpiresAt:     parsedExpiresAt,
	}, nil
}
