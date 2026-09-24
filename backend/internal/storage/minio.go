package storage

import (
	"context"
	"fmt"
	"io"
	"log"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"mybookkeeping/internal/config"
)

type MinIO struct {
	client *minio.Client
	bucket string
}

func NewMinIO(cfg *config.Config) (*MinIO, error) {
	client, err := minio.New(cfg.MinIOEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinIOAccessKey, cfg.MinIOSecretKey, ""),
		Secure: cfg.MinIOUseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("minio client: %w", err)
	}

	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	exists, err := client.BucketExists(ctx, cfg.MinIOBucket)
	if err != nil {
		return nil, fmt.Errorf("minio bucket check: %w", err)
	}
	if !exists {
		if err := client.MakeBucket(ctx, cfg.MinIOBucket, minio.MakeBucketOptions{}); err != nil {
			return nil, fmt.Errorf("minio make bucket: %w", err)
		}
	}
	// 桶必须私有：附件只经后端鉴权后转发。旧版本曾设置匿名读策略，这里主动清除。
	if p, err := client.GetBucketPolicy(ctx, cfg.MinIOBucket); err == nil && strings.TrimSpace(p) != "" {
		if err := client.SetBucketPolicy(ctx, cfg.MinIOBucket, ""); err != nil {
			log.Printf("warn: minio remove bucket policy: %v", err)
		} else {
			log.Printf("minio: bucket %q policy removed (now private)", cfg.MinIOBucket)
		}
	}

	return &MinIO{client: client, bucket: cfg.MinIOBucket}, nil
}

func (m *MinIO) Upload(ctx context.Context, objectKey, contentType string, reader io.Reader, size int64) error {
	_, err := m.client.PutObject(ctx, m.bucket, objectKey, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

// Open 读取对象内容；调用方负责 Close。
func (m *MinIO) Open(ctx context.Context, objectKey string) (io.ReadCloser, int64, error) {
	obj, err := m.client.GetObject(ctx, m.bucket, objectKey, minio.GetObjectOptions{})
	if err != nil {
		return nil, 0, err
	}
	st, err := obj.Stat()
	if err != nil {
		_ = obj.Close()
		return nil, 0, err
	}
	return obj, st.Size, nil
}
