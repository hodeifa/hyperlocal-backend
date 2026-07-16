package grpcclient

import (
	"errors"
	"testing"

	"github.com/sony/gobreaker"
)

func TestCircuitBreaker_TripsAfterConsecutiveFailures(t *testing.T) {
	cb := NewCircuitBreaker("test-service")
	dummyErr := errors.New("downstream error")

	for i := 0; i < 5; i++ {
		_, _ = cb.Execute(func() (interface{}, error) {
			return nil, dummyErr
		})
	}

	_, err := cb.Execute(func() (interface{}, error) {
		return "success", nil
	})

	if !IsCircuitOpen(err) {
		t.Errorf("expected circuit to be open, got %v", err)
	}
}

func TestIsCircuitOpen(t *testing.T) {
	tests := []struct {
		err      error
		name     string
		expected bool
	}{
		{err: gobreaker.ErrOpenState, name: "ErrOpenState", expected: true},
		{err: gobreaker.ErrTooManyRequests, name: "ErrTooManyRequests", expected: true},
		{err: errors.New("rpc error: code = Unavailable"), name: "GenericError", expected: false},
		{err: nil, name: "NilError", expected: false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := IsCircuitOpen(tt.err); got != tt.expected {
				t.Errorf("IsCircuitOpen(%v) = %v, want %v", tt.err, got, tt.expected)
			}
		})
	}
}
