package myddlmaker

import (
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"fmt"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/go-sql-driver/mysql"
	"github.com/google/go-cmp/cmp"
)

type Foo1 struct {
	ID int32
}

func (*Foo1) PrimaryKey() *PrimaryKey {
	return NewPrimaryKey("id")
}

type Foo2 struct {
	ID   int32
	Name string
}

func (*Foo2) PrimaryKey() *PrimaryKey {
	return NewPrimaryKey("id")
}

func (*Foo2) Indexes() []*Index {
	return []*Index{
		NewIndex("idx_name", "name"),
	}
}

type Foo3 struct {
	ID   int32
	Name string
}

func (*Foo3) PrimaryKey() *PrimaryKey {
	return NewPrimaryKey("id")
}

func (*Foo3) UniqueIndexes() []*UniqueIndex {
	return []*UniqueIndex{
		NewUniqueIndex("idx_name", "name"),
	}
}

func testMaker(t *testing.T, structs []any, ddl string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	m, err := New(&Config{})
	if err != nil {
		t.Fatalf("failed to initialize Maker: %v", err)
	}

	m.AddStructs(structs...)

	var buf bytes.Buffer
	if err := m.Generate(&buf); err != nil {
		t.Fatalf("failed to generate ddl: %v", err)
	}

	got := buf.String()
	if diff := cmp.Diff(ddl, got); diff != "" {
		t.Errorf("ddl is not match: (-want/+got)\n%s", diff)
	}

	// check the ddl syntax with MySQL Server
	user := os.Getenv("MYSQL_TEST_USER")
	pass := os.Getenv("MYSQL_TEST_PASS")
	addr := os.Getenv("MYSQL_TEST_ADDR")
	if user == "" || pass == "" || addr == "" {
		return
	}

	// connect to the server
	cfg := mysql.NewConfig()
	cfg.User = user
	cfg.Passwd = pass
	cfg.Addr = addr
	db0, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db0.Close()

	// create a new database
	var buf2 [4]byte
	_, err = rand.Read(buf2[:])
	if err != nil {
		t.Fatal(err)
	}
	dbName := fmt.Sprintf("myddlmaker_%x", buf2[:])
	_, err = db0.ExecContext(ctx, "CREATE DATABASE "+dbName)
	if err != nil {
		t.Fatalf("failed to create database %q: %v", dbName, err)
	}
	defer db0.ExecContext(ctx, "DROP DATABASE"+dbName)

	cfg.DBName = dbName
	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	conn, err := db.Conn(ctx)
	if err != nil {
		t.Fatalf("failed to get a connection: %v", err)
	}
	defer conn.Close()

	// check the ddl syntax
	lines := strings.Split(got, ";\n")
	for _, q := range lines {
		q := strings.TrimSpace(q)
		if q == "" {
			continue
		}
		_, err := conn.ExecContext(ctx, q)
		if err != nil {
			t.Errorf("failed to execute %q: %v", q, err)
		}
	}
}

func TestMaker(t *testing.T) {
	testMaker(t, []any{&Foo1{}}, "SET foreign_key_checks=0;\n"+
		"DROP TABLE IF EXISTS `foo1`;\n\n"+
		"CREATE TABLE `foo1` (\n"+
		"    `id` INTEGER NOT NULL,\n"+
		"    PRIMARY KEY (`id`)\n"+
		") ENGINE=InnoDB DEFAULT CHARACTER SET = 'utf8mb4';\n\n"+
		"SET foreign_key_checks=1;\n")

	testMaker(t, []any{&Foo2{}}, "SET foreign_key_checks=0;\n"+
		"DROP TABLE IF EXISTS `foo2`;\n\n"+
		"CREATE TABLE `foo2` (\n"+
		"    `id` INTEGER NOT NULL,\n"+
		"    `name` VARCHAR(191) NOT NULL,\n"+
		"    INDEX `idx_name` (`name`),\n"+
		"    PRIMARY KEY (`id`)\n"+
		") ENGINE=InnoDB DEFAULT CHARACTER SET = 'utf8mb4';\n\n"+
		"SET foreign_key_checks=1;\n")

	testMaker(t, []any{&Foo3{}}, "SET foreign_key_checks=0;\n"+
		"DROP TABLE IF EXISTS `foo3`;\n\n"+
		"CREATE TABLE `foo3` (\n"+
		"    `id` INTEGER NOT NULL,\n"+
		"    `name` VARCHAR(191) NOT NULL,\n"+
		"    UNIQUE `idx_name` (`name`),\n"+
		"    PRIMARY KEY (`id`)\n"+
		") ENGINE=InnoDB DEFAULT CHARACTER SET = 'utf8mb4';\n\n"+
		"SET foreign_key_checks=1;\n")
}
