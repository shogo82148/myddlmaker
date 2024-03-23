//go:build go1.22
// +build go1.22

package myddlmaker

import (
	"database/sql"
	"testing"
)

type Foo24 struct {
	ID     int32 `ddl:",auto"`
	Number sql.Null[uint32]
}

func (*Foo24) PrimaryKey() *PrimaryKey {
	return NewPrimaryKey("id")
}

func TestMaker_Null(t *testing.T) {
	testMaker(t, []any{&Foo24{}}, "SET foreign_key_checks=0;\n\n"+
		"DROP TABLE IF EXISTS `foo24`;\n\n"+
		"CREATE TABLE `foo24` (\n"+
		"    `id` INTEGER NOT NULL AUTO_INCREMENT,\n"+
		"    `number` INTEGER UNSIGNED NOT NULL,\n"+
		"    PRIMARY KEY (`id`)\n"+
		") ENGINE=InnoDB DEFAULT CHARACTER SET=utf8mb4 DEFAULT COLLATE=utf8mb4_bin;\n\n"+
		"SET foreign_key_checks=1;\n")
}
