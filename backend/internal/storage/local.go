package storage

import (
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Local 附件存本地目录，单容器部署时挂载数据卷即可持久化
type Local struct {
	root string
}

func NewLocal(root string) (*Local, error) {
	abs, err := filepath.Abs(root)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(abs, 0o755); err != nil {
		return nil, fmt.Errorf("local storage dir: %w", err)
	}
	return &Local{root: abs}, nil
}

var errBadObjectKey = errors.New("invalid object key")

func (l *Local) path(objectKey string) (string, error) {
	if objectKey == "" || strings.Contains(objectKey, "..") || filepath.IsAbs(objectKey) {
		return "", errBadObjectKey
	}
	p := filepath.Join(l.root, filepath.FromSlash(objectKey))
	if !strings.HasPrefix(p, l.root+string(filepath.Separator)) {
		return "", errBadObjectKey
	}
	return p, nil
}

func (l *Local) Upload(_ context.Context, objectKey, _ string, reader io.Reader, size int64) error {
	p, err := l.path(objectKey)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(p), 0o755); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(p), ".upload-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name())
	n, err := io.Copy(tmp, io.LimitReader(reader, size+1))
	if cerr := tmp.Close(); err == nil {
		err = cerr
	}
	if err != nil {
		return err
	}
	if n != size {
		return fmt.Errorf("upload size mismatch: got %d, want %d", n, size)
	}
	return os.Rename(tmp.Name(), p)
}

func (l *Local) Open(_ context.Context, objectKey string) (io.ReadCloser, int64, error) {
	p, err := l.path(objectKey)
	if err != nil {
		return nil, 0, err
	}
	f, err := os.Open(p)
	if err != nil {
		return nil, 0, err
	}
	st, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return nil, 0, err
	}
	return f, st.Size(), nil
}
