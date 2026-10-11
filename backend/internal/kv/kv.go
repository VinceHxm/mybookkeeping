// Package kv 会话、登录限流、重置令牌等短期数据的存储；可用 Redis，也可直接存业务数据库（单容器部署）。
package kv

import (
	"context"
	"time"
)

// Store 只覆盖业务实际用到的少量操作；Get 不存在或已过期时 ok=false。
type Store interface {
	Get(ctx context.Context, key string) (val string, ok bool, err error)
	Set(ctx context.Context, key, val string, ttl time.Duration) error
	Del(ctx context.Context, keys ...string) error
	// Incr 计数 +1；键首次出现时设置 ttl，之后不续期（固定窗口）
	Incr(ctx context.Context, key string, ttl time.Duration) (int64, error)
	// TTL 剩余有效期；不存在或无过期时间时返回 0
	TTL(ctx context.Context, key string) (time.Duration, error)

	// 集合：同一 key 下多个成员，整体共享过期时间（SAdd 时刷新）
	SAdd(ctx context.Context, key, member string, ttl time.Duration) error
	SRem(ctx context.Context, key string, members ...string) error
	SMembers(ctx context.Context, key string) ([]string, error)
}
