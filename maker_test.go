package myddlmaker

import (
	"bytes"
	"testing"
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

func TestMaker(t *testing.T) {
	m, err := New(&Config{})
	if err != nil {
		t.Fatal(err)
	}

	m.AddStructs(&Foo1{})

	var buf bytes.Buffer
	if err := m.Generate(&buf); err != nil {
		t.Fatal(err)
	}

	got := buf.String()
	want := "SET foreign_key_checks=0;\n" +
		"DROP TABLE IF EXISTS `foo1`;\n\n" +
		"CREATE TABLE `foo1` (\n" +
		"    `id` INTEGER NOT NULL,\n" +
		"    PRIMARY KEY (`id`)\n" +
		") ENGINE=InnoDB DEFAULT CHARACTER SET = 'utf8mb4';\n\n" +
		"SET foreign_key_checks=1;\n"
	if got != want {
		t.Errorf("want %q, got %q", want, got)
	}
}

func TestMaker2(t *testing.T) {
	m, err := New(&Config{})
	if err != nil {
		t.Fatal(err)
	}

	m.AddStructs(&Foo2{})

	var buf bytes.Buffer
	if err := m.Generate(&buf); err != nil {
		t.Fatal(err)
	}

	got := buf.String()
	want := "SET foreign_key_checks=0;\n" +
		"DROP TABLE IF EXISTS `foo2`;\n\n" +
		"CREATE TABLE `foo2` (\n" +
		"    `id` INTEGER NOT NULL,\n" +
		"    `name` VARCHAR(191) NOT NULL,\n" +
		"    INDEX `idx_name` (`name`),\n" +
		"    PRIMARY KEY (`id`)\n" +
		") ENGINE=InnoDB DEFAULT CHARACTER SET = 'utf8mb4';\n\n" +
		"SET foreign_key_checks=1;\n"
	if got != want {
		t.Errorf("want %q, got %q", want, got)
	}
}
