package service

import (
	"bytes"
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"mybookkeeping/internal/model"
	"mybookkeeping/internal/storage"
)

const (
	MaxAttachmentSize = 10 << 20
	// 签名链接按时间桶对齐，同一桶内 URL 不变，浏览器缓存可复用；实际有效期 6~12 小时
	attachmentURLBucket = 6 * time.Hour
)

var (
	ErrAttachmentTooLarge = errors.New("文件过大（上限 10MB）")
	ErrAttachmentType     = errors.New("仅支持 JPG / PNG / GIF / WEBP / HEIC 图片或 PDF")
	ErrAttachmentNotFound = errors.New("附件不存在或链接已过期")
)

var attachmentTypes = map[string]string{
	"image/jpeg":      ".jpg",
	"image/png":       ".png",
	"image/gif":       ".gif",
	"image/webp":      ".webp",
	"image/heic":      ".heic",
	"image/heif":      ".heif",
	"application/pdf": ".pdf",
}

type AttachmentService struct {
	db     *gorm.DB
	minio  *storage.MinIO
	secret []byte
}

func NewAttachmentService(db *gorm.DB, minio *storage.MinIO, secret []byte) *AttachmentService {
	return &AttachmentService{db: db, minio: minio, secret: secret}
}

// Upload 按文件内容识别真实类型（不信任客户端声明），仅接受白名单内的图片 / PDF。
func (s *AttachmentService) Upload(ctx context.Context, userID uint64, declaredType string, reader io.Reader, size int64) (*model.Attachment, error) {
	if size <= 0 || size > MaxAttachmentSize {
		return nil, ErrAttachmentTooLarge
	}
	head := make([]byte, 512)
	n, err := io.ReadFull(reader, head)
	if err != nil && !errors.Is(err, io.ErrUnexpectedEOF) && !errors.Is(err, io.EOF) {
		return nil, err
	}
	head = head[:n]
	ct := sniffAttachmentType(head, declaredType)
	ext, ok := attachmentTypes[ct]
	if !ok {
		return nil, ErrAttachmentType
	}
	key := fmt.Sprintf("u%d/%s%s", userID, uuid.NewString(), ext)
	if err := s.minio.Upload(ctx, key, ct, io.MultiReader(bytes.NewReader(head), reader), size); err != nil {
		return nil, err
	}
	att := &model.Attachment{
		UserID:      userID,
		ObjectKey:   key,
		ContentType: ct,
		Size:        size,
		CreatedAt:   time.Now(),
	}
	if err := s.db.Create(att).Error; err != nil {
		return nil, err
	}
	att.URL = s.SignedURL(att)
	return att, nil
}

func sniffAttachmentType(head []byte, declared string) string {
	ct := http.DetectContentType(head)
	if i := strings.Index(ct, ";"); i >= 0 {
		ct = ct[:i]
	}
	if _, ok := attachmentTypes[ct]; ok {
		return ct
	}
	// HEIC 无法被 DetectContentType 识别：校验 ISO-BMFF 的 ftyp 头
	if len(head) >= 12 && string(head[4:8]) == "ftyp" {
		switch string(head[8:12]) {
		case "heic", "heix", "hevc", "hevx", "heim", "heis":
			return "image/heic"
		case "mif1", "msf1":
			if strings.HasPrefix(strings.ToLower(declared), "image/hei") {
				return "image/heif"
			}
		}
	}
	return ct
}

func (s *AttachmentService) sign(id, ownerID uint64, exp int64) string {
	mac := hmac.New(sha256.New, s.secret)
	fmt.Fprintf(mac, "att:%d:%d:%d", id, ownerID, exp)
	return hex.EncodeToString(mac.Sum(nil))
}

// SignedURL 生成相对路径的短期访问链接；不含会话令牌，泄露后到期自动失效。
func (s *AttachmentService) SignedURL(att *model.Attachment) string {
	exp := (time.Now().Unix()/int64(attachmentURLBucket.Seconds()) + 2) * int64(attachmentURLBucket.Seconds())
	return fmt.Sprintf("/api/attachments/%d/file?exp=%d&sig=%s", att.ID, exp, s.sign(att.ID, att.UserID, exp))
}

// OpenSigned 校验签名后返回文件流。任何失败都统一为 ErrAttachmentNotFound，避免泄露附件是否存在。
func (s *AttachmentService) OpenSigned(ctx context.Context, id uint64, expStr, sig string) (io.ReadCloser, int64, string, error) {
	exp, err := strconv.ParseInt(expStr, 10, 64)
	if err != nil || exp < time.Now().Unix() || sig == "" {
		return nil, 0, "", ErrAttachmentNotFound
	}
	var att model.Attachment
	if err := s.db.Select("id", "user_id", "object_key", "content_type").First(&att, id).Error; err != nil {
		return nil, 0, "", ErrAttachmentNotFound
	}
	if !hmac.Equal([]byte(sig), []byte(s.sign(att.ID, att.UserID, exp))) {
		return nil, 0, "", ErrAttachmentNotFound
	}
	rc, size, err := s.minio.Open(ctx, att.ObjectKey)
	if err != nil {
		return nil, 0, "", ErrAttachmentNotFound
	}
	return rc, size, att.ContentType, nil
}

func (s *AttachmentService) RefreshURLs(list []model.Attachment) {
	if s == nil {
		return
	}
	for i := range list {
		list[i].URL = s.SignedURL(&list[i])
	}
}

func (s *AttachmentService) RefreshTxAttachments(t *model.Transaction) {
	if t == nil {
		return
	}
	s.RefreshURLs(t.Attachments)
}

func (s *AttachmentService) RefreshTxListAttachments(list []model.Transaction) {
	for i := range list {
		s.RefreshURLs(list[i].Attachments)
	}
}

func (s *AttachmentService) RefreshAccountAttachments(a *model.Account) {
	if a == nil {
		return
	}
	s.RefreshURLs(a.Attachments)
}

func (s *AttachmentService) RefreshAccountListAttachments(list []model.Account) {
	for i := range list {
		s.RefreshURLs(list[i].Attachments)
	}
}
