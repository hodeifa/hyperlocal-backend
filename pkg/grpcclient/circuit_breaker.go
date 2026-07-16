package grpcclient

import (
	"log"
	"time"

	"github.com/sony/gobreaker"
)

// NewCircuitBreaker membuat instance circuit breaker baru untuk sebuah service target.
func NewCircuitBreaker(name string) *gobreaker.CircuitBreaker {
	return gobreaker.NewCircuitBreaker(gobreaker.Settings{
		Name:        name,
		MaxRequests: 5,
		Interval:    0,
		Timeout:     30 * time.Second,
		ReadyToTrip: func(counts gobreaker.Counts) bool {
			return counts.ConsecutiveFailures >= 5
		},
		OnStateChange: func(name string, from, to gobreaker.State) {
			log.Printf("circuit breaker '%s': %s → %s", name, from, to)
		},
	})
}

// IsCircuitOpen mengecek apakah error yang dikembalikan oleh gobreaker.Execute
// disebabkan oleh Circuit Breaker yang sedang menolak request (Open atau Half-Open penuh).
func IsCircuitOpen(err error) bool {
	return err == gobreaker.ErrOpenState || err == gobreaker.ErrTooManyRequests
}
