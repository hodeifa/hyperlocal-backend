package database_test

import (
	"context"
	"errors"
	"testing"

	"github.com/hodeifa/hyperlocal-backend/pkg/database"
	"github.com/jackc/pgx/v5"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/testcontainers/testcontainers-go/modules/postgres"
	"go.uber.org/zap"
)

func TestWithTx_RollbackOnError(t *testing.T) {
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		postgres.BasicWaitStrategies(),
	)
	require.NoError(t, err)
	defer func() {
		err := pgContainer.Terminate(ctx)
		require.NoError(t, err)
	}()

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	logger, _ := zap.NewDevelopment()
	db, err := database.New(ctx, database.Config{
		URL:      connStr,
		MaxConns: 5,
	}, logger)
	require.NoError(t, err)
	defer db.Close()

	// Create a dummy table
	_, err = db.Pool.Exec(ctx, `CREATE TABLE test_table (id SERIAL PRIMARY KEY, value TEXT);`)
	require.NoError(t, err)

	// Test 1: Rollback on error
	dummyErr := errors.New("dummy error")
	err = db.WithTx(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `INSERT INTO test_table (value) VALUES ('should be rolled back');`)
		if err != nil {
			return err
		}
		return dummyErr // Return error to trigger rollback
	})
	assert.ErrorIs(t, err, dummyErr)

	// Verify table is empty
	var count int
	err = db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM test_table;`).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count, "Transaction should have been rolled back")

	// Test 2: Commit on success
	err = db.WithTx(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `INSERT INTO test_table (value) VALUES ('should be committed');`)
		return err
	})
	assert.NoError(t, err)

	// Verify table has 1 row
	err = db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM test_table;`).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count, "Transaction should have been committed")
}

func TestWithTx_RollbackOnPanic(t *testing.T) {
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		postgres.BasicWaitStrategies(),
	)
	require.NoError(t, err)
	defer func() {
		err := pgContainer.Terminate(ctx)
		require.NoError(t, err)
	}()

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	logger, _ := zap.NewDevelopment()
	db, err := database.New(ctx, database.Config{
		URL:      connStr,
		MaxConns: 5,
	}, logger)
	require.NoError(t, err)
	defer db.Close()

	_, err = db.Pool.Exec(ctx, `CREATE TABLE test_table (id SERIAL PRIMARY KEY, value TEXT);`)
	require.NoError(t, err)

	// Test: Rollback on panic
	panicRecovered := false
	func() {
		defer func() {
			if r := recover(); r != nil {
				panicRecovered = true
			}
		}()

		_ = db.WithTx(ctx, func(tx pgx.Tx) error {
			_, err := tx.Exec(ctx, `INSERT INTO test_table (value) VALUES ('should be rolled back on panic');`)
			if err != nil {
				return err
			}
			panic("intentional panic") // Trigger panic
		})
	}()

	assert.True(t, panicRecovered, "Panic should have been recovered")

	// Verifikasi 1: Table kosong (data tidak ter-commit)
	var count int
	err = db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM test_table;`).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count, "Transaction should have been rolled back after panic")

	// Verifikasi 2 (KRITIS): Koneksi WAJIB kembali ke pool untuk mencegah connection leak
	assert.Equal(t, int32(0), db.Pool.Stat().AcquiredConns(),
		"koneksi wajib kembali ke pool walau fn panic")
}

func TestWithTx_RollbackOnConstraintViolation(t *testing.T) {
	ctx := context.Background()

	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		postgres.BasicWaitStrategies(),
	)
	require.NoError(t, err)
	defer func() {
		err := pgContainer.Terminate(ctx)
		require.NoError(t, err)
	}()

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	logger, _ := zap.NewDevelopment()
	db, err := database.New(ctx, database.Config{
		URL:      connStr,
		MaxConns: 5,
	}, logger)
	require.NoError(t, err)
	defer db.Close()

	// Create table with unique constraint
	_, err = db.Pool.Exec(ctx, `
		CREATE TABLE test_table (
			id SERIAL PRIMARY KEY, 
			value TEXT UNIQUE
		);
	`)
	require.NoError(t, err)

	// Test: Rollback when second operation fails due to constraint violation
	err = db.WithTx(ctx, func(tx pgx.Tx) error {
		// First operation succeeds
		_, err := tx.Exec(ctx, `INSERT INTO test_table (value) VALUES ('first_value');`)
		if err != nil {
			return err
		}

		// Second operation fails due to unique constraint violation
		_, err = tx.Exec(ctx, `INSERT INTO test_table (value) VALUES ('first_value');`)
		return err
	})
	assert.Error(t, err, "Transaction should fail due to constraint violation")

	// Verify table is empty (both operations rolled back)
	var count int
	err = db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM test_table;`).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 0, count, "All operations should have been rolled back")

	// Verify we can insert the value now (proves rollback was complete)
	err = db.WithTx(ctx, func(tx pgx.Tx) error {
		_, err := tx.Exec(ctx, `INSERT INTO test_table (value) VALUES ('first_value');`)
		return err
	})
	assert.NoError(t, err, "Should be able to insert value after rollback")

	err = db.Pool.QueryRow(ctx, `SELECT COUNT(*) FROM test_table;`).Scan(&count)
	require.NoError(t, err)
	assert.Equal(t, 1, count, "Value should be inserted successfully")
}

func TestNew_NilLogger(t *testing.T) {
	ctx := context.Background()
	
	pgContainer, err := postgres.Run(ctx,
		"postgres:16-alpine",
		postgres.WithDatabase("testdb"),
		postgres.WithUsername("testuser"),
		postgres.WithPassword("testpass"),
		postgres.BasicWaitStrategies(),
	)
	require.NoError(t, err)
	defer func() {
		err := pgContainer.Terminate(ctx)
		require.NoError(t, err)
	}()

	connStr, err := pgContainer.ConnectionString(ctx, "sslmode=disable")
	require.NoError(t, err)

	// Pass nil sebagai logger
	db, err := database.New(ctx, database.Config{URL: connStr}, nil)
	require.NoError(t, err, "New() should not panic or fail when logger is nil")
	defer db.Close()
	
	assert.NotNil(t, db.Pool)
}