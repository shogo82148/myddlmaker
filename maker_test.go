package myddlmaker

import (
	"bytes"
	"testing"

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

func (*Foo2) Indexes() []Index {
	return []Index{
		NewIndex("idx_name", "name"),
	}
}

func testMaker(t *testing.T, structs []any, ddl string) {
	m, err := New(&Config{})
	if err != nil {
		t.Fatal(err)
	}

	m.AddStructs(structs...)

	var buf bytes.Buffer
	if err := m.Generate(&buf); err != nil {
		t.Fatal(err)
	}

	got := buf.String()
	if diff := cmp.Diff(ddl, got); diff != "" {
		t.Errorf("ddl is not match: (-want/+got)\n%s", diff)
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
}
