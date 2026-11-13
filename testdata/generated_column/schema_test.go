package schema

import (
	"context"
	"database/sql"
	"os"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
)

func TestItem(t *testing.T) {
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

	// the values of generated columns are ignored.
	if err := InsertItem(ctx, db, &Item{
		Price:   100,
		Count:   3,
		Total:   -1,
		Summary: "ignored",
	}); err != nil {
		t.Fatalf("failed to insert: %v", err)
	}

	item, err := SelectItem(ctx, db, &Item{ID: 1})
	if err != nil {
		t.Fatalf("failed to select: %v", err)
	}
	if item.Total != 300 {
		t.Errorf("unexpected total: want 300, got %d", item.Total)
	}
	if item.Summary != "100 x 3" {
		t.Errorf("unexpected summary: want %q, got %q", "100 x 3", item.Summary)
	}

	// update
	if err := UpdateItem(ctx, db, &Item{
		ID:      1,
		Price:   200,
		Count:   5,
		Total:   -1,
		Summary: "ignored",
	}); err != nil {
		t.Fatalf("failed to update: %v", err)
	}

	items, err := SelectAllItem(ctx, db)
	if err != nil {
		t.Fatalf("failed to select: %v", err)
	}
	if len(items) != 1 {
		t.Fatalf("unexpected length: want 1, got %d", len(items))
	}
	if items[0].Total != 1000 {
		t.Errorf("unexpected total: want 1000, got %d", items[0].Total)
	}
	if items[0].Summary != "200 x 5" {
		t.Errorf("unexpected summary: want %q, got %q", "200 x 5", items[0].Summary)
	}
}
