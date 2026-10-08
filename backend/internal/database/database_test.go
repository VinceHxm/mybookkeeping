package database

import (
	"testing"
	"time"

	mysqldriver "github.com/go-sql-driver/mysql"
)

func TestWithTimeouts(t *testing.T) {
	c, err := mysqldriver.ParseDSN(withTimeouts("u:p@tcp(127.0.0.1:3306)/db?parseTime=true&loc=Local"))
	if err != nil {
		t.Fatal(err)
	}
	if c.Timeout != mysqlDialTimeout || c.ReadTimeout != mysqlIOTimeout || c.WriteTimeout != mysqlIOTimeout {
		t.Fatalf("defaults not applied: %+v", c)
	}
	if !c.ParseTime || c.DBName != "db" || c.Passwd != "p" {
		t.Fatalf("original params lost: %+v", c)
	}

	c, _ = mysqldriver.ParseDSN(withTimeouts("u:p@tcp(h:1)/db?timeout=2s&readTimeout=7s"))
	if c.Timeout != 2*time.Second || c.ReadTimeout != 7*time.Second || c.WriteTimeout != mysqlIOTimeout {
		t.Fatalf("explicit values overridden: %+v", c)
	}

	if got := withTimeouts("not a dsn"); got != "not a dsn" {
		t.Fatalf("invalid dsn changed: %q", got)
	}
}
