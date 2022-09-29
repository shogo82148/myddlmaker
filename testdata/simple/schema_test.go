package schema

import (
	"context"
	"database/sql"
	"errors"
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

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	if err := InsertFoo1(ctx, db, &Foo1{ID: 42}); err != nil {
		t.Errorf("failed to insert: %v", err)
	}

	if _, err := SelectFoo1(ctx, db, &Foo1{ID: 42}); err != nil {
		t.Errorf("failed to select: %v", err)
	}

	if _, err := SelectFoo1(ctx, db, &Foo1{ID: 43}); !errors.Is(err, sql.ErrNoRows) {
		t.Errorf("want sql.ErrNoRows, but got %v", err)
	}

	// multiple insert
	foo := []*Foo1{}
	for i := 0; i < 1000; i++ {
		foo = append(foo, &Foo1{ID: 1000 + int32(i)})
	}
	if err := InsertFoo1(ctx, db, foo...); err != nil {
		t.Errorf("failed to insert: %v", err)
	}

	all, err := SelectAllFoo1(ctx, db)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 1001 {
		t.Errorf("unexpected count: want %d, got %d", 1001, len(all))
	}
}
