package kv

import (
	"context"
	"errors"
	"time"

	"github.com/redis/go-redis/v9"
)

type RedisStore struct {
	rdb *redis.Client
}

func NewRedisStore(rdb *redis.Client) *RedisStore {
	return &RedisStore{rdb: rdb}
}

func (s *RedisStore) Get(ctx context.Context, key string) (string, bool, error) {
	v, err := s.rdb.Get(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return "", false, nil
	}
	if err != nil {
		return "", false, err
	}
	return v, true, nil
}

func (s *RedisStore) Set(ctx context.Context, key, val string, ttl time.Duration) error {
	return s.rdb.Set(ctx, key, val, ttl).Err()
}

func (s *RedisStore) Del(ctx context.Context, keys ...string) error {
	if len(keys) == 0 {
		return nil
	}
	return s.rdb.Del(ctx, keys...).Err()
}

func (s *RedisStore) Incr(ctx context.Context, key string, ttl time.Duration) (int64, error) {
	n, err := s.rdb.Incr(ctx, key).Result()
	if err != nil {
		return 0, err
	}
	if n == 1 && ttl > 0 {
		_ = s.rdb.Expire(ctx, key, ttl).Err()
	}
	return n, nil
}

func (s *RedisStore) TTL(ctx context.Context, key string) (time.Duration, error) {
	d, err := s.rdb.TTL(ctx, key).Result()
	if err != nil || d < 0 {
		return 0, err
	}
	return d, nil
}

func (s *RedisStore) SAdd(ctx context.Context, key, member string, ttl time.Duration) error {
	pipe := s.rdb.Pipeline()
	pipe.SAdd(ctx, key, member)
	if ttl > 0 {
		pipe.Expire(ctx, key, ttl)
	}
	_, err := pipe.Exec(ctx)
	return err
}

func (s *RedisStore) SRem(ctx context.Context, key string, members ...string) error {
	if len(members) == 0 {
		return nil
	}
	args := make([]any, len(members))
	for i, m := range members {
		args[i] = m
	}
	return s.rdb.SRem(ctx, key, args...).Err()
}

func (s *RedisStore) SMembers(ctx context.Context, key string) ([]string, error) {
	v, err := s.rdb.SMembers(ctx, key).Result()
	if errors.Is(err, redis.Nil) {
		return nil, nil
	}
	return v, err
}
