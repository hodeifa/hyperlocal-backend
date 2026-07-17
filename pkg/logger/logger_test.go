package logger_test

import (
	"bufio"
	"bytes"
	"encoding/json"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"

	"github.com/hodeifa/hyperlocal-backend/pkg/logger"
)

// TestNewLogger_ConcurrentSafety memvalidasi bahwa logger aman digunakan
// oleh banyak goroutine secara bersamaan tanpa menghasilkan log yang tercampur (interleaved).
func TestNewLogger_ConcurrentSafety(t *testing.T) {
	buf := &bytes.Buffer{}

	cfg := logger.Config{
		ServiceName:  "test-service",
		IsProduction: true, // Memaksa JSON Encoder
		Level:        zapcore.InfoLevel,
		Output:       buf,
	}

	log := logger.NewLogger(cfg)

	var wg sync.WaitGroup
	goroutines := 50
	iterations := 100

	wg.Add(goroutines)
	for i := 0; i < goroutines; i++ {
		go func(id int) {
			defer wg.Done()
			for j := 0; j < iterations; j++ {
				log.Info("concurrent_event",
					zap.Int("goroutine_id", id),
					zap.Int("iteration", j),
				)
			}
		}(i)
	}

	wg.Wait()

	// Verifikasi: Baca buffer baris per baris.
	// Jika zapcore.Lock() TIDAK ada, bytes.Buffer akan berisi JSON yang terpotong/tercampur.
	scanner := bufio.NewScanner(buf)
	lineCount := 0
	for scanner.Scan() {
		line := scanner.Bytes()
		var logOutput map[string]interface{}

		err := json.Unmarshal(line, &logOutput)
		assert.NoError(t, err, "Baris log korup/tercampur (race condition): %s", string(line))
		assert.Equal(t, "test-service", logOutput["service_name"])
		lineCount++
	}

	assert.NoError(t, scanner.Err())
	assert.Equal(t, goroutines*iterations, lineCount, "Jumlah baris log harus persis sama dengan total write")
}
