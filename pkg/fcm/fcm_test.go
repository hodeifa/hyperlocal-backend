package fcm

import (
	"context"
	"errors"
	"testing"

	"firebase.google.com/go/v4/messaging"
	"github.com/stretchr/testify/mock"
	"go.uber.org/zap"
)

// MockSender implements fcmSender
type MockSender struct {
	mock.Mock
}

func (m *MockSender) Send(ctx context.Context, msg *messaging.Message) (string, error) {
	args := m.Called(ctx, msg)
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
	client := &Client{sender: nil, logger: logger} // Sender tidak dipanggil jika token kosong

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
