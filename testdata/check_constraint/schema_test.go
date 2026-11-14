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

func TestUser(t *testing.T) {
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

	// valid values
	if err := InsertUser(ctx, db, &User{
		Name:     "Alice",
		Age:      20,
		MinScore: 10,
		MaxScore: 100,
	}); err != nil {
		t.Fatalf("failed to insert: %v", err)
	}

	// NOT ENFORCED constraints are not checked.
	if err := InsertUser(ctx, db, &User{
		Name:     "Bob",
		Age:      30,
		MinScore: 10,
		MaxScore: 100,
		Memo:     "forbidden",
	}); err != nil {
		t.Fatalf("failed to insert: %v", err)
	}

	// invalid values
	invalid := []*User{
		{Name: "", Age: 20, MinScore: 10, MaxScore: 100},
		{Name: "Carol", Age: -1, MinScore: 10, MaxScore: 100},
		{Name: "Carol", Age: 151, MinScore: 10, MaxScore: 100},
		{Name: "Carol", Age: 20, MinScore: 100, MaxScore: 10},
	}
	for _, u := range invalid {
		err := InsertUser(ctx, db, u)
		assertCheckConstraintViolated(t, err)
	}

	// update with invalid values
	err = UpdateUser(ctx, db, &User{ID: 1, Name: "Alice", Age: 200, MinScore: 10, MaxScore: 100})
	assertCheckConstraintViolated(t, err)

	users, err := SelectAllUser(ctx, db)
	if err != nil {
		t.Fatalf("failed to select: %v", err)
	}
	if len(users) != 2 {
		t.Fatalf("unexpected length: want 2, got %d", len(users))
	}
	if users[0].Age != 20 {
		t.Errorf("unexpected age: want 20, got %d", users[0].Age)
	}
}

func assertCheckConstraintViolated(t *testing.T, err error) {
	t.Helper()
	var myErr *mysql.MySQLError
	if !errors.As(err, &myErr) {
		t.Errorf("want MySQLError, got %T: %v", err, err)
		return
	}
	// ER_CHECK_CONSTRAINT_VIOLATED
	// https://dev.mysql.com/doc/mysql-errors/8.0/en/server-error-reference.html#error_er_check_constraint_violated
	if myErr.Number != 3819 {
		t.Errorf("unexpected error number: want 3819, got %d: %v", myErr.Number, myErr)
	}
}
