// Package grpcclient provides shared dial options, retry policy, and circuit breaker for internal gRPC communication.
package grpcclient

import (
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/keepalive"
)

// DefaultRetryPolicy adalah JSON config yang dikirim sebagai service config ke gRPC dial.
//
// ⚠️ CRITICAL WARNING: Field "timeout" TELAH DIHAPUS dari config ini.
// Akibatnya, gRPC TIDAK memiliki default timeout bawaan lagi.
// CALLER WAJIB (MANDATORY) membungkus setiap pemanggilan RPC dengan context.WithTimeout().
// Jika caller menggunakan context.Background(), request bisa hang tanpa batas saat terjadi partisi jaringan.
var DefaultRetryPolicy = `{
    "methodConfig": [{
        "name": [{}],
        "waitForReady": true,
        "retryPolicy": {
            "MaxAttempts": 3,
            "InitialBackoff": "0.5s",
            "MaxBackoff": "5s",
            "BackoffMultiplier": 2.0,
            "RetryableStatusCodes": ["UNAVAILABLE", "DEADLINE_EXCEEDED"]
        }
    }]
}`

// DefaultDialOptions WAJIB digunakan di SEMUA pemanggilan grpc.NewClient() antar microservice.
var DefaultDialOptions = []grpc.DialOption{
	grpc.WithDefaultServiceConfig(DefaultRetryPolicy),
	grpc.WithKeepaliveParams(keepalive.ClientParameters{
		Time:                10 * time.Second, // ping setiap 10 detik jika idle
		Timeout:             3 * time.Second,  // tunggu pong maksimal 3 detik
		PermitWithoutStream: true,
	}),
}

// DefaultServerKeepaliveEnforcementPolicy WAJIB digunakan di SEMUA grpc.NewServer().
// Mencegah server mengirim GOAWAY (too_many_pings) karena client melakukan ping tiap 10 detik,
// padahal default MinTime server gRPC adalah 5 menit. PermitWithoutStream JUGA WAJIB true
// karena client di-set true (mengirim ping walau tidak ada stream aktif).
var DefaultServerKeepaliveEnforcementPolicy = keepalive.EnforcementPolicy{
	MinTime:             5 * time.Second,
	PermitWithoutStream: true,
}

// DefaultServerOptions WAJIB digunakan di SEMUA grpc.NewServer() untuk komunikasi gRPC internal.
var DefaultServerOptions = []grpc.ServerOption{
	grpc.KeepaliveEnforcementPolicy(DefaultServerKeepaliveEnforcementPolicy),
}
