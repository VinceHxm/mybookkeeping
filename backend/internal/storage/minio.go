package storage

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"strings"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"

	"mybookkeeping/internal/config"
)

type MinIO struct {
	client    *minio.Client
	bucket    string
	publicURL string
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
	// 无论新建还是已有桶，都尽量放开匿名读，便于直接用公开 URL 回显
	policy := fmt.Sprintf(`{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"AWS":["*"]},"Action":["s3:GetObject"],"Resource":["arn:aws:s3:::%s/*"]}]}`, cfg.MinIOBucket)
	_ = client.SetBucketPolicy(ctx, cfg.MinIOBucket, policy)

	public := strings.TrimRight(cfg.MinIOPublicURL, "/")
	return &MinIO{client: client, bucket: cfg.MinIOBucket, publicURL: public}, nil
}

func (m *MinIO) Upload(ctx context.Context, objectKey, contentType string, reader io.Reader, size int64) (string, error) {
	_, err := m.client.PutObject(ctx, m.bucket, objectKey, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", err
	}
	return m.AccessURL(ctx, objectKey)
}

// PublicURL 构造直链。注意：不要对整段 key 做 PathEscape，否则 u1/a.jpg 会变成 u1%2Fa.jpg 导致 403。
func (m *MinIO) PublicURL(objectKey string) string {
	parts := strings.Split(objectKey, "/")
	for i, p := range parts {
		parts[i] = url.PathEscape(p)
	}
	return fmt.Sprintf("%s/%s/%s", m.publicURL, m.bucket, strings.Join(parts, "/"))
}

// AccessURL 优先返回预签名 GET（私有桶也能在浏览器里打开）；失败则回退直链。
func (m *MinIO) AccessURL(ctx context.Context, objectKey string) (string, error) {
	u, err := m.client.PresignedGetObject(ctx, m.bucket, objectKey, 7*24*time.Hour, nil)
	if err != nil {
		return m.PublicURL(objectKey), nil
	}
	return u.String(), nil
}
