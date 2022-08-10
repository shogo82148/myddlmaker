package schema

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
)

func TestFoo1(t *testing.T) {
	user := os.Getenv("MYSQL_TEST_USER")
	pass := os.Getenv("MYSQL_TEST_PASS")
	addr := os.Getenv("MYSQL_TEST_ADDR")
	name := os.Getenv("MYSQL_TEST_DB")
	if name == "" {
		return
	}
	cfg := mysql.NewConfig()
	cfg.User = user
	cfg.Passwd = pass
	cfg.Addr = addr
	cfg.DBName = name
	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	if err := InsertFoo1(ctx, db, &Foo1{ID: 42}); err != nil {
		t.Errorf("failed to insert: %v", err)
	}

	if _, err := SelectFoo1(ctx, db, &Foo1{ID: 42}); err != nil {
		t.Errorf("failed to select: %v", err)
	}

	if _, err := SelectFoo1(ctx, db, &Foo1{ID: 43}); err != nil {
		t.Errorf("want sql.ErrNoRows, but got %v", err)
	}
}
