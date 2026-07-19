// Package vatopup implements the Provider interface for Flip Business, a virtual account provider in Indonesia.
package vatopup

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// FlipProvider implements the Provider interface for Flip Business.
type FlipProvider struct {
	client    *http.Client
	baseURL   string
	apiKey    string
	secretKey string
}

// FlipProviderConfig holds configuration for FlipProvider.
type FlipProviderConfig struct {
	BaseURL   string
	APIKey    string
	SecretKey string
}

// NewFlipProvider creates a new instance of FlipProvider.
func NewFlipProvider(cfg FlipProviderConfig) *FlipProvider {
	return &FlipProvider{
		client: &http.Client{
			Timeout: 5 * time.Second,
		},
		baseURL:   cfg.BaseURL,
		apiKey:    cfg.APIKey,
		secretKey: cfg.SecretKey,
	}
}

// ValidateSignature validates the HMAC-SHA256 signature of a webhook payload.
func (f *FlipProvider) ValidateSignature(payload []byte, signature string) bool {
	mac := hmac.New(sha256.New, []byte(f.secretKey))
	mac.Write(payload)
	expectedMAC := mac.Sum(nil)
	expectedSignature := hex.EncodeToString(expectedMAC)
	return hmac.Equal([]byte(expectedSignature), []byte(signature))
}

type flipCallbackPayload struct {
	ID     string `json:"id"`
	BillID string `json:"bill_id"`
	Status string `json:"status"`
	Amount int64  `json:"amount"` // flip.md §4.4 code snippet uses "amount"
	Fee    int64  `json:"fee"`
}

// ParseCallback parses the webhook payload from Flip.
func (f *FlipProvider) ParseCallback(payload []byte) (*CallbackEvent, error) {
	var flipCb flipCallbackPayload
	if err := json.Unmarshal(payload, &flipCb); err != nil {
		return nil, fmt.Errorf("failed to unmarshal flip callback: %w", err)
	}

	return &CallbackEvent{
		EventID:        flipCb.ID,
		BillID:         flipCb.BillID,
		Status:         flipCb.Status,
		AmountReceived: flipCb.Amount,
		Fee:            flipCb.Fee,
	}, nil
}

type flipCreateVAPayload struct {
	Title       string `json:"title"`
	ExpiredDate string `json:"expired_date"`
	BankCode    string `json:"bank_code"`
	Amount      int64  `json:"amount"`
}

type flipCreateVAResponse struct {
	ID          string `json:"id"`
	VANumber    string `json:"va_number"`
	BankName    string `json:"bank_name"`
	ExpiredDate string `json:"expired_date"`
}

// CreateVA creates a new Virtual Account via Flip Business API.
func (f *FlipProvider) CreateVA(ctx context.Context, req *CreateVARequest) (*CreateVAResponse, error) {
	if req == nil {
		return nil, fmt.Errorf("create VA request cannot be nil")
	}

	// Use FixedZone to avoid panic if tzdata is missing in minimal Docker images.
	wib := time.FixedZone("WIB", 7*60*60)
	localExpiredAt := time.Now().In(wib).Add(time.Duration(req.ExpiryHours) * time.Hour)

	payload := flipCreateVAPayload{
		Title:       req.Name,
		ExpiredDate: localExpiredAt.Format("2006-01-02 15:04:05"),
		BankCode:    req.BankCode,
		Amount:      req.Amount,
	}

	bodyBytes, err := json.Marshal(payload)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal create va payload: %w", err)
	}

	httpReq, err := http.NewRequestWithContext(ctx, http.MethodPost, f.baseURL+"/v2/disbursement/bill", bytes.NewReader(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("failed to create http request: %w", err)
	}

	httpReq.Header.Set("Content-Type", "application/json")
	httpReq.SetBasicAuth(f.apiKey, "")

	resp, err := f.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("failed to execute http request: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read response body: %w", err)
	}

	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, fmt.Errorf("flip api returned non-2xx status code: %d, body: %s", resp.StatusCode, string(respBody))
	}

	var flipResp flipCreateVAResponse
	if err = json.Unmarshal(respBody, &flipResp); err != nil {
		return nil, fmt.Errorf("failed to unmarshal flip response: %w", err)
	}

	expiresAt, err := time.ParseInLocation("2006-01-02 15:04:05", flipResp.ExpiredDate, wib)
	if err != nil {
		return nil, fmt.Errorf("failed to parse expired_date from flip response: %w", err)
	}

	return &CreateVAResponse{
		ExpiresAt:     expiresAt,
		BillID:        flipResp.ID,
		AccountNumber: flipResp.VANumber,
		BankName:      flipResp.BankName,
	}, nil
}
