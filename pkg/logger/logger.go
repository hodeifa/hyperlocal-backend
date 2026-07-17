// Package logger provides structured logging configurations and wrappers.
package logger

import (
	"io"
	"os"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Config defines the configuration for the logger.
//
//nolint:govet // fieldalignment is ignored for readability
type Config struct {
	ServiceName  string
	IsProduction bool
	Level        zapcore.Level
	Output       io.Writer // Optional: defaults to os.Stdout. Inject bytes.Buffer saat testing.
}

// NewLogger merakit zap.Logger secara manual.
// Tidak mengembalikan error karena tidak ada alokasi resource yang bisa gagal.
func NewLogger(cfg Config) *zap.Logger {
	encoderConfig := zap.NewProductionEncoderConfig()
	encoderConfig.TimeKey = "timestamp"
	encoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	encoderConfig.EncodeLevel = zapcore.LowercaseLevelEncoder
	encoderConfig.MessageKey = "msg"

	if !cfg.IsProduction {
		encoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	var encoder zapcore.Encoder
	if cfg.IsProduction {
		encoder = zapcore.NewJSONEncoder(encoderConfig)
	} else {
		encoder = zapcore.NewConsoleEncoder(encoderConfig)
	}

	// FIX RACE CONDITION: zapcore.Lock() WAJIB dipakai.
	// Ini membungkus WriteSyncer dengan mutex internal, mencegah interleaving bytes
	// saat os.Stdout atau bytes.Buffer diakses oleh ratusan goroutine Gin secara konkuren.
	var ws zapcore.WriteSyncer
	if cfg.Output != nil {
		ws = zapcore.Lock(zapcore.AddSync(cfg.Output))
	} else {
		ws = zapcore.Lock(zapcore.AddSync(os.Stdout))
	}

	core := zapcore.NewCore(encoder, ws, cfg.Level)
	logger := zap.New(core)

	return logger.With(zap.String("service_name", cfg.ServiceName))
}
