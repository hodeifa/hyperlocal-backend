package fcm

import (
	"context"
	"errors"
	"testing"

	"firebase.google.com/go/v4/messaging"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockSender implements fcmSender.
type MockSender struct {
	mock.Mock
}

func (m *MockSender) Send(ctx context.Context, msg *messaging.Message) (string, error) {
	args := m.Called(ctx, msg)
	return args.String(0), args.Error(1)
}

// MockFetcher implements TokenFetcher.
type MockFetcher struct {
	mock.Mock
}

func (m *MockFetcher) GetCustomerFCMToken(ctx context.Context, customerID string) (string, error) {
	args := m.Called(ctx, customerID)
	return args.String(0), args.Error(1)
}

func (m *MockFetcher) GetDriverFCMToken(ctx context.Context, driverID string) (string, error) {
	args := m.Called(ctx, driverID)
	return args.String(0), args.Error(1)
}

func TestSend_Success(t *testing.T) {
	mockSender := new(MockSender)
	mockSender.On("Send", mock.Anything, mock.Anything).Return("msg-id-123", nil)

	logger, _ := zap.NewDevelopment()
	client := &Client{sender: mockSender, logger: logger}

	err := client.Send(context.Background(), "dummy-token", "Title", "Body", nil)
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	mockSender.AssertExpectations(t)
}

func TestSend_EmptyToken(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	client := &Client{sender: nil, logger: logger}

	err := client.Send(context.Background(), "", "Title", "Body", nil)
	if err == nil {
		t.Error("Expected error for empty token")
	}
}

func TestSend_SenderError(t *testing.T) {
	mockSender := new(MockSender)
	mockSender.On("Send", mock.Anything, mock.Anything).Return("", errors.New("firebase down"))

	logger, _ := zap.NewDevelopment()
	client := &Client{sender: mockSender, logger: logger}

	err := client.Send(context.Background(), "token", "Title", "Body", nil)
	if err == nil {
		t.Error("Expected error from sender")
	}
}

func TestSendToDriver_Success(t *testing.T) {
	mockSender := new(MockSender)
	mockFetcher := new(MockFetcher)

	mockFetcher.On("GetDriverFCMToken", mock.Anything, "driver-123").Return("driver-token-abc", nil)
	mockSender.On("Send", mock.Anything, mock.MatchedBy(func(msg *messaging.Message) bool {
		return msg.Token == "driver-token-abc" && msg.Data["deep_link"] == "hmitra://order/1"
	})).Return("msg-id-123", nil)

	logger, _ := zap.NewDevelopment()
	client := &Client{sender: mockSender, fetcher: mockFetcher, logger: logger}

	err := client.SendToDriver(context.Background(), "driver-123", "Order Masuk", "Ada pesanan baru", "hmitra://order/1")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	mockFetcher.AssertExpectations(t)
	mockSender.AssertExpectations(t)
}

func TestSendToDriver_FetcherError(t *testing.T) {
	mockSender := new(MockSender)
	mockFetcher := new(MockFetcher)

	mockFetcher.On("GetDriverFCMToken", mock.Anything, "driver-123").Return("", errors.New("db down"))

	logger, _ := zap.NewDevelopment()
	client := &Client{sender: mockSender, fetcher: mockFetcher, logger: logger}

	err := client.SendToDriver(context.Background(), "driver-123", "Title", "Body", "")
	if err == nil {
		t.Error("Expected error from fetcher")
	}
}

func TestSendToDriver_EmptyToken(t *testing.T) {
	mockSender := new(MockSender)
	mockFetcher := new(MockFetcher)

	mockFetcher.On("GetDriverFCMToken", mock.Anything, "driver-123").Return("", nil)

	logger, _ := zap.NewDevelopment()
	client := &Client{sender: mockSender, fetcher: mockFetcher, logger: logger}

	err := client.SendToDriver(context.Background(), "driver-123", "Title", "Body", "")
	if err != nil {
		t.Errorf("Expected silent skip (no error), got %v", err)
	}
	mockSender.AssertNotCalled(t, "Send", mock.Anything, mock.Anything)
}

func TestSendToDriver_NilFetcher(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	client := &Client{sender: nil, fetcher: nil, logger: logger}

	err := client.SendToDriver(context.Background(), "driver-123", "Title", "Body", "")
	if err == nil {
		t.Error("Expected error for nil fetcher")
	}
}

func TestSendToCustomer_Success(t *testing.T) {
	mockSender := new(MockSender)
	mockFetcher := new(MockFetcher)

	mockFetcher.On("GetCustomerFCMToken", mock.Anything, "cust-123").Return("cust-token-xyz", nil)
	mockSender.On("Send", mock.Anything, mock.MatchedBy(func(msg *messaging.Message) bool {
		return msg.Token == "cust-token-xyz" && msg.Data["deep_link"] == "hcust://order/1"
	})).Return("msg-id-456", nil)

	logger, _ := zap.NewDevelopment()
	client := &Client{sender: mockSender, fetcher: mockFetcher, logger: logger}

	err := client.SendToCustomer(context.Background(), "cust-123", "Mitra Ditemukan", "Mitra sedang menuju lokasi", "hcust://order/1")
	if err != nil {
		t.Errorf("Expected no error, got %v", err)
	}
	mockFetcher.AssertExpectations(t)
	mockSender.AssertExpectations(t)
}

func TestSendToCustomer_FetcherError(t *testing.T) {
	mockSender := new(MockSender)
	mockFetcher := new(MockFetcher)

	mockFetcher.On("GetCustomerFCMToken", mock.Anything, "cust-123").Return("", errors.New("db down"))

	logger, _ := zap.NewDevelopment()
	client := &Client{sender: mockSender, fetcher: mockFetcher, logger: logger}

	err := client.SendToCustomer(context.Background(), "cust-123", "Title", "Body", "")
	if err == nil {
		t.Error("Expected error from fetcher")
	}
}

func TestSendToCustomer_EmptyToken(t *testing.T) {
	mockSender := new(MockSender)
	mockFetcher := new(MockFetcher)

	mockFetcher.On("GetCustomerFCMToken", mock.Anything, "cust-123").Return("", nil)

	logger, _ := zap.NewDevelopment()
	client := &Client{sender: mockSender, fetcher: mockFetcher, logger: logger}

	err := client.SendToCustomer(context.Background(), "cust-123", "Title", "Body", "")
	if err != nil {
		t.Errorf("Expected silent skip (no error), got %v", err)
	}
	mockSender.AssertNotCalled(t, "Send", mock.Anything, mock.Anything)
}

func TestSendToCustomer_NilFetcher(t *testing.T) {
	logger, _ := zap.NewDevelopment()
	client := &Client{sender: nil, fetcher: nil, logger: logger}

	err := client.SendToCustomer(context.Background(), "cust-123", "Title", "Body", "")
	if err == nil {
		t.Error("Expected error for nil fetcher")
	}
}