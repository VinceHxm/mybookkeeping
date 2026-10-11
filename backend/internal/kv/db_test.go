package kv

import (
	"context"
	"fmt"
	"sort"
	"strings"
	"testing"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

func newTestStore(t *testing.T) *DBStore {
	t.Helper()
	dsn := fmt.Sprintf("file:%s?mode=memory&cache=shared", strings.ReplaceAll(t.Name(), "/", "_"))
	db, err := gorm.Open(sqlite.Open(dsn), &gorm.Config{Logger: logger.Discard})
	if err != nil {
		t.Fatal(err)
	}
	s, err := NewDBStore(db)
	if err != nil {
		t.Fatal(err)
	}
	return s
}

func TestDBStoreGetSetDel(t *testing.T) {
	s, ctx := newTestStore(t), context.Background()
	if _, ok, err := s.Get(ctx, "a"); err != nil || ok {
		t.Fatalf("missing key: ok=%v err=%v", ok, err)
	}
	if err := s.Set(ctx, "a", "1", time.Hour); err != nil {
		t.Fatal(err)
	}
	if err := s.Set(ctx, "a", "2", time.Hour); err != nil {
		t.Fatal(err)
	}
	if v, ok, _ := s.Get(ctx, "a"); !ok || v != "2" {
		t.Fatalf("overwrite: got %q ok=%v", v, ok)
	}
	if d, _ := s.TTL(ctx, "a"); d <= 59*time.Minute || d > time.Hour {
		t.Fatalf("ttl = %v", d)
	}
	_ = s.Del(ctx, "a")
	if _, ok, _ := s.Get(ctx, "a"); ok {
		t.Fatal("deleted key still readable")
	}
}

func TestDBStoreExpiry(t *testing.T) {
	s, ctx := newTestStore(t), context.Background()
	_ = s.Set(ctx, "gone", "x", time.Millisecond)
	_ = s.Set(ctx, "keep", "y", 0)
	time.Sleep(20 * time.Millisecond)
	if _, ok, _ := s.Get(ctx, "gone"); ok {
		t.Fatal("expired key readable")
	}
	if v, ok, _ := s.Get(ctx, "keep"); !ok || v != "y" {
		t.Fatal("non-expiring key lost")
	}
	if n, err := s.PurgeExpired(ctx); err != nil || n != 1 {
		t.Fatalf("purge n=%d err=%v", n, err)
	}
}

func TestDBStoreIncrFixedWindow(t *testing.T) {
	s, ctx := newTestStore(t), context.Background()
	for i := int64(1); i <= 3; i++ {
		n, err := s.Incr(ctx, "cnt", time.Hour)
		if err != nil || n != i {
			t.Fatalf("incr #%d = %d err=%v", i, n, err)
		}
	}
	_ = s.Set(ctx, "old", "9", time.Millisecond)
	time.Sleep(20 * time.Millisecond)
	if n, _ := s.Incr(ctx, "old", time.Hour); n != 1 {
		t.Fatalf("expired counter should restart at 1, got %d", n)
	}
}

func TestDBStoreSet(t *testing.T) {
	s, ctx := newTestStore(t), context.Background()
	_ = s.SAdd(ctx, "usess:1", "t1", time.Hour)
	_ = s.SAdd(ctx, "usess:1", "t2", time.Hour)
	_ = s.SAdd(ctx, "usess:1", "t2", time.Hour)
	_ = s.Set(ctx, "usess:1x", "scalar", time.Hour)
	got, err := s.SMembers(ctx, "usess:1")
	sort.Strings(got)
	if err != nil || strings.Join(got, ",") != "t1,t2" {
		t.Fatalf("members = %v err=%v", got, err)
	}
	_ = s.SRem(ctx, "usess:1", "t1")
	if got, _ := s.SMembers(ctx, "usess:1"); len(got) != 1 || got[0] != "t2" {
		t.Fatalf("after srem = %v", got)
	}
	_ = s.Del(ctx, "usess:1")
	if got, _ := s.SMembers(ctx, "usess:1"); len(got) != 0 {
		t.Fatalf("after del = %v", got)
	}
	if _, ok, _ := s.Get(ctx, "usess:1x"); !ok {
		t.Fatal("unrelated key removed")
	}
}
