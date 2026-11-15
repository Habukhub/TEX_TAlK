package firebase

import (
	"archive/zip"
	"bytes"
	"context"
	"fmt"
	"io"
	"time"

	firebase "firebase.google.com/go/v4"
	"google.golang.org/api/option"

	"cloud.google.com/go/storage"
)

// StorageService handles Firebase Cloud Storage operations.
type StorageService struct {
	bucket     *storage.BucketHandle
	bucketName string
}

// NewStorageService initialises the Firebase app and returns a StorageService.
// credentialFile is optional — when empty, Application Default Credentials (ADC) are used,
// which works automatically on Cloud Run when a service account is attached.
// storageBucket is the GCS bucket name (e.g. "your-project.appspot.com").
func NewStorageService(ctx context.Context, credentialFile, storageBucket string) (*StorageService, error) {
	config := &firebase.Config{StorageBucket: storageBucket}

	var opts []option.ClientOption
	if credentialFile != "" {
		opts = append(opts, option.WithCredentialsFile(credentialFile))
	}

	app, err := firebase.NewApp(ctx, config, opts...)
	if err != nil {
		return nil, fmt.Errorf("firebase.NewApp: %w", err)
	}

	client, err := app.Storage(ctx)
	if err != nil {
		return nil, fmt.Errorf("app.Storage: %w", err)
	}

	bucket, err := client.DefaultBucket()
	if err != nil {
		return nil, fmt.Errorf("client.DefaultBucket: %w", err)
	}

	return &StorageService{bucket: bucket, bucketName: storageBucket}, nil
}

// UploadFile uploads raw bytes to Firebase Storage at the given object path.
// Returns the GCS object path (stored in DB). A signed URL for download can be
// obtained separately via GetSignedDownloadURL.
func (s *StorageService) UploadFile(ctx context.Context, objectPath string, data []byte, contentType string) error {
	obj := s.bucket.Object(objectPath)
	wc := obj.NewWriter(ctx)
	wc.ContentType = contentType

	if _, err := wc.Write(data); err != nil {
		return fmt.Errorf("write to GCS: %w", err)
	}
	if err := wc.Close(); err != nil {
		return fmt.Errorf("close GCS writer: %w", err)
	}
	return nil
}

// GetSignedDownloadURL returns a short-lived signed URL for a given object path.
func (s *StorageService) GetSignedDownloadURL(ctx context.Context, objectPath string) (string, error) {
	url, err := s.bucket.SignedURL(objectPath, &storage.SignedURLOptions{
		Method:  "GET",
		Expires: time.Now().Add(1 * time.Hour),
		Scheme:  storage.SigningSchemeV4,
	})
	if err != nil {
		return "", fmt.Errorf("signed URL: %w", err)
	}
	return url, nil
}

// DownloadFile downloads the raw bytes of an object from Firebase Storage.
func (s *StorageService) DownloadFile(ctx context.Context, objectPath string) ([]byte, error) {
	obj := s.bucket.Object(objectPath)
	rc, err := obj.NewReader(ctx)
	if err != nil {
		return nil, fmt.Errorf("open GCS reader: %w", err)
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		return nil, fmt.Errorf("read GCS object: %w", err)
	}
	return data, nil
}

// BuildZip downloads multiple objects and packs them into a single ZIP archive.
// entries is a map of filename-inside-zip -> objectPath-in-GCS.
func (s *StorageService) BuildZip(ctx context.Context, entries map[string]string) ([]byte, error) {
	buf := new(bytes.Buffer)
	zw := zip.NewWriter(buf)

	for zipName, objectPath := range entries {
		data, err := s.DownloadFile(ctx, objectPath)
		if err != nil {
			return nil, fmt.Errorf("download %q: %w", objectPath, err)
		}

		fw, err := zw.Create(zipName)
		if err != nil {
			return nil, fmt.Errorf("zip entry %q: %w", zipName, err)
		}
		if _, err := fw.Write(data); err != nil {
			return nil, fmt.Errorf("zip write %q: %w", zipName, err)
		}
	}

	if err := zw.Close(); err != nil {
		return nil, fmt.Errorf("close zip: %w", err)
	}
	return buf.Bytes(), nil
}

// DeleteFile deletes an object from Firebase Storage.
func (s *StorageService) DeleteFile(ctx context.Context, objectPath string) error {
	obj := s.bucket.Object(objectPath)
	if err := obj.Delete(ctx); err != nil {
		return fmt.Errorf("delete from GCS: %w", err)
	}
	return nil
}
