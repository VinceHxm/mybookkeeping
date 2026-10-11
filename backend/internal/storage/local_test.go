package storage

import (
	"context"
	"io"
	"strings"
	"testing"
)

func TestLocalUploadOpen(t *testing.T) {
	l, err := NewLocal(t.TempDir())
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	if err := l.Upload(ctx, "u1/a.jpg", "image/jpeg", strings.NewReader("hello"), 5); err != nil {
		t.Fatal(err)
	}
	rc, size, err := l.Open(ctx, "u1/a.jpg")
	if err != nil {
		t.Fatal(err)
	}
	defer rc.Close()
	b, _ := io.ReadAll(rc)
	if size != 5 || string(b) != "hello" {
		t.Fatalf("got %q size=%d", b, size)
	}
	if err := l.Upload(ctx, "u1/b.jpg", "image/jpeg", strings.NewReader("abc"), 5); err == nil {
		t.Fatal("size mismatch should fail")
	}
}

func TestLocalRejectsTraversal(t *testing.T) {
	l, _ := NewLocal(t.TempDir())
	for _, key := range []string{"../x", "u1/../../x", "", "/etc/passwd"} {
		if _, _, err := l.Open(context.Background(), key); err == nil {
			t.Fatalf("key %q should be rejected", key)
		}
	}
}
