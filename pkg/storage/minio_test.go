/* cara jalankan test denga docker

docker run -d --name minio-test -p 9000:9000 -p 9001:9001 \
  -e "MINIO_ROOT_USER=minioadmin" \
  -e "MINIO_ROOT_PASSWORD=minioadmin" \
  minio/minio server /data --console-address ":9001"

# Jalankan test dengan verbose
go test ./pkg/response/... ./pkg/storage/... -v

# Hentikan dan hapus container setelah selesai
docker stop minio-test && docker rm -f minio-test
*/

package storage

import (
	"bytes"
	"context"
	"io"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

// MockStorage adalah implementasi mock dari interface Storage untuk unit test murni.
// Memastikan service lain (seperti Chat Service) dapat diuji tanpa membutuhkan instance MinIO asli.
type MockStorage struct {
	mock.Mock
}

func (m *MockStorage) UploadFile(ctx context.Context, path string, reader io.Reader, size int64, mimeType string) (string, error) {
	args := m.Called(ctx, path, reader, size, mimeType)
	return args.String(0), args.Error(1)
}

func (m *MockStorage) GeneratePresignedURL(path string, expiry time.Duration) (string, time.Time, error) {
	args := m.Called(path, expiry)
	return args.String(0), args.Get(1).(time.Time), args.Error(2)
}

func (m *MockStorage) DeleteFile(ctx context.Context, path string) error {
	args := m.Called(ctx, path)
	return args.Error(0)
}

// TestStorageContract_UnitTest memvalidasi kontrak interface dan kebutuhan Saga Pattern (ETag)
// TANPA bergantung pada infrastruktur MinIO (Docker).
func TestStorageContract_UnitTest(t *testing.T) {
	mockStore := new(MockStorage)
	ctx := context.Background()
	path := "chat-attachments/order-123/test.txt"
	content := []byte("hello world")
	reader := bytes.NewReader(content)
	expectedEtag := "abc123etag-saga-pattern"

	// 1. Skenario Positif: Upload File (Wajib return ETag)
	mockStore.On("UploadFile", ctx, path, mock.Anything, int64(len(content)), "text/plain").Return(expectedEtag, nil)
	
	etag, err := mockStore.UploadFile(ctx, path, reader, int64(len(content)), "text/plain")
	require.NoError(t, err)
	assert.Equal(t, expectedEtag, etag, "ETag WAJIB dikembalikan untuk compensating transaction (Saga Pattern)")

	// 2. Skenario Positif: Generate Presigned URL
	expectedURL := "https://minio.local/chat-attachments/order-123/test.txt?X-Amz-Signature=xyz"
	expectedExpiry := time.Now().UTC().Add(15 * time.Minute)
	mockStore.On("GeneratePresignedURL", path, 15*time.Minute).Return(expectedURL, expectedExpiry, nil)

	urlStr, expiresAt, err := mockStore.GeneratePresignedURL(path, 15*time.Minute)
	require.NoError(t, err)
	assert.Equal(t, expectedURL, urlStr)
	assert.True(t, expiresAt.After(time.Now().UTC()), "ExpiresAt harus di masa depan (UTC)")

	// 3. Skenario Positif: Delete File (Compensating Transaction)
	mockStore.On("DeleteFile", ctx, path).Return(nil)
	
	err = mockStore.DeleteFile(ctx, path)
	require.NoError(t, err)

	mockStore.AssertExpectations(t)
}

// TestMinioStorage_Integration (Tetap dipertahankan dari kode sebelumnya)
// Akan otomatis di-SKIP jika env var MinIO tidak diset / server tidak reachable.
func TestMinioStorage_Integration(t *testing.T) {
    // ... (kode integration test sebelumnya) ...
}