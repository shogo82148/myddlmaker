package myddlmaker

import (
	"fmt"
	"testing"
)

type myInt int64

type FooBar struct {
	// primitive types
	Int8   int8
	Int16  int16
	Int32  int32
	Int64  int64
	Uint8  uint8
	Uint16 uint16
	Uint32 uint32
	Uint64 uint64
	String string
	Bool   bool

	// custom type
	MyInt myInt

	// pointers
	PInt8  *int8
	PPInt8 **int8

	// customize the name
	Hoge int32 `ddl:"fuga"`
}

func TestTable(t *testing.T) {
	table, err := newTable(&FooBar{})
	if err != nil {
		t.Fatal(err)
	}
	if table.name != "foo_bar" {
		t.Errorf("unexpected table name: want %q, got %q", "foo_bar", table.name)
	}

	for _, col := range table.columns {
		fmt.Println(col)
	}
}
