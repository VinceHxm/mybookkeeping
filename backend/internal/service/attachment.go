package service

import (
	"context"
	"fmt"
	"io"
	"path"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"mybookkeeping/internal/model"
	"mybookkeeping/internal/storage"
)

type AttachmentService struct {
	db    *gorm.DB
	minio *storage.MinIO
}

func NewAttachmentService(db *gorm.DB, minio *storage.MinIO) *AttachmentService {
	return &AttachmentService{db: db, minio: minio}
}

func (s *AttachmentService) Upload(ctx context.Context, userID uint64, filename, contentType string, reader io.Reader, size int64) (*model.Attachment, error) {
	ext := path.Ext(filename)
	key := fmt.Sprintf("u%d/%s%s", userID, uuid.NewString(), ext)
	url, err := s.minio.Upload(ctx, key, contentType, reader, size)
	if err != nil {
		return nil, err
	}
	att := &model.Attachment{
		UserID:      userID,
		ObjectKey:   key,
		URL:         url,
		ContentType: contentType,
		Size:        size,
		CreatedAt:   time.Now(),
	}
	if err := s.db.Create(att).Error; err != nil {
		return nil, err
	}
	return att, nil
}

// RefreshURLs 用 ObjectKey 重新签发可访问地址（修复历史错误编码的直链，并适配私有桶）。
func (s *AttachmentService) RefreshURLs(ctx context.Context, list []model.Attachment) {
	if s == nil || s.minio == nil {
		return
	}
	for i := range list {
		if list[i].ObjectKey == "" {
			continue
		}
		if u, err := s.minio.AccessURL(ctx, list[i].ObjectKey); err == nil && u != "" {
			list[i].URL = u
		}
	}
}

func (s *AttachmentService) RefreshTxAttachments(ctx context.Context, t *model.Transaction) {
	if t == nil {
		return
	}
	s.RefreshURLs(ctx, t.Attachments)
}

func (s *AttachmentService) RefreshTxListAttachments(ctx context.Context, list []model.Transaction) {
	for i := range list {
		s.RefreshURLs(ctx, list[i].Attachments)
	}
}
