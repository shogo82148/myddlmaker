package myddlmaker

import (
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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
	ID   int32  `ddl:",auto"`
	Name string `ddl:",comment='コメント',invisible"`
}

func (*Foo2) Table() string {
	return "foo2_customized"
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

type Foo4 struct {
	ID   int32
	Name string
}

func (*Foo4) PrimaryKey() *PrimaryKey {
	return NewPrimaryKey("id")
}

func (*Foo4) ForeignKeys() []*ForeignKey {
	return []*ForeignKey{
		NewForeignKey("fk_foo1", []string{"id"}, "foo1", []string{"id"}),
	}
}

type Foo5 struct {
	ID   int32
	Name string
}

func (*Foo5) PrimaryKey() *PrimaryKey {
	return NewPrimaryKey("id")
}

func (*Foo5) ForeignKeys() []*ForeignKey {
	return []*ForeignKey{
		NewForeignKey("fk_foo1", []string{"id"}, "foo1", []string{"id"}).OnUpdate(ForeignKeyOptionCascade).OnDelete(ForeignKeyOptionCascade),
	}
}

type Foo6 struct {
	ID    int32
	Name  string
	Email string
}

func (*Foo6) PrimaryKey() *PrimaryKey {
	return NewPrimaryKey("id")
}

func (*Foo6) Indexes() []*Index {
	return []*Index{
		// Indexes with comments.
		NewIndex("idx_name", "name").Comment("an index\n\twith 'comment'"),
	}
}

func (*Foo6) UniqueIndexes() []*UniqueIndex {
	return []*UniqueIndex{
		// Indexes with comments.
		NewUniqueIndex("uniq_email", "email").Comment("a unique index\n\twith 'comment'"),
	}
}

type Foo7 struct {
	ID   int32
	Name string `ddl:",default='John Doe'"`
}

func (*Foo7) PrimaryKey() *PrimaryKey {
	return NewPrimaryKey("id")
}

type Foo8 struct {
	ID   int32 `ddl:",auto"`
	Name string
}

func (*Foo8) PrimaryKey() *PrimaryKey {
	return NewPrimaryKey("id")
}

func (*Foo8) Indexes() []*Index {
	return []*Index{
		NewIndex("idx_name", "name").Invisible(),
	}
}

type Foo9 struct {
	ID   int32  `ddl:",auto"`
	Name string `ddl:",charset=utf8,collate=utf8_bin"`
}

func (*Foo9) PrimaryKey() *PrimaryKey {
	return NewPrimaryKey("id")
}

type Foo10 struct {
	ID   int32 `ddl:",auto"`
	Text string
}

func (*Foo10) PrimaryKey() *PrimaryKey {
	return NewPrimaryKey("id")
}

func (*Foo10) FullTextIndexes() []*FullTextIndex {
	return []*FullTextIndex{
		NewFullTextIndex("idx_text", "text").WithParser("ngram").Comment("FULLTEXT INDEX"),
	}
}

type Foo11 struct {
	ID    int32  `ddl:",auto"`
	Point string `ddl:",type=GEOMETRY,srid=4326"`
}

func (*Foo11) PrimaryKey() *PrimaryKey {
	return NewPrimaryKey("id")
}

func (*Foo11) SpatialIndexes() []*SpatialIndex {
	return []*SpatialIndex{
		NewSpatialIndex("idx_point", "point").Comment("SPATIAL INDEX"),
	}
}

type Foo12 struct {
	ID int32
}

func (*Foo12) Table() string {
	return "foo11"
}

func (*Foo12) PrimaryKey() *PrimaryKey {
	return NewPrimaryKey("id")
}

type Foo13 struct {
	ID               int32 `ddl:"id"`
	DuplicatedColumn int32 `ddl:"id"`
}

func (*Foo13) PrimaryKey() *PrimaryKey {
	return NewPrimaryKey("id")
}

type Foo14 struct {
	ID int32
}

func (*Foo14) PrimaryKey() *PrimaryKey {
	return NewPrimaryKey("unknown_column")
}

func (*Foo14) Indexes() []*Index {
	return []*Index{
		NewIndex("idx", "unknown_column"),
	}
}

func (*Foo14) UniqueIndexes() []*UniqueIndex {
	return []*UniqueIndex{
		NewUniqueIndex("uniq", "unknown_column"),
	}
}

type Foo15 struct {
	ID   int32
	Name string
}

func (*Foo15) PrimaryKey() *PrimaryKey {
	return NewPrimaryKey("id")
}

func (*Foo15) Indexes() []*Index {
	return []*Index{
		NewIndex("idx_name", "name"),
	}
}

func (*Foo15) UniqueIndexes() []*UniqueIndex {
	return []*UniqueIndex{
		NewUniqueIndex("idx_name", "name"),
	}
}

func (*Foo15) FullTextIndexes() []*FullTextIndex {
	return []*FullTextIndex{
		NewFullTextIndex("idx_name", "name").WithParser("ngram").Comment("FULLTEXT INDEX"),
	}
}

func (*Foo15) SpatialIndexes() []*SpatialIndex {
	return []*SpatialIndex{
		NewSpatialIndex("idx_name", "name").Comment("SPATIAL INDEX"),
	}
}

type Foo16 struct {
	ID   int32
	Name string
}

func (*Foo16) PrimaryKey() *PrimaryKey {
	return NewPrimaryKey("id")
}

func (*Foo16) ForeignKeys() []*ForeignKey {
	return []*ForeignKey{
		NewForeignKey("fk_duplicated", []string{"id"}, "foo16", []string{"id"}),
		NewForeignKey("fk_duplicated", []string{"id"}, "foo16", []string{"id"}),
	}
}

type Foo17 struct {
	ID int32
}

func (*Foo17) PrimaryKey() *PrimaryKey {
	return NewPrimaryKey("id")
}

func (*Foo17) ForeignKeys() []*ForeignKey {
	return []*ForeignKey{
		NewForeignKey("fk_foo17", []string{"unknown_column"}, "unknown_table", []string{"id"}),
	}
}

type Foo18 struct {
	ID      int32
	Foo19ID int32
}

func (*Foo18) PrimaryKey() *PrimaryKey {
	return NewPrimaryKey("id")
}

func (*Foo18) ForeignKeys() []*ForeignKey {
	return []*ForeignKey{
		// Foo18.Foo19ID is int32, but Foo19.ID is int64
		// it causes a type error
		NewForeignKey("fk_foo19", []string{"foo19_id"}, "foo19", []string{"id"}),
	}
}

type Foo19 struct {
	ID int64
}

func (*Foo19) PrimaryKey() *PrimaryKey {
	return NewPrimaryKey("id")
}

type Foo20 struct {
	ID   int32 `ddl:",auto"`
	JSON JSON[struct {
		A string `json:"a"`
		B int    `json:"b"`
	}]
}

func (*Foo20) PrimaryKey() *PrimaryKey {
	return NewPrimaryKey("id")
}

type Foo21 struct {
	ID       int32
	ParentID nullUint32 `ddl:",type=INTEGER,unsigned,null"`
}

func (*Foo21) PrimaryKey() *PrimaryKey {
	return NewPrimaryKey("id")
}

type Fkp1 struct {
	ID string
}

func (*Fkp1) PrimaryKey() *PrimaryKey {
	return NewPrimaryKey("id")
}

type Fkc1 struct {
	ID       string
	ParentID sql.NullString `ddl:",null"`
}

func (*Fkc1) PrimaryKey() *PrimaryKey {
	return NewPrimaryKey("id")
}

func (*Fkc1) Indexes() []*Index {
	return []*Index{
		NewIndex("idx_parent_id", "parent_id"),
	}
}

func (*Fkc1) ForeignKeys() []*ForeignKey {
	return []*ForeignKey{
		NewForeignKey(
			"fk_fkc1_parent_id",
			[]string{"parent_id"},
			"fkp1",
			[]string{"id"},
		),
	}
}

type Fkp2 struct {
	ID int64
}

func (*Fkp2) PrimaryKey() *PrimaryKey {
	return NewPrimaryKey("id")
}

type Fkc2 struct {
	ID       string
	ParentID sql.NullInt64 `ddl:",null"`
}

func (*Fkc2) PrimaryKey() *PrimaryKey {
	return NewPrimaryKey("id")
}

func (*Fkc2) Indexes() []*Index {
	return []*Index{
		NewIndex("idx_parent_id", "parent_id"),
	}
}

func (*Fkc2) ForeignKeys() []*ForeignKey {
	return []*ForeignKey{
		NewForeignKey(
			"fk_fkc2_parent_id",
			[]string{"parent_id"},
			"fkp2",
			[]string{"id"},
		),
	}
}

type Fkp3 struct {
	ID int32
}

func (*Fkp3) PrimaryKey() *PrimaryKey {
	return NewPrimaryKey("id")
}

type Fkc3 struct {
	ID       string
	ParentID sql.NullInt32 `ddl:",null"`
}

func (*Fkc3) PrimaryKey() *PrimaryKey {
	return NewPrimaryKey("id")
}

func (*Fkc3) Indexes() []*Index {
	return []*Index{
		NewIndex("idx_parent_id", "parent_id"),
	}
}

func (*Fkc3) ForeignKeys() []*ForeignKey {
	return []*ForeignKey{
		NewForeignKey(
			"fk_fkc3_parent_id",
			[]string{"parent_id"},
			"fkp3",
			[]string{"id"},
		),
	}
}

type Fkp4 struct {
	ID int16
}

func (*Fkp4) PrimaryKey() *PrimaryKey {
	return NewPrimaryKey("id")
}

type Fkc4 struct {
	ID       string
	ParentID sql.NullInt16 `ddl:",null"`
}

func (*Fkc4) PrimaryKey() *PrimaryKey {
	return NewPrimaryKey("id")
}

func (*Fkc4) Indexes() []*Index {
	return []*Index{
		NewIndex("idx_parent_id", "parent_id"),
	}
}

func (*Fkc4) ForeignKeys() []*ForeignKey {
	return []*ForeignKey{
		NewForeignKey(
			"fk_fkc4_parent_id",
			[]string{"parent_id"},
			"fkp4",
			[]string{"id"},
		),
	}
}

type Fkp5 struct {
	ID byte
}

func (*Fkp5) PrimaryKey() *PrimaryKey {
	return NewPrimaryKey("id")
}

type Fkc5 struct {
	ID       string
	ParentID sql.NullByte `ddl:",null"`
}

func (*Fkc5) PrimaryKey() *PrimaryKey {
	return NewPrimaryKey("id")
}

func (*Fkc5) Indexes() []*Index {
	return []*Index{
		NewIndex("idx_parent_id", "parent_id"),
	}
}

func (*Fkc5) ForeignKeys() []*ForeignKey {
	return []*ForeignKey{
		NewForeignKey(
			"fk_fkc5_parent_id",
			[]string{"parent_id"},
			"fkp5",
			[]string{"id"},
		),
	}
}

type Fkp6 struct {
	ID float64
}

func (*Fkp6) PrimaryKey() *PrimaryKey {
	return NewPrimaryKey("id")
}

type Fkc6 struct {
	ID       string
	ParentID sql.NullFloat64 `ddl:",null"`
}

func (*Fkc6) PrimaryKey() *PrimaryKey {
	return NewPrimaryKey("id")
}

func (*Fkc6) Indexes() []*Index {
	return []*Index{
		NewIndex("idx_parent_id", "parent_id"),
	}
}

func (*Fkc6) ForeignKeys() []*ForeignKey {
	return []*ForeignKey{
		NewForeignKey(
			"fk_fkc6_parent_id",
			[]string{"parent_id"},
			"fkp6",
			[]string{"id"},
		),
	}
}

type Fkp7 struct {
	ID bool
}

func (*Fkp7) PrimaryKey() *PrimaryKey {
	return NewPrimaryKey("id")
}

type Fkc7 struct {
	ID       string
	ParentID sql.NullBool `ddl:",null"`
}

func (*Fkc7) PrimaryKey() *PrimaryKey {
	return NewPrimaryKey("id")
}

func (*Fkc7) Indexes() []*Index {
	return []*Index{
		NewIndex("idx_parent_id", "parent_id"),
	}
}

func (*Fkc7) ForeignKeys() []*ForeignKey {
	return []*ForeignKey{
		NewForeignKey(
			"fk_fkc7_parent_id",
			[]string{"parent_id"},
			"fkp7",
			[]string{"id"},
		),
	}
}

type Fkp8 struct {
	ID uint32
}

func (*Fkp8) PrimaryKey() *PrimaryKey {
	return NewPrimaryKey("id")
}

type NullUint32 struct {
	Uint32 uint32
	Valid  bool
}
type Fkc8 struct {
	ID       string
	ParentID NullUint32 `ddl:",type=INTEGER UNSIGNED,null"`
}

func (*Fkc8) PrimaryKey() *PrimaryKey {
	return NewPrimaryKey("id")
}

func (*Fkc8) Indexes() []*Index {
	return []*Index{
		NewIndex("idx_parent_id", "parent_id"),
	}
}

func (*Fkc8) ForeignKeys() []*ForeignKey {
	return []*ForeignKey{
		NewForeignKey(
			"fk_fkc8_parent_id",
			[]string{"parent_id"},
			"fkp8",
			[]string{"id"},
		),
	}
}

func testMaker(t *testing.T, structs []any, ddl string) {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	m, err := New(&Config{
		DB: &DBConfig{
			Engine:  "InnoDB",
			Charset: "utf8mb4",
			Collate: "utf8mb4_bin",
		},
	})
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

	db, ok := setupDatabase(ctx, t)
	if !ok {
		return
	}

	// check the ddl syntax
	if _, err := db.ExecContext(ctx, got); err != nil {
		t.Errorf("failed to execute %q: %v", got, err)
	}
}

func setupDatabase(ctx context.Context, t testing.TB) (db *sql.DB, ok bool) {
	// check the ddl syntax with MySQL Server
	user := os.Getenv("MYSQL_TEST_USER")
	pass := os.Getenv("MYSQL_TEST_PASS")
	addr := os.Getenv("MYSQL_TEST_ADDR")
	if user == "" || pass == "" || addr == "" {
		return nil, false
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

	cfg.DBName = dbName
	cfg.MultiStatements = true
	db, err = sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	t.Cleanup(func() {
		db0.ExecContext(ctx, "DROP DATABASE "+dbName)
		db.Close()
	})
	return db, true
}

func testMakerError(t *testing.T, structs []any, wantErr []string) {
	t.Helper()

	m, err := New(&Config{
		DB: &DBConfig{
			Engine:  "InnoDB",
			Charset: "utf8mb4",
			Collate: "utf8mb4_bin",
		},
	})
	if err != nil {
		t.Fatalf("failed to initialize Maker: %v", err)
	}

	m.AddStructs(structs...)

	var buf bytes.Buffer
	err = m.Generate(&buf)
	if err == nil {
		t.Error("want some error, but not")
		return
	}

	var errs *validationError
	if !errors.As(err, &errs) {
		t.Errorf("unexpected error type: %T", err)
	}

	if diff := cmp.Diff(wantErr, errs.errs); diff != "" {
		t.Errorf("unexpected errors (-want/+got):\n%s", diff)
	}
}

func TestMaker_Generate(t *testing.T) {
	testMaker(t, []any{&Foo1{}}, "SET foreign_key_checks=0;\n\n"+
		"DROP TABLE IF EXISTS `foo1`;\n\n"+
		"CREATE TABLE `foo1` (\n"+
		"    `id` INTEGER NOT NULL,\n"+
		"    PRIMARY KEY (`id`)\n"+
		") ENGINE=InnoDB DEFAULT CHARACTER SET=utf8mb4 DEFAULT COLLATE=utf8mb4_bin;\n\n"+
		"SET foreign_key_checks=1;\n")

	testMaker(t, []any{&Foo2{}}, "SET foreign_key_checks=0;\n\n"+
		"DROP TABLE IF EXISTS `foo2_customized`;\n\n"+
		"CREATE TABLE `foo2_customized` (\n"+
		"    `id` INTEGER NOT NULL AUTO_INCREMENT,\n"+
		"    `name` VARCHAR(191) NOT NULL INVISIBLE COMMENT '\\'コメント\\'',\n"+
		"    INDEX `idx_name` (`name`),\n"+
		"    PRIMARY KEY (`id`)\n"+
		") ENGINE=InnoDB DEFAULT CHARACTER SET=utf8mb4 DEFAULT COLLATE=utf8mb4_bin;\n\n"+
		"SET foreign_key_checks=1;\n")

	testMaker(t, []any{&Foo3{}}, "SET foreign_key_checks=0;\n\n"+
		"DROP TABLE IF EXISTS `foo3`;\n\n"+
		"CREATE TABLE `foo3` (\n"+
		"    `id` INTEGER NOT NULL,\n"+
		"    `name` VARCHAR(191) NOT NULL,\n"+
		"    UNIQUE `idx_name` (`name`),\n"+
		"    PRIMARY KEY (`id`)\n"+
		") ENGINE=InnoDB DEFAULT CHARACTER SET=utf8mb4 DEFAULT COLLATE=utf8mb4_bin;\n\n"+
		"SET foreign_key_checks=1;\n")

	testMaker(t, []any{&Foo1{}, &Foo4{}}, "SET foreign_key_checks=0;\n\n"+
		"DROP TABLE IF EXISTS `foo1`;\n\n"+
		"CREATE TABLE `foo1` (\n"+
		"    `id` INTEGER NOT NULL,\n"+
		"    PRIMARY KEY (`id`)\n"+
		") ENGINE=InnoDB DEFAULT CHARACTER SET=utf8mb4 DEFAULT COLLATE=utf8mb4_bin;\n\n\n"+
		"DROP TABLE IF EXISTS `foo4`;\n\n"+
		"CREATE TABLE `foo4` (\n"+
		"    `id` INTEGER NOT NULL,\n"+
		"    `name` VARCHAR(191) NOT NULL,\n"+
		"    CONSTRAINT `fk_foo1` FOREIGN KEY (`id`) REFERENCES `foo1` (`id`),\n"+
		"    PRIMARY KEY (`id`)\n"+
		") ENGINE=InnoDB DEFAULT CHARACTER SET=utf8mb4 DEFAULT COLLATE=utf8mb4_bin;\n\n"+
		"SET foreign_key_checks=1;\n")

	testMaker(t, []any{&Foo5{}, &Foo1{}}, "SET foreign_key_checks=0;\n\n"+
		"DROP TABLE IF EXISTS `foo5`;\n\n"+
		"CREATE TABLE `foo5` (\n"+
		"    `id` INTEGER NOT NULL,\n"+
		"    `name` VARCHAR(191) NOT NULL,\n"+
		"    CONSTRAINT `fk_foo1` FOREIGN KEY (`id`) REFERENCES `foo1` (`id`) ON DELETE CASCADE ON UPDATE CASCADE,\n"+
		"    PRIMARY KEY (`id`)\n"+
		") ENGINE=InnoDB DEFAULT CHARACTER SET=utf8mb4 DEFAULT COLLATE=utf8mb4_bin;\n\n\n"+
		"DROP TABLE IF EXISTS `foo1`;\n\n"+
		"CREATE TABLE `foo1` (\n"+
		"    `id` INTEGER NOT NULL,\n"+
		"    PRIMARY KEY (`id`)\n"+
		") ENGINE=InnoDB DEFAULT CHARACTER SET=utf8mb4 DEFAULT COLLATE=utf8mb4_bin;\n\n"+
		"SET foreign_key_checks=1;\n")

	testMaker(t, []any{&Foo6{}}, "SET foreign_key_checks=0;\n\n"+
		"DROP TABLE IF EXISTS `foo6`;\n\n"+
		"CREATE TABLE `foo6` (\n"+
		"    `id` INTEGER NOT NULL,\n"+
		"    `name` VARCHAR(191) NOT NULL,\n"+
		"    `email` VARCHAR(191) NOT NULL,\n"+
		"    INDEX `idx_name` (`name`) COMMENT 'an index\\n\\twith \\'comment\\'',\n"+
		"    UNIQUE `uniq_email` (`email`) COMMENT 'a unique index\\n\\twith \\'comment\\'',\n"+
		"    PRIMARY KEY (`id`)\n"+
		") ENGINE=InnoDB DEFAULT CHARACTER SET=utf8mb4 DEFAULT COLLATE=utf8mb4_bin;\n\n"+
		"SET foreign_key_checks=1;\n")

	testMaker(t, []any{&Foo7{}}, "SET foreign_key_checks=0;\n\n"+
		"DROP TABLE IF EXISTS `foo7`;\n\n"+
		"CREATE TABLE `foo7` (\n"+
		"    `id` INTEGER NOT NULL,\n"+
		"    `name` VARCHAR(191) NOT NULL DEFAULT 'John Doe',\n"+
		"    PRIMARY KEY (`id`)\n"+
		") ENGINE=InnoDB DEFAULT CHARACTER SET=utf8mb4 DEFAULT COLLATE=utf8mb4_bin;\n\n"+
		"SET foreign_key_checks=1;\n")

	// invisible index
	testMaker(t, []any{&Foo8{}}, "SET foreign_key_checks=0;\n\n"+
		"DROP TABLE IF EXISTS `foo8`;\n\n"+
		"CREATE TABLE `foo8` (\n"+
		"    `id` INTEGER NOT NULL AUTO_INCREMENT,\n"+
		"    `name` VARCHAR(191) NOT NULL,\n"+
		"    INDEX `idx_name` (`name`) INVISIBLE,\n"+
		"    PRIMARY KEY (`id`)\n"+
		") ENGINE=InnoDB DEFAULT CHARACTER SET=utf8mb4 DEFAULT COLLATE=utf8mb4_bin;\n\n"+
		"SET foreign_key_checks=1;\n")

	// charset and collate
	testMaker(t, []any{&Foo9{}}, "SET foreign_key_checks=0;\n\n"+
		"DROP TABLE IF EXISTS `foo9`;\n\n"+
		"CREATE TABLE `foo9` (\n"+
		"    `id` INTEGER NOT NULL AUTO_INCREMENT,\n"+
		"    `name` VARCHAR(191) CHARACTER SET utf8 COLLATE utf8_bin NOT NULL,\n"+
		"    PRIMARY KEY (`id`)\n"+
		") ENGINE=InnoDB DEFAULT CHARACTER SET=utf8mb4 DEFAULT COLLATE=utf8mb4_bin;\n\n"+
		"SET foreign_key_checks=1;\n")

	// FULLTEXT INDEX
	testMaker(t, []any{&Foo10{}}, "SET foreign_key_checks=0;\n\n"+
		"DROP TABLE IF EXISTS `foo10`;\n\n"+
		"CREATE TABLE `foo10` (\n"+
		"    `id` INTEGER NOT NULL AUTO_INCREMENT,\n"+
		"    `text` VARCHAR(191) NOT NULL,\n"+
		"    FULLTEXT INDEX `idx_text` (`text`) WITH PARSER ngram COMMENT 'FULLTEXT INDEX',\n"+
		"    PRIMARY KEY (`id`)\n"+
		") ENGINE=InnoDB DEFAULT CHARACTER SET=utf8mb4 DEFAULT COLLATE=utf8mb4_bin;\n\n"+
		"SET foreign_key_checks=1;\n")

	// SPATIAL INDEX
	testMaker(t, []any{&Foo11{}}, "SET foreign_key_checks=0;\n\n"+
		"DROP TABLE IF EXISTS `foo11`;\n\n"+
		"CREATE TABLE `foo11` (\n"+
		"    `id` INTEGER NOT NULL AUTO_INCREMENT,\n"+
		"    `point` GEOMETRY NOT NULL,\n"+
		"    SPATIAL INDEX `idx_point` (`point`) COMMENT 'SPATIAL INDEX',\n"+
		"    PRIMARY KEY (`id`)\n"+
		") ENGINE=InnoDB DEFAULT CHARACTER SET=utf8mb4 DEFAULT COLLATE=utf8mb4_bin;\n\n"+
		"SET foreign_key_checks=1;\n")

	// JSON
	testMaker(t, []any{&Foo20{}}, "SET foreign_key_checks=0;\n\n"+
		"DROP TABLE IF EXISTS `foo20`;\n\n"+
		"CREATE TABLE `foo20` (\n"+
		"    `id` INTEGER NOT NULL AUTO_INCREMENT,\n"+
		"    `json` JSON NOT NULL,\n"+
		"    PRIMARY KEY (`id`)\n"+
		") ENGINE=InnoDB DEFAULT CHARACTER SET=utf8mb4 DEFAULT COLLATE=utf8mb4_bin;\n\n"+
		"SET foreign_key_checks=1;\n")

	// unsigned integer specified in struct tag
	testMaker(t, []any{&Foo21{}}, "SET foreign_key_checks=0;\n\n"+
		"DROP TABLE IF EXISTS `foo21`;\n\n"+
		"CREATE TABLE `foo21` (\n"+
		"    `id` INTEGER NOT NULL,\n"+
		"    `parent_id` INTEGER UNSIGNED NULL,\n"+
		"    PRIMARY KEY (`id`)\n"+
		") ENGINE=InnoDB DEFAULT CHARACTER SET=utf8mb4 DEFAULT COLLATE=utf8mb4_bin;\n\n"+
		"SET foreign_key_checks=1;\n")

	// NULL foreign key
	testMaker(t, []any{&Fkp1{}, &Fkc1{}}, "SET foreign_key_checks=0;\n\n"+
		"DROP TABLE IF EXISTS `fkp1`;\n\n"+
		"CREATE TABLE `fkp1` (\n"+
		"    `id` VARCHAR(191) NOT NULL,\n"+
		"    PRIMARY KEY (`id`)\n"+
		") ENGINE=InnoDB DEFAULT CHARACTER SET=utf8mb4 DEFAULT COLLATE=utf8mb4_bin;\n\n\n"+
		"DROP TABLE IF EXISTS `fkc1`;\n\n"+
		"CREATE TABLE `fkc1` (\n"+
		"    `id` VARCHAR(191) NOT NULL,\n"+
		"    `parent_id` VARCHAR(191) NULL,\n"+
		"    INDEX `idx_parent_id` (`parent_id`),\n"+
		"    CONSTRAINT `fk_fkc1_parent_id` FOREIGN KEY (`parent_id`) REFERENCES `fkp1` (`id`),\n"+
		"    PRIMARY KEY (`id`)\n"+
		") ENGINE=InnoDB DEFAULT CHARACTER SET=utf8mb4 DEFAULT COLLATE=utf8mb4_bin;\n\n"+
		"SET foreign_key_checks=1;\n")

	testMaker(t, []any{&Fkp2{}, &Fkc2{}}, "SET foreign_key_checks=0;\n\n"+
		"DROP TABLE IF EXISTS `fkp2`;\n\n"+
		"CREATE TABLE `fkp2` (\n"+
		"    `id` BIGINT NOT NULL,\n"+
		"    PRIMARY KEY (`id`)\n"+
		") ENGINE=InnoDB DEFAULT CHARACTER SET=utf8mb4 DEFAULT COLLATE=utf8mb4_bin;\n\n\n"+
		"DROP TABLE IF EXISTS `fkc2`;\n\n"+
		"CREATE TABLE `fkc2` (\n"+
		"    `id` VARCHAR(191) NOT NULL,\n"+
		"    `parent_id` BIGINT NULL,\n"+
		"    INDEX `idx_parent_id` (`parent_id`),\n"+
		"    CONSTRAINT `fk_fkc2_parent_id` FOREIGN KEY (`parent_id`) REFERENCES `fkp2` (`id`),\n"+
		"    PRIMARY KEY (`id`)\n"+
		") ENGINE=InnoDB DEFAULT CHARACTER SET=utf8mb4 DEFAULT COLLATE=utf8mb4_bin;\n\n"+
		"SET foreign_key_checks=1;\n")

	testMaker(t, []any{&Fkp3{}, &Fkc3{}}, "SET foreign_key_checks=0;\n\n"+
		"DROP TABLE IF EXISTS `fkp3`;\n\n"+
		"CREATE TABLE `fkp3` (\n"+
		"    `id` INTEGER NOT NULL,\n"+
		"    PRIMARY KEY (`id`)\n"+
		") ENGINE=InnoDB DEFAULT CHARACTER SET=utf8mb4 DEFAULT COLLATE=utf8mb4_bin;\n\n\n"+
		"DROP TABLE IF EXISTS `fkc3`;\n\n"+
		"CREATE TABLE `fkc3` (\n"+
		"    `id` VARCHAR(191) NOT NULL,\n"+
		"    `parent_id` INTEGER NULL,\n"+
		"    INDEX `idx_parent_id` (`parent_id`),\n"+
		"    CONSTRAINT `fk_fkc3_parent_id` FOREIGN KEY (`parent_id`) REFERENCES `fkp3` (`id`),\n"+
		"    PRIMARY KEY (`id`)\n"+
		") ENGINE=InnoDB DEFAULT CHARACTER SET=utf8mb4 DEFAULT COLLATE=utf8mb4_bin;\n\n"+
		"SET foreign_key_checks=1;\n")

	testMaker(t, []any{&Fkp4{}, &Fkc4{}}, "SET foreign_key_checks=0;\n\n"+
		"DROP TABLE IF EXISTS `fkp4`;\n\n"+
		"CREATE TABLE `fkp4` (\n"+
		"    `id` SMALLINT NOT NULL,\n"+
		"    PRIMARY KEY (`id`)\n"+
		") ENGINE=InnoDB DEFAULT CHARACTER SET=utf8mb4 DEFAULT COLLATE=utf8mb4_bin;\n\n\n"+
		"DROP TABLE IF EXISTS `fkc4`;\n\n"+
		"CREATE TABLE `fkc4` (\n"+
		"    `id` VARCHAR(191) NOT NULL,\n"+
		"    `parent_id` SMALLINT NULL,\n"+
		"    INDEX `idx_parent_id` (`parent_id`),\n"+
		"    CONSTRAINT `fk_fkc4_parent_id` FOREIGN KEY (`parent_id`) REFERENCES `fkp4` (`id`),\n"+
		"    PRIMARY KEY (`id`)\n"+
		") ENGINE=InnoDB DEFAULT CHARACTER SET=utf8mb4 DEFAULT COLLATE=utf8mb4_bin;\n\n"+
		"SET foreign_key_checks=1;\n")

	testMaker(t, []any{&Fkp5{}, &Fkc5{}}, "SET foreign_key_checks=0;\n\n"+
		"DROP TABLE IF EXISTS `fkp5`;\n\n"+
		"CREATE TABLE `fkp5` (\n"+
		"    `id` TINYINT UNSIGNED NOT NULL,\n"+
		"    PRIMARY KEY (`id`)\n"+
		") ENGINE=InnoDB DEFAULT CHARACTER SET=utf8mb4 DEFAULT COLLATE=utf8mb4_bin;\n\n\n"+
		"DROP TABLE IF EXISTS `fkc5`;\n\n"+
		"CREATE TABLE `fkc5` (\n"+
		"    `id` VARCHAR(191) NOT NULL,\n"+
		"    `parent_id` TINYINT UNSIGNED NULL,\n"+
		"    INDEX `idx_parent_id` (`parent_id`),\n"+
		"    CONSTRAINT `fk_fkc5_parent_id` FOREIGN KEY (`parent_id`) REFERENCES `fkp5` (`id`),\n"+
		"    PRIMARY KEY (`id`)\n"+
		") ENGINE=InnoDB DEFAULT CHARACTER SET=utf8mb4 DEFAULT COLLATE=utf8mb4_bin;\n\n"+
		"SET foreign_key_checks=1;\n")

	testMaker(t, []any{&Fkp6{}, &Fkc6{}}, "SET foreign_key_checks=0;\n\n"+
		"DROP TABLE IF EXISTS `fkp6`;\n\n"+
		"CREATE TABLE `fkp6` (\n"+
		"    `id` DOUBLE NOT NULL,\n"+
		"    PRIMARY KEY (`id`)\n"+
		") ENGINE=InnoDB DEFAULT CHARACTER SET=utf8mb4 DEFAULT COLLATE=utf8mb4_bin;\n\n\n"+
		"DROP TABLE IF EXISTS `fkc6`;\n\n"+
		"CREATE TABLE `fkc6` (\n"+
		"    `id` VARCHAR(191) NOT NULL,\n"+
		"    `parent_id` DOUBLE NULL,\n"+
		"    INDEX `idx_parent_id` (`parent_id`),\n"+
		"    CONSTRAINT `fk_fkc6_parent_id` FOREIGN KEY (`parent_id`) REFERENCES `fkp6` (`id`),\n"+
		"    PRIMARY KEY (`id`)\n"+
		") ENGINE=InnoDB DEFAULT CHARACTER SET=utf8mb4 DEFAULT COLLATE=utf8mb4_bin;\n\n"+
		"SET foreign_key_checks=1;\n")

	testMaker(t, []any{&Fkp7{}, &Fkc7{}}, "SET foreign_key_checks=0;\n\n"+
		"DROP TABLE IF EXISTS `fkp7`;\n\n"+
		"CREATE TABLE `fkp7` (\n"+
		"    `id` TINYINT(1) NOT NULL,\n"+
		"    PRIMARY KEY (`id`)\n"+
		") ENGINE=InnoDB DEFAULT CHARACTER SET=utf8mb4 DEFAULT COLLATE=utf8mb4_bin;\n\n\n"+
		"DROP TABLE IF EXISTS `fkc7`;\n\n"+
		"CREATE TABLE `fkc7` (\n"+
		"    `id` VARCHAR(191) NOT NULL,\n"+
		"    `parent_id` TINYINT(1) NULL,\n"+
		"    INDEX `idx_parent_id` (`parent_id`),\n"+
		"    CONSTRAINT `fk_fkc7_parent_id` FOREIGN KEY (`parent_id`) REFERENCES `fkp7` (`id`),\n"+
		"    PRIMARY KEY (`id`)\n"+
		") ENGINE=InnoDB DEFAULT CHARACTER SET=utf8mb4 DEFAULT COLLATE=utf8mb4_bin;\n\n"+
		"SET foreign_key_checks=1;\n")

	testMaker(t, []any{&Fkp8{}, &Fkc8{}}, "SET foreign_key_checks=0;\n\n"+
		"DROP TABLE IF EXISTS `fkp8`;\n\n"+
		"CREATE TABLE `fkp8` (\n"+
		"    `id` INTERGER UNSIGNED NOT NULL,\n"+
		"    PRIMARY KEY (`id`)\n"+
		") ENGINE=InnoDB DEFAULT CHARACTER SET=utf8mb4 DEFAULT COLLATE=utf8mb4_bin;\n\n\n"+
		"DROP TABLE IF EXISTS `fkc8`;\n\n"+
		"CREATE TABLE `fkc8` (\n"+
		"    `id` VARCHAR(191) NOT NULL,\n"+
		"    `parent_id` INTERGER UNSIGNED NULL,\n"+
		"    INDEX `idx_parent_id` (`parent_id`),\n"+
		"    CONSTRAINT `fk_fkc8_parent_id` FOREIGN KEY (`parent_id`) REFERENCES `fkp8` (`id`),\n"+
		"    PRIMARY KEY (`id`)\n"+
		") ENGINE=InnoDB DEFAULT CHARACTER SET=utf8mb4 DEFAULT COLLATE=utf8mb4_bin;\n\n"+
		"SET foreign_key_checks=1;\n")

	// testMaker(t, []any{&Fkp2{}, &Fkc2{}}, "SET foreign_key_checks=0;\n\n"+
	// 	"DROP TABLE IF EXISTS `fkp2`;\n\n"+
	// 	"CREATE TABLE `fkp2` (\n"+
	// 	"    `id` INTEGER UNSIGNED NOT NULL,\n"+
	// 	"    PRIMARY KEY (`id`)\n"+
	// 	") ENGINE=InnoDB DEFAULT CHARACTER SET=utf8mb4 DEFAULT COLLATE=utf8mb4_bin;\n\n\n"+
	// 	"DROP TABLE IF EXISTS `fkc1`;\n\n"+
	// 	"CREATE TABLE `fkc1` (\n"+
	// 	"    `id` VARCHAR(191) NOT NULL,\n"+
	// 	"    `parent_id` VARCHAR(191) NULL,\n"+
	// 	"    CONSTRAINT `fk_fkc2_parent_id` FOREIGN KEY (`id`) REFERENCES `fkp2` (`id`),\n"+
	// 	"    PRIMARY KEY (`id`)\n"+
	// 	") ENGINE=InnoDB DEFAULT CHARACTER SET=utf8mb4 DEFAULT COLLATE=utf8mb4_bin;\n\n"+
	// 	"SET foreign_key_checks=1;\n")

	testMakerError(t, []any{&Foo11{}, &Foo12{}}, []string{
		`duplicated name of table: "foo11"`,
	})

	testMakerError(t, []any{&Foo13{}}, []string{
		`table "foo13": duplicated name of column: "id"`,
	})

	testMakerError(t, []any{&Foo14{}}, []string{
		`table "foo14", primary key: column "unknown_column" not found`,
		`table "foo14", index "idx": column "unknown_column" not found`,
		`table "foo14", unique index "uniq": column "unknown_column" not found`,
	})

	testMakerError(t, []any{&Foo15{}}, []string{
		`table "foo15": duplicated name of index: "idx_name"`,
		`table "foo15": duplicated name of index: "idx_name"`,
		`table "foo15": duplicated name of index: "idx_name"`,
	})

	testMakerError(t, []any{&Foo16{}}, []string{
		`table "foo16": duplicated name of foreign key constraint: "fk_duplicated"`,
	})

	testMakerError(t, []any{&Foo17{}}, []string{
		`table "foo17", foreign key "fk_foo17": column "unknown_column" not found`,
		`table "foo17", foreign key "fk_foo17": referenced table "unknown_table" not found`,
	})

	testMakerError(t, []any{&Foo18{}, &Foo19{}}, []string{
		`table "foo18", foreign key "fk_foo19": index required on table "foo18"`,
		`table "foo18", foreign key "fk_foo19": column "foo19_id" and referenced column "foo19"."id" type mismatch`,
	})
}

func TestMaker_GenerateGo(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fn := func(dir string) func(t *testing.T) {
		// gen generates 'scheme.sql' and 'scheme_gen.go'
		gen := func(t *testing.T) error {
			ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
			defer cancel()

			var buf bytes.Buffer
			args := []string{"run", "-tags", "myddlmaker", filepath.Join("gen", "main.go")}
			cmd := exec.CommandContext(ctx, goTool(), args...)
			cmd.Stdout = &buf
			cmd.Stderr = &buf
			cmd.Dir = dir
			if err := cmd.Run(); err != nil {
				t.Errorf("failed to generate: %v, output:\n%s", err, buf.String())
				return err
			}
			return nil
		}

		runTests := func(t *testing.T) error {
			ctx, cancel := context.WithTimeout(ctx, 15*time.Second)
			defer cancel()

			var buf bytes.Buffer
			args := []string{"test"}
			cmd := exec.CommandContext(ctx, goTool(), args...)
			cmd.Stdout = &buf
			cmd.Stderr = &buf
			cmd.Dir = dir
			if err := cmd.Run(); err != nil {
				t.Errorf("failed to run test: %v, output:\n%s", err, buf.String())
				return err
			}
			return nil
		}
		return func(t *testing.T) {
			if err := gen(t); err != nil {
				return
			}

			// check the ddl syntax with MySQL Server
			user := os.Getenv("MYSQL_TEST_USER")
			pass := os.Getenv("MYSQL_TEST_PASS")
			addr := os.Getenv("MYSQL_TEST_ADDR")
			if user != "" && pass != "" && addr != "" {
				ddl, err := os.ReadFile(filepath.Join(dir, "schema.sql"))
				if err != nil {
					t.Errorf("failed read schema.sql: %v", err)
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
				defer db0.ExecContext(ctx, "DROP DATABASE "+dbName)
				t.Setenv("MYSQL_TEST_DB", dbName)

				// apply the ddl
				cfg.DBName = dbName
				cfg.MultiStatements = true
				db, err := sql.Open("mysql", cfg.FormatDSN())
				if err != nil {
					t.Fatalf("failed to open db: %v", err)
				}
				defer db.Close()
				if _, err := db.ExecContext(ctx, string(ddl)); err != nil {
					t.Errorf("failed to execute %q: %v", string(ddl), err)
				}
			}

			if err := runTests(t); err != nil {
				return
			}
		}
	}
	dirs, err := filepath.Glob("./testdata/*")
	if err != nil {
		t.Fatal(err)
	}
	for _, dir := range dirs {
		stat, err := os.Stat(dir)
		if err != nil {
			t.Error(err)
			continue
		}
		if !stat.IsDir() {
			continue
		}
		t.Run(dir, fn(dir))
	}
}

// goTool reports the path of the go tool to use to run the tests.
// If possible, use the same Go used to run run.go, otherwise
// fallback to the go version found in the PATH.
func goTool() string {
	var exeSuffix string
	if runtime.GOOS == "windows" {
		exeSuffix = ".exe"
	}
	path := filepath.Join(runtime.GOROOT(), "bin", "go"+exeSuffix)
	if _, err := os.Stat(path); err == nil {
		return path
	}
	// Just run "go" from PATH
	return "go"
}
