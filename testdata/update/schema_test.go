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

	u1 := &User{
		ID:   42,
		Name: "chooblarin",
	}
	if err := InsertUser(ctx, db, u1); err != nil {
		t.Errorf("failed to insert: %v", err)
	}

	u2 := &User{
		ID:   43,
		Name: "nekosushi",
	}
	if err := InsertUser(ctx, db, u2); err != nil {
		t.Errorf("failed to insert: %v", err)
	}

	// multiple update
	u1 = &User{
		ID:   1,
		Name: "CHOOBLARIN",
	}
	u2 = &User{
		ID:   2,
		Name: "NEKOSUSHI",
	}
	if err := UpdateUser(ctx, db, u1, u2); err != nil {
		t.Errorf("failed to update: %v", err)
	}

	u, err := SelectUser(ctx, db, &User{ID: 1})
	if err != nil {
		t.Errorf("failed to select: %v", err)
	}
	if u.ID != 1 {
		t.Errorf("unexpected id: want 42, got %d", u.ID)
	}
	if u.Name != "CHOOBLARIN" {
		t.Errorf("unexpected name: want CHOOBLARIN, got %s", u.Name)
	}
}
