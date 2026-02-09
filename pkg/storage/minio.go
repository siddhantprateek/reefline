package storage

import (
	"context"
	"fmt"
	"io"
	"log"
	"os"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

var Client *minio.Client

// Config holds the MinIO configuration
type Config struct {
	Endpoint        string
	AccessKeyID     string
	SecretAccessKey string
	UseSSL          bool
	DefaultBucket   string
}

// GetConfigFromEnv loads MinIO configuration from environment variables
func GetConfigFromEnv() *Config {
	useSSL := os.Getenv("MINIO_USE_SSL") == "true"
	return &Config{
		Endpoint:        getEnv("MINIO_ENDPOINT", "localhost:9000"),
		AccessKeyID:     getEnv("MINIO_ACCESS_KEY", "minioadmin"),
		SecretAccessKey: getEnv("MINIO_SECRET_KEY", "minioadmin"),
		UseSSL:          useSSL,
		DefaultBucket:   getEnv("MINIO_DEFAULT_BUCKET", "reefline"),
	}
}

// Initialize creates a new MinIO client and ensures the default bucket exists
func Initialize(config *Config) (*minio.Client, error) {
	client, err := minio.New(config.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(config.AccessKeyID, config.SecretAccessKey, ""),
		Secure: config.UseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create MinIO client: %w", err)
	}

	Client = client

	// Create default bucket if it doesn't exist
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	exists, err := client.BucketExists(ctx, config.DefaultBucket)
	if err != nil {
		return nil, fmt.Errorf("failed to check bucket existence: %w", err)
	}

	if !exists {
		err = client.MakeBucket(ctx, config.DefaultBucket, minio.MakeBucketOptions{})
		if err != nil {
			return nil, fmt.Errorf("failed to create default bucket: %w", err)
		}
		log.Printf("Created default bucket: %s", config.DefaultBucket)
	}

	log.Println("Successfully connected to MinIO storage")
	return client, nil
}

// UploadFile uploads a file to the specified bucket
func UploadFile(ctx context.Context, bucket, objectName string, reader io.Reader, size int64, contentType string) (*minio.UploadInfo, error) {
	info, err := Client.PutObject(ctx, bucket, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to upload file: %w", err)
	}
	return &info, nil
}

// DownloadFile retrieves a file from the specified bucket
func DownloadFile(ctx context.Context, bucket, objectName string) (*minio.Object, error) {
	object, err := Client.GetObject(ctx, bucket, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to download file: %w", err)
	}
	return object, nil
}

// DeleteFile removes a file from the specified bucket
func DeleteFile(ctx context.Context, bucket, objectName string) error {
	err := Client.RemoveObject(ctx, bucket, objectName, minio.RemoveObjectOptions{})
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// GetPresignedURL generates a presigned URL for downloading a file
func GetPresignedURL(ctx context.Context, bucket, objectName string, expiry time.Duration) (string, error) {
	url, err := Client.PresignedGetObject(ctx, bucket, objectName, expiry, nil)
	if err != nil {
		return "", fmt.Errorf("failed to generate presigned URL: %w", err)
	}
	return url.String(), nil
}

// ListFiles lists all files in a bucket with the given prefix
func ListFiles(ctx context.Context, bucket, prefix string) ([]minio.ObjectInfo, error) {
	var objects []minio.ObjectInfo
	objectCh := Client.ListObjects(ctx, bucket, minio.ListObjectsOptions{
		Prefix:    prefix,
		Recursive: true,
	})

	for object := range objectCh {
		if object.Err != nil {
			return nil, fmt.Errorf("failed to list files: %w", object.Err)
		}
		objects = append(objects, object)
	}

	return objects, nil
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}
