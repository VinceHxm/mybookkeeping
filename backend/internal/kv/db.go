package kv

import (
	"context"
	"errors"
	"strconv"
	"time"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

// Entry 普通键 M 为空串；集合成员各占一行，M 为成员值
type Entry struct {
	K         string     `gorm:"column:k;primaryKey;size:191"`
	M         string     `gorm:"column:m;primaryKey;size:191;default:''"`
	V         string     `gorm:"column:v;type:text"`
	ExpiresAt *time.Time `gorm:"column:expires_at;index"`
}

func (Entry) TableName() string { return "kv_entries" }

// DBStore 用业务数据库存短期数据，免去 Redis；过期数据读取时忽略，由 PurgeExpired 定期清理。
type DBStore struct {
	db *gorm.DB
}

func NewDBStore(db *gorm.DB) (*DBStore, error) {
	if err := db.AutoMigrate(&Entry{}); err != nil {
		return nil, err
	}
	return &DBStore{db: db}, nil
}

func expiry(now time.Time, ttl time.Duration) *time.Time {
	if ttl <= 0 {
		return nil
	}
	t := now.Add(ttl)
	return &t
}

func (s *DBStore) alive(ctx context.Context) *gorm.DB {
	return s.db.WithContext(ctx).Model(&Entry{}).Where("(expires_at IS NULL OR expires_at > ?)", time.Now())
}

func (s *DBStore) Get(ctx context.Context, key string) (string, bool, error) {
	var e Entry
	err := s.alive(ctx).Where("k = ? AND m = ''", key).Take(&e).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return e.V, true, nil
}

func (s *DBStore) upsert(db *gorm.DB, e *Entry) error {
	return db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "k"}, {Name: "m"}},
		DoUpdates: clause.AssignmentColumns([]string{"v", "expires_at"}),
	}).Create(e).Error
}

func (s *DBStore) Set(ctx context.Context, key, val string, ttl time.Duration) error {
	return s.upsert(s.db.WithContext(ctx), &Entry{K: key, V: val, ExpiresAt: expiry(time.Now(), ttl)})
}

func (s *DBStore) Del(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	return s.db.WithContext(ctx).Where("k IN ?", keys).Delete(&Entry{}).Error
}

func (s *DBStore) Incr(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	var n int64
	err := s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		now := time.Now()
		var e Entry
		err := tx.Clauses(clause.Locking{Strength: "UPDATE"}).Where("k = ? AND m = ''", key).Take(&e).Error
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			return err
		}
		if err != nil || (e.ExpiresAt != nil && !e.ExpiresAt.After(now)) {
			n = 1
			return s.upsert(tx, &Entry{K: key, V: "1", ExpiresAt: expiry(now, ttl)})
		}
		cur, _ := strconv.ParseInt(e.V, 10, 64)
		n = cur + 1
		return tx.Model(&Entry{}).Where("k = ? AND m = ''", key).Update("v", strconv.FormatInt(n, 10)).Error
	})
	return n, err
}

func (s *DBStore) TTL(ctx context.Context, key string) (time.Duration, error) {
	var e Entry
	err := s.alive(ctx).Where("k = ? AND m = ''", key).Take(&e).Error
	if errors.Is(err, gorm.ErrRecordNotFound) {
		return 0, nil
	}
	if err != nil || e.ExpiresAt == nil {
		return 0, err
	}
	if d := time.Until(*e.ExpiresAt); d > 0 {
		return d, nil
	}
	return 0, nil
}

func (s *DBStore) SAdd(ctx context.Context, key, member string, ttl time.Duration) error {
	return s.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		exp := expiry(time.Now(), ttl)
		if err := s.upsert(tx, &Entry{K: key, M: member, ExpiresAt: exp}); err != nil {
			return err
		}
		return tx.Model(&Entry{}).Where("k = ? AND m <> ''", key).Update("expires_at", exp).Error
	})
}

func (s *DBStore) SRem(ctx context.Context, key string, members ...string) error {
	if len(members) == 0 {
		return nil
	}
	return s.db.WithContext(ctx).Where("k = ? AND m IN ?", key, members).Delete(&Entry{}).Error
}

func (s *DBStore) SMembers(ctx context.Context, key string) ([]string, error) {
	var out []string
	err := s.alive(ctx).Where("k = ? AND m <> ''", key).Pluck("m", &out).Error
	return out, err
}

// PurgeExpired 删除已过期的数据，返回删除行数
func (s *DBStore) PurgeExpired(ctx context.Context) (int64, error) {
	res := s.db.WithContext(ctx).Where("expires_at IS NOT NULL AND expires_at <= ?", time.Now()).Delete(&Entry{})
	return res.RowsAffected, res.Error
}
