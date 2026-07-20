package fcm

import (
	"context"
	"errors"
	"testing"

	"firebase.google.com/go/v4/messaging"
	"go.uber.org/zap"
)

// MockSender implements fcmSender for testing.
type MockSender struct {
	SendFunc func(ctx context.Context, message *messaging.Message) (string, error)
}

// Send executes the mock SendFunc.
func (m *MockSender) Send(ctx context.Context, message *messaging.Message) (string, error) {
	return m.SendFunc(ctx, message)
}

//nolint:govet // fieldalignment: field order in test mocks is not critical to optimize.
type mockFetcher struct {
	token string
	err   error
}

func (m *mockFetcher) GetCustomerFCMToken(ctx context.Context, customerID string) (string, error) {
	return m.token, m.err
}

func (m *mockFetcher) GetDriverFCMToken(ctx context.Context, driverID string) (string, error) {
	return m.token, m.err
}

func TestSend_Success(t *testing.T) {
	mockSender := &MockSender{
		SendFunc: func(ctx context.Context, message *messaging.Message) (string, error) {
			//nolint:staticcheck // message.Token is the FCM registration token, not Fid (Installation ID).
			if message.Token != "valid-token" {
				t.Errorf("expected token 'valid-token', got %s", message.Token)
			}
			return "projects/test/messages/123", nil
		},
	}

	client := &Client{sender: mockSender, logger: zap.NewNop()}
	err := client.Send(context.Background(), "valid-token", "Test", "Body", nil)
	if err != nil {
		t.Errorf("expected no error, got %v", err)
	}
}

func TestSend_SenderError(t *testing.T) {
	mockSender := &MockSender{
		SendFunc: func(ctx context.Context, message *messaging.Message) (string, error) {
			return "", errors.New("firebase network error")
		},
	}

	client := &Client{sender: mockSender, logger: zap.NewNop()}
	err := client.Send(context.Background(), "valid-token", "Test", "Body", nil)
	if err == nil {
		t.Errorf("expected error from sender, got nil")
	}
}

func TestSendToDriver_FetcherError(t *testing.T) {
	fetcher := &mockFetcher{err: errors.New("db error")}
	client := &Client{fetcher: fetcher, logger: zap.NewNop()}

	err := client.SendToDriver(context.Background(), "driver-123", "Title", "Body", "hmitra://home")
	if err == nil {
		t.Errorf("expected error from fetcher, got nil")
	}
}

func TestSendToDriver_EmptyToken(t *testing.T) {
	fetcher := &mockFetcher{token: ""}
	client := &Client{fetcher: fetcher, logger: zap.NewNop()}

	err := client.SendToDriver(context.Background(), "driver-123", "Title", "Body", "hmitra://home")
	if err != nil {
		t.Errorf("expected nil error for empty token, got %v", err)
	}
}

func TestNewClient_MissingCredentials(t *testing.T) {
	cfg := Config{}
	// It is safe to pass a nil logger here because NewClient will immediately
	// return an error on credential validation (the else branch) before
	// any line of code that touches the logger is executed.
	_, err := NewClient(context.Background(), cfg, nil, nil)
	if err == nil {
		t.Errorf("expected error for missing credentials, got nil")
	}
}
