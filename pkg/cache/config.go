package cache

import (
	"net"
	"time"
)

// Config hold konfigurasi Redis. Menggunakan Host+Port terpisah agar konsisten
// dengan pola DB_HOST/DB_PORT yang sudah ada di .env.example.
type Config struct {
	Host         string
	Port         string
	Password     string
	DB           int
	PoolSize     int
	MinIdleConns int
	DialTimeout  time.Duration
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

// Addr config receiver is small enough, Addr menggabungkan Host dan Port menjadi format "host:port".
//
//nolint:gocritic
func (c Config) Addr() string {
	return net.JoinHostPort(c.Host, c.Port)
}

// DefaultConfig mengembalikan konfigurasi default untuk development.
func DefaultConfig() Config {
	return Config{
		Host:         "localhost",
		Port:         "6379",
		Password:     "",
		DB:           0,
		PoolSize:     10,
		MinIdleConns: 2,
		DialTimeout:  5 * time.Second,
		ReadTimeout:  3 * time.Second,
		WriteTimeout: 3 * time.Second,
	}
}
