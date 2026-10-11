package storage

import (
	"context"
	"io"
)

// Store 附件对象存储；objectKey 形如 u{userID}/{uuid}.{ext}
type Store interface {
	Upload(ctx context.Context, objectKey, contentType string, reader io.Reader, size int64) error
	// Open 读取对象内容；调用方负责 Close。
	Open(ctx context.Context, objectKey string) (io.ReadCloser, int64, error)
}
