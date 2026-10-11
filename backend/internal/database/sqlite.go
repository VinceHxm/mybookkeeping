package database

import (
	"context"
	"database/sql"
	"database/sql/driver"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/glebarez/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

const sqliteLocalDriver = "mbk-sqlite"

// WAL：读不阻塞写；immediate：写事务开头即拿写锁，避免两个事务同时"读后升级写"互相等待到超时
const sqliteDSNParams = "?_pragma=busy_timeout(10000)&_pragma=journal_mode(WAL)&_pragma=synchronous(NORMAL)&_txlock=immediate"

var registerSQLiteOnce sync.Once

// OpenSQLite 打开（不存在则创建）本地 SQLite 文件并完成建表迁移
func OpenSQLite(path string) (*gorm.DB, error) {
	if dir := filepath.Dir(path); dir != "" {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			return nil, fmt.Errorf("sqlite data dir: %w", err)
		}
	}
	registerSQLiteLocalDriver()
	db, err := gorm.Open(sqlite.Dialector{DriverName: sqliteLocalDriver, DSN: path + sqliteDSNParams}, &gorm.Config{
		Logger: logger.Default.LogMode(logger.Warn),
	})
	if err != nil {
		return nil, fmt.Errorf("sqlite open: %w", err)
	}
	sqlDB, err := db.DB()
	if err != nil {
		return nil, err
	}
	sqlDB.SetMaxOpenConns(8)
	sqlDB.SetMaxIdleConns(8)

	if err := migrate(db); err != nil {
		return nil, err
	}
	return db, nil
}

// SQLite 以文本存时间、按字符串比较范围；参数时区不一致（如前端传 UTC 的 Z 结尾时间）会导致范围查询错位，
// 因此所有 time.Time 参数统一换算到本地时区后再交给驱动。
func registerSQLiteLocalDriver() {
	registerSQLiteOnce.Do(func() {
		probe, err := sql.Open(sqlite.DriverName, ":memory:")
		if err != nil {
			panic(err)
		}
		inner := probe.Driver()
		_ = probe.Close()
		sql.Register(sqliteLocalDriver, localTimeDriver{inner})
	})
}

type localTimeDriver struct{ driver.Driver }

func (d localTimeDriver) Open(name string) (driver.Conn, error) {
	c, err := d.Driver.Open(name)
	if err != nil {
		return nil, err
	}
	return &localTimeConn{Conn: c}, nil
}

type localTimeConn struct{ driver.Conn }

func (c *localTimeConn) CheckNamedValue(nv *driver.NamedValue) error {
	v, err := driver.DefaultParameterConverter.ConvertValue(nv.Value)
	if err != nil {
		return err
	}
	if t, ok := v.(time.Time); ok {
		v = t.In(time.Local)
	}
	nv.Value = v
	return nil
}

func (c *localTimeConn) Ping(ctx context.Context) error {
	if p, ok := c.Conn.(driver.Pinger); ok {
		return p.Ping(ctx)
	}
	return nil
}

func (c *localTimeConn) BeginTx(ctx context.Context, opts driver.TxOptions) (driver.Tx, error) {
	if b, ok := c.Conn.(driver.ConnBeginTx); ok {
		return b.BeginTx(ctx, opts)
	}
	return c.Conn.Begin() //nolint:staticcheck
}

func (c *localTimeConn) PrepareContext(ctx context.Context, query string) (driver.Stmt, error) {
	if p, ok := c.Conn.(driver.ConnPrepareContext); ok {
		return p.PrepareContext(ctx, query)
	}
	return c.Conn.Prepare(query)
}

func (c *localTimeConn) ExecContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Result, error) {
	if e, ok := c.Conn.(driver.ExecerContext); ok {
		return e.ExecContext(ctx, query, args)
	}
	return nil, driver.ErrSkip
}

func (c *localTimeConn) QueryContext(ctx context.Context, query string, args []driver.NamedValue) (driver.Rows, error) {
	if q, ok := c.Conn.(driver.QueryerContext); ok {
		return q.QueryContext(ctx, query, args)
	}
	return nil, driver.ErrSkip
}
