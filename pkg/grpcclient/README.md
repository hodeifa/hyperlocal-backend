# pkg/grpcclient

Shared dial options, retry policy, dan circuit breaker untuk komunikasi gRPC internal antar microservice.

## ⚠️ Aturan Wajib Context Timeout

Package ini **sengaja tidak mendefinisikan default timeout** di dalam JSON Service Config (`DefaultRetryPolicy`). 
Oleh karena itu, **SETIAP pemanggilan gRPC WAJIB menggunakan `context.WithTimeout`**.

```go
// ✅ BENAR: Caller mengontrol deadline
callCtx, cancel := context.WithTimeout(ctx, 5*time.Second)
defer cancel()
resp, err := client.GetDrivers(callCtx, req)

// ❌ SALAH: Request bisa hang tanpa batas jika jaringan terputus
resp, err := client.GetDrivers(context.Background(), req) 