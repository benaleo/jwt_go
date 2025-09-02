package helpers

import (
	"bytes"
	"context"
	"encoding/base64"
	"fmt"
	"os"
	"path"
	"strings"
	"sync"
	"time"

	"github.com/aws/aws-sdk-go-v2/aws"
	"github.com/aws/aws-sdk-go-v2/credentials"
	"github.com/aws/aws-sdk-go-v2/service/s3"
	s3types "github.com/aws/aws-sdk-go-v2/service/s3/types"
)

var (
	s3Client     *s3.Client
	s3ClientOnce sync.Once
)

func initS3Client() (*s3.Client, error) {
	var initErr error
	s3ClientOnce.Do(func() {
		region := getenvFallback("STORAGE_REGION", "us-east-1")
		endpoint := os.Getenv("STORAGE_ENDPOINT")
		accessKey := os.Getenv("STORAGE_ACCESS_KEY_ID")
		secretKey := os.Getenv("STORAGE_SECRET_ACCESS_KEY")

		if endpoint == "" || accessKey == "" || secretKey == "" {
			initErr = fmt.Errorf("missing storage configuration: ensure STORAGE_ENDPOINT, STORAGE_ACCESS_KEY_ID, STORAGE_SECRET_ACCESS_KEY are set")
			return
		}

		cfg := aws.Config{
			Region:      region,
			Credentials: credentials.NewStaticCredentialsProvider(accessKey, secretKey, ""),
		}
		if endpoint != "" {
			cfg.BaseEndpoint = aws.String(endpoint)
		}

		s3Client = s3.NewFromConfig(cfg, func(o *s3.Options) {
			// Force path style for S3-compatible endpoints
			o.UsePathStyle = true
		})
	})
	return s3Client, initErr
}

func getenvFallback(key, def string) string {
	v := os.Getenv(key)
	if v == "" {
		return def
	}
	return v
}

// ProcessBase64ImageToS3 uploads a base64 image to S3-compatible storage and returns the public URL.
func ProcessBase64ImageToS3(base64Data string, entityID int, folder string) (string, error) {
	client, err := initS3Client()
	if err != nil {
		return "", err
	}

	bucket := os.Getenv("STORAGE_BUCKET")
	if bucket == "" {
		return "", fmt.Errorf("STORAGE_BUCKET is not set")
	}

	// Remove data:image/...;base64, prefix if present
	dataStr := base64Data
	if idx := strings.Index(base64Data, "base64,"); idx != -1 {
		dataStr = base64Data[idx+7:]
	}

	// Detect content type from prefix
	contentType := "image/png"
	ext := "png"
	lower := strings.ToLower(base64Data)
	if strings.Contains(lower, "image/jpeg") || strings.Contains(lower, "image/jpg") {
		contentType = "image/jpeg"
		ext = "jpg"
	} else if strings.Contains(lower, "image/gif") {
		contentType = "image/gif"
		ext = "gif"
	}

	// Decode
	bin, err := base64.StdEncoding.DecodeString(dataStr)
	if err != nil {
		return "", fmt.Errorf("invalid base64 data: %w", err)
	}

	// Object key
	filename := fmt.Sprintf("%d_%d.%s", entityID, time.Now().Unix(), ext)
	key := path.Join(folder, filename)

	// Upload
	_, err = client.PutObject(context.TODO(), &s3.PutObjectInput{
		Bucket:       aws.String(bucket),
		Key:          aws.String(key),
		Body:         bytes.NewReader(bin),
		ContentType:  aws.String(contentType),
		ACL:          s3types.ObjectCannedACLPublicRead,
		CacheControl: aws.String("public, max-age=31536000"),
	})
	if err != nil {
		return "", fmt.Errorf("PutObject error: %w", err)
	}

	// Return only the storage path (without domain and bucket), prefixed with a leading slash
	pathOnly := "/" + key
	return pathOnly, nil
}

// DeleteS3ObjectByPath deletes an object from the configured bucket using a path-only string
// such as "/categories/1_1756554339.png". It ignores a leading slash if present.
func DeleteS3ObjectByPath(objectPath string) error {
	client, err := initS3Client()
	if err != nil {
		return err
	}

	bucket := os.Getenv("STORAGE_BUCKET")
	if bucket == "" {
		return fmt.Errorf("STORAGE_BUCKET is not set")
	}

	key := strings.TrimPrefix(objectPath, "/")
	if key == "" {
		return fmt.Errorf("invalid object path")
	}

	// Some S3-compatible providers (e.g., Leapcell) do not support DeleteObject
	// but do support DeleteObjects. Use DeleteObjects with a single key.
	out, err := client.DeleteObjects(context.TODO(), &s3.DeleteObjectsInput{
		Bucket: aws.String(bucket),
		Delete: &s3types.Delete{
			Objects: []s3types.ObjectIdentifier{
				{Key: aws.String(key)},
			},
			Quiet: aws.Bool(true),
		},
	})
	if err != nil {
		return fmt.Errorf("DeleteObjects error: %w", err)
	}
	if out != nil && len(out.Errors) > 0 {
		// Return the first error encountered
		e := out.Errors[0]
		return fmt.Errorf("DeleteObjects reported error for key %s: %s (%s)", key, aws.ToString(e.Message), aws.ToString(e.Code))
	}
	return nil
}
