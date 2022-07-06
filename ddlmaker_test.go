package myddlmaker

import (
	"bytes"
	"database/sql"
	"testing"
	"time"
)

type Test1 struct {
	ID        uint64
	Name      string
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (t1 Test1) PrimaryKey() *PrimaryKey {
	return AddPrimaryKey("id")
}

type Test2 struct {
	ID        uint64
	Test1ID   uint64
	Comment   sql.NullString `ddl:"null"`
	CreatedAt time.Time
	UpdatedAt time.Time
}

func (t2 *Test2) PrimaryKey() *PrimaryKey {
	return AddPrimaryKey("id", "created_at")
}

func TestNew(t *testing.T) {
	conf := Config{}
	_, err := New(conf)
	if err == nil {
		t.Fatal("Not set driver name", err)
	}

	conf = Config{
		DB: DBConfig{Driver: "dummy"},
	}
	_, err = New(conf)
	if err == nil {
		t.Fatal("Set unsupport driver name", err)
	}

	conf = Config{
		DB: DBConfig{Driver: "mysql"},
	}
	_, err = New(conf)
	if err != nil {
		t.Fatal(err)
	}
}

func TestAddStruct(t *testing.T) {
	dm, err := New(Config{
		DB: DBConfig{Driver: "mysql"},
	})
	if err != nil {
		t.Fatal(err)
	}
	err = dm.AddStruct(nil)
	if err == nil {
		t.Fatal("nil is not support")
	}

	dm.AddStruct(Test1{}, Test2{})
	if len(dm.Structs) != 2 {
		t.Fatal("[error] add stuct")
	}

	err = dm.AddStruct(Test1{})
	if err != nil {
		t.Fatal("[error] add duplicate struct")
	}
}

func TestGenerate(t *testing.T) {
	generatedDDL := "SET foreign_key_checks=0;\n" +
		"\n" +
		"DROP TABLE IF EXISTS `test1`;\n" +
		"\n" +
		"CREATE TABLE `test1` (\n" +
		"    `id` BIGINT unsigned NOT NULL,\n" +
		"    `name` VARCHAR(191) NOT NULL,\n" +
		"    `created_at` DATETIME NOT NULL,\n" +
		"    `updated_at` DATETIME NOT NULL,\n" +
		"    PRIMARY KEY (`id`)\n" +
		") ENGINE=InnoDB DEFAULT CHARACTER SET utf8mb4;\n" +
		"\n" +
		"SET foreign_key_checks=1;\n"

	generatedDDL2 := "SET foreign_key_checks=0;\n" +
		"\n" +
		"DROP TABLE IF EXISTS `test2`;\n" +
		"\n" +
		"CREATE TABLE `test2` (\n" +
		"    `id` BIGINT unsigned NOT NULL,\n" +
		"    `test1_id` BIGINT unsigned NOT NULL,\n" +
		"    `comment` VARCHAR(191) NULL,\n" +
		"    `created_at` DATETIME NOT NULL,\n" +
		"    `updated_at` DATETIME NOT NULL,\n" +
		"    PRIMARY KEY (`id`, `created_at`)\n" +
		") ENGINE=InnoDB DEFAULT CHARACTER SET utf8mb4;\n" +
		"\n" +
		"SET foreign_key_checks=1;\n"

	dm, err := New(Config{
		DB: DBConfig{
			Driver:  "mysql",
			Engine:  "InnoDB",
			Charset: "utf8mb4",
		},
	})
	if err != nil {
		t.Fatal("error new maker", err)
	}

	err = dm.AddStruct(&Test1{})
	if err != nil {
		t.Fatal("error add struct", err)
	}
	dm.parse()

	var ddl1 bytes.Buffer
	err = dm.generate(&ddl1)
	if err != nil {
		t.Fatal("error generate ddl", err)
	}

	if ddl1.String() != generatedDDL {
		t.Log(ddl1.String())
		t.Fatalf("generatedDDL: %s \n checkDDLL: %s \n", ddl1.String(), generatedDDL)
	}

	dm2, err := New(Config{
		DB: DBConfig{
			Driver:  "mysql",
			Engine:  "InnoDB",
			Charset: "utf8mb4",
		},
	})
	if err != nil {
		t.Fatal("error new maker", err)
	}

	err = dm2.AddStruct(&Test2{})
	if err != nil {
		t.Fatal("error add pointer struct", err)
	}
	dm2.parse()

	var ddl2 bytes.Buffer
	err = dm2.generate(&ddl2)
	if err != nil {
		t.Fatal("error generate ddl", err)
	}

	if ddl2.String() != generatedDDL2 {
		t.Fatalf("generatedDDL: %s \n checkDDLL: %s \n", ddl2.String(), generatedDDL2)
	}
}
