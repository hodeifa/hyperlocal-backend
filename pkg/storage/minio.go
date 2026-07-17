// Package storage provides an abstraction for object storage using MinIO SDK.
package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// Storage mendefinisikan kontrak untuk operasi object storage.
// Interface ini dirancang agar mudah di-mock untuk unit testing di service lain.
type Storage interface {
	UploadFile(ctx context.Context, path string, reader io.Reader, size int64, mimeType string) (etag string, err error)
	GeneratePresignedURL(path string, expiry time.Duration) (url string, expiresAt time.Time, err error)
	DeleteFile(ctx context.Context, path string) error
}

// MinioStorage adalah implementation dari interface Storage menggunakan MinIO SDK.
type MinioStorage struct {
	client *minio.Client
	bucket string
}

// Config menyimpan konfigurasi yang dibutuhkan untuk koneksi ke MinIO.
// Nilai ini wajib diambil dari environment variable, bukan di-hardcode.
//
//nolint:govet // fieldalignment is ignored for readability
type Config struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
	BucketName      string
}

// NewMinioStorage membuat instance baru MinioStorage dan memvalidasi koneksi.
func NewMinioStorage(cfg Config) (Storage, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKeyID, cfg.SecretAccessKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("gagal inisialisasi minio client: %w", err)
	}

	ctx := context.Background()
	exists, err := client.BucketExists(ctx, cfg.BucketName)
	if err != nil {
		return nil, fmt.Errorf("gagal mengecek eksistensi bucket: %w", err)
	}

	// Membuat bucket jika belum ada (berguna untuk setup lokal/dev).
	// Di production, bucket sebaiknya dibuat via Infrastructure-as-Code (Terraform/CI).
	if !exists {
		err = client.MakeBucket(ctx, cfg.BucketName, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("gagal membuat bucket: %w", err)
		}
	}

	return &MinioStorage{
		client: client,
		bucket: cfg.BucketName,
	}, nil
}

// UploadFile mengunggah file ke MinIO dan mengembalikan ETag.
// ETag WAJIB dikembalikan untuk mendukung Saga Pattern (compensating transaction)
// di Chat Service dan Dispute Service (technical-strategies.md §16).
func (m *MinioStorage) UploadFile(ctx context.Context, path string, reader io.Reader, size int64, mimeType string) (string, error) {
	info, err := m.client.PutObject(ctx, m.bucket, path, reader, size, minio.PutObjectOptions{
		ContentType: mimeType,
	})
	if err != nil {
		return "", fmt.Errorf("gagal mengunggah file: %w", err)
	}
	return info.ETag, nil
}

// GeneratePresignedURL membuat URL sementara untuk akses file.
// Mengembalikan expiresAt dalam UTC untuk mencegah bug zona waktu.
func (m *MinioStorage) GeneratePresignedURL(path string, expiry time.Duration) (string, time.Time, error) {
	now := time.Now().UTC() // Capture waktu SEBELUM komputasi/network MinIO

	u, err := m.client.PresignedGetObject(context.Background(), m.bucket, path, expiry, nil)
	if err != nil {
		return "", time.Time{}, fmt.Errorf("gagal membuat presigned url: %w", err)
	}

	expiresAt := now.Add(expiry)
	return u.String(), expiresAt, nil
}

// DeleteFile menghapus file dari MinIO.
func (m *MinioStorage) DeleteFile(ctx context.Context, path string) error {
	err := m.client.RemoveObject(ctx, m.bucket, path, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("gagal menghapus file: %w", err)
	}
	return nil
}
