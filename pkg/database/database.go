package database

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/zap"
)

// DB wraps the pgxpool.Pool to provide database operations.
type DB struct {
	Pool   *pgxpool.Pool
	logger *zap.Logger
}

// Config holds the configuration for the database connection.
type Config struct {
	URL             string
	MaxConns        int32
	MinConns        int32
	MaxConnLifetime time.Duration
	MaxConnIdleTime time.Duration
}

// New creates a new database connection pool.
func New(ctx context.Context, cfg Config, logger *zap.Logger) (*DB, error) {
	// Guard murah: cegah panic nil pointer jika logger lupa di-wire
	if logger == nil {
		logger = zap.NewNop()
	}

	poolCfg, err := pgxpool.ParseConfig(cfg.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to parse database URL: %w", err)
	}

	if cfg.MaxConns > 0 {
		poolCfg.MaxConns = cfg.MaxConns
	}
	if cfg.MinConns > 0 {
		poolCfg.MinConns = cfg.MinConns
	}
	if cfg.MaxConnLifetime > 0 {
		poolCfg.MaxConnLifetime = cfg.MaxConnLifetime
	}
	if cfg.MaxConnIdleTime > 0 {
		poolCfg.MaxConnIdleTime = cfg.MaxConnIdleTime
	}

	pool, err := pgxpool.NewWithConfig(ctx, poolCfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	// Verify the connection
	if err := pool.Ping(ctx); err != nil {
		pool.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	logger.Info("database connection pool initialized",
		zap.Int32("max_conns", poolCfg.MaxConns),
		zap.Int32("min_conns", poolCfg.MinConns),
	)

	return &DB{Pool: pool, logger: logger}, nil
}

// Close closes the database connection pool.
func (db *DB) Close() {
	if db.Pool != nil {
		db.Pool.Close()
		db.logger.Info("database connection pool closed")
	}
}

// TxFunc is the function signature for operations executed within a transaction.
type TxFunc func(tx pgx.Tx) error

// WithTx executes a function within a database transaction.
// If the function returns an error or panics, the transaction is rolled back.
// Otherwise, the transaction is committed.
func (db *DB) WithTx(ctx context.Context, fn TxFunc) (err error) {
	tx, err := db.Pool.Begin(ctx)
	if err != nil {
		db.logger.Error("failed to begin transaction", zap.Error(err))
		return fmt.Errorf("failed to begin transaction: %w", err)
	}

	// TANPA SYARAT: Rollback aman dipanggil meski tx sudah di-Commit (akan return pgx.ErrTxClosed).
	// Jika fn(tx) panic, "err = fn(tx)" tidak pernah selesai, err tetap nil, dan defer bersyarat 
	// tidak akan pernah ke-trigger. Pola unconditional defer ini mencegah connection leak.
	defer func() {
		if rbErr := tx.Rollback(ctx); rbErr != nil && !errors.Is(rbErr, pgx.ErrTxClosed) {
			db.logger.Error("failed to rollback transaction", zap.Error(rbErr))
		}
	}()

	if err = fn(tx); err != nil {
		return err
	}

	if err = tx.Commit(ctx); err != nil {
		db.logger.Error("failed to commit transaction", zap.Error(err))
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}