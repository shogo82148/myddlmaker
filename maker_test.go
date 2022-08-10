package myddlmaker

import (
	"bytes"
	"context"
	"crypto/rand"
	"database/sql"
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

func testMaker(t *testing.T, structs []any, ddl string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	m, err := New(&Config{
		DB: &DBConfig{
			Engine:  "InnoDB",
			Charset: "utf8mb4",
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
	defer db0.ExecContext(ctx, "DROP DATABASE "+dbName)

	cfg.DBName = dbName
	cfg.MultiStatements = true
	db, err := sql.Open("mysql", cfg.FormatDSN())
	if err != nil {
		t.Fatalf("failed to open db: %v", err)
	}
	defer db.Close()

	// check the ddl syntax
	if _, err := db.ExecContext(ctx, got); err != nil {
		t.Errorf("failed to execute %q: %v", got, err)
	}
}

func TestMaker_Generate(t *testing.T) {
	testMaker(t, []any{&Foo1{}}, "SET foreign_key_checks=0;\n\n"+
		"DROP TABLE IF EXISTS `foo1`;\n\n"+
		"CREATE TABLE `foo1` (\n"+
		"    `id` INTEGER NOT NULL,\n"+
		"    PRIMARY KEY (`id`)\n"+
		") ENGINE = InnoDB DEFAULT CHARACTER SET = utf8mb4;\n\n"+
		"SET foreign_key_checks=1;\n")

	testMaker(t, []any{&Foo2{}}, "SET foreign_key_checks=0;\n\n"+
		"DROP TABLE IF EXISTS `foo2_customized`;\n\n"+
		"CREATE TABLE `foo2_customized` (\n"+
		"    `id` INTEGER NOT NULL AUTO_INCREMENT,\n"+
		"    `name` VARCHAR(191) NOT NULL INVISIBLE COMMENT '\\'コメント\\'',\n"+
		"    INDEX `idx_name` (`name`),\n"+
		"    PRIMARY KEY (`id`)\n"+
		") ENGINE = InnoDB DEFAULT CHARACTER SET = utf8mb4;\n\n"+
		"SET foreign_key_checks=1;\n")

	testMaker(t, []any{&Foo3{}}, "SET foreign_key_checks=0;\n\n"+
		"DROP TABLE IF EXISTS `foo3`;\n\n"+
		"CREATE TABLE `foo3` (\n"+
		"    `id` INTEGER NOT NULL,\n"+
		"    `name` VARCHAR(191) NOT NULL,\n"+
		"    UNIQUE `idx_name` (`name`),\n"+
		"    PRIMARY KEY (`id`)\n"+
		") ENGINE = InnoDB DEFAULT CHARACTER SET = utf8mb4;\n\n"+
		"SET foreign_key_checks=1;\n")

	testMaker(t, []any{&Foo1{}, &Foo4{}}, "SET foreign_key_checks=0;\n\n"+
		"DROP TABLE IF EXISTS `foo1`;\n\n"+
		"CREATE TABLE `foo1` (\n"+
		"    `id` INTEGER NOT NULL,\n"+
		"    PRIMARY KEY (`id`)\n"+
		") ENGINE = InnoDB DEFAULT CHARACTER SET = utf8mb4;\n\n\n"+
		"DROP TABLE IF EXISTS `foo4`;\n\n"+
		"CREATE TABLE `foo4` (\n"+
		"    `id` INTEGER NOT NULL,\n"+
		"    `name` VARCHAR(191) NOT NULL,\n"+
		"    CONSTRAINT `fk_foo1` FOREIGN KEY (`id`) REFERENCES `foo1` (`id`),\n"+
		"    PRIMARY KEY (`id`)\n"+
		") ENGINE = InnoDB DEFAULT CHARACTER SET = utf8mb4;\n\n"+
		"SET foreign_key_checks=1;\n")

	testMaker(t, []any{&Foo5{}, &Foo1{}}, "SET foreign_key_checks=0;\n\n"+
		"DROP TABLE IF EXISTS `foo5`;\n\n"+
		"CREATE TABLE `foo5` (\n"+
		"    `id` INTEGER NOT NULL,\n"+
		"    `name` VARCHAR(191) NOT NULL,\n"+
		"    CONSTRAINT `fk_foo1` FOREIGN KEY (`id`) REFERENCES `foo1` (`id`) ON UPDATE CASCADE ON DELETE CASCADE,\n"+
		"    PRIMARY KEY (`id`)\n"+
		") ENGINE = InnoDB DEFAULT CHARACTER SET = utf8mb4;\n\n\n"+
		"DROP TABLE IF EXISTS `foo1`;\n\n"+
		"CREATE TABLE `foo1` (\n"+
		"    `id` INTEGER NOT NULL,\n"+
		"    PRIMARY KEY (`id`)\n"+
		") ENGINE = InnoDB DEFAULT CHARACTER SET = utf8mb4;\n\n"+
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
		") ENGINE = InnoDB DEFAULT CHARACTER SET = utf8mb4;\n\n"+
		"SET foreign_key_checks=1;\n")

	testMaker(t, []any{&Foo7{}}, "SET foreign_key_checks=0;\n\n"+
		"DROP TABLE IF EXISTS `foo7`;\n\n"+
		"CREATE TABLE `foo7` (\n"+
		"    `id` INTEGER NOT NULL,\n"+
		"    `name` VARCHAR(191) NOT NULL DEFAULT 'John Doe',\n"+
		"    PRIMARY KEY (`id`)\n"+
		") ENGINE = InnoDB DEFAULT CHARACTER SET = utf8mb4;\n\n"+
		"SET foreign_key_checks=1;\n")

	// invisible index
	testMaker(t, []any{&Foo8{}}, "SET foreign_key_checks=0;\n\n"+
		"DROP TABLE IF EXISTS `foo8`;\n\n"+
		"CREATE TABLE `foo8` (\n"+
		"    `id` INTEGER NOT NULL AUTO_INCREMENT,\n"+
		"    `name` VARCHAR(191) NOT NULL,\n"+
		"    INDEX `idx_name` (`name`) INVISIBLE,\n"+
		"    PRIMARY KEY (`id`)\n"+
		") ENGINE = InnoDB DEFAULT CHARACTER SET = utf8mb4;\n\n"+
		"SET foreign_key_checks=1;\n")

	// charset and collate
	testMaker(t, []any{&Foo9{}}, "SET foreign_key_checks=0;\n\n"+
		"DROP TABLE IF EXISTS `foo9`;\n\n"+
		"CREATE TABLE `foo9` (\n"+
		"    `id` INTEGER NOT NULL AUTO_INCREMENT,\n"+
		"    `name` VARCHAR(191) CHARACTER SET utf8 COLLATE utf8_bin NOT NULL,\n"+
		"    PRIMARY KEY (`id`)\n"+
		") ENGINE = InnoDB DEFAULT CHARACTER SET = utf8mb4;\n\n"+
		"SET foreign_key_checks=1;\n")

	// FULLTEXT INDEX
	testMaker(t, []any{&Foo10{}}, "SET foreign_key_checks=0;\n\n"+
		"DROP TABLE IF EXISTS `foo10`;\n\n"+
		"CREATE TABLE `foo10` (\n"+
		"    `id` INTEGER NOT NULL AUTO_INCREMENT,\n"+
		"    `text` VARCHAR(191) NOT NULL,\n"+
		"    FULLTEXT INDEX `idx_text` (`text`) WITH PARSER ngram COMMENT 'FULLTEXT INDEX',\n"+
		"    PRIMARY KEY (`id`)\n"+
		") ENGINE = InnoDB DEFAULT CHARACTER SET = utf8mb4;\n\n"+
		"SET foreign_key_checks=1;\n")

	// SPATIAL INDEX
	testMaker(t, []any{&Foo11{}}, "SET foreign_key_checks=0;\n\n"+
		"DROP TABLE IF EXISTS `foo11`;\n\n"+
		"CREATE TABLE `foo11` (\n"+
		"    `id` INTEGER NOT NULL AUTO_INCREMENT,\n"+
		"    `point` GEOMETRY NOT NULL,\n"+
		"    SPATIAL INDEX `idx_point` (`point`) COMMENT 'SPATIAL INDEX',\n"+
		"    PRIMARY KEY (`id`)\n"+
		") ENGINE = InnoDB DEFAULT CHARACTER SET = utf8mb4;\n\n"+
		"SET foreign_key_checks=1;\n")
}

func TestMaker_GenerateGo(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	fn := func(dir string) func(t *testing.T) {
		// gen generates 'scheme.sql' and 'scheme_gen.go'
		gen := func(t *testing.T) error {
			ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
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
			ctx, cancel := context.WithTimeout(ctx, 3*time.Second)
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
