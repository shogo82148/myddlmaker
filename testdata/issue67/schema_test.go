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

	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()

	u1 := &User{}
	if err := InsertUser(ctx, db, u1); err != nil {
		t.Errorf("failed to insert: %v", err)
	}

	all, err := SelectAllUser(ctx, db)
	if err != nil {
		t.Fatal(err)
	}
	if len(all) != 1 {
		t.Errorf("want 1, but got %d", len(all))
	}
	if all[0].ID != 1 {
		t.Errorf("want 1, but got %d", all[0].ID)
	}
}
