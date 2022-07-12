package myddlmaker

import (
	"database/sql"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
)

type myInt int64
type customType struct{}

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
	MyInt      myInt
	CustomType customType `ddl:",type=TIMESTAMP"`

	// pointers
	PInt8  *int8
	PPInt8 **int8

	// customize the name
	Hoge int32 `ddl:"fuga"`
	Fuga int32 `ddl:"-"`

	// well-known struct types
	Time        time.Time
	NullTime    sql.NullTime
	NullString  sql.NullString
	NullBool    sql.NullBool
	NullByte    sql.NullByte
	NullFloat64 sql.NullFloat64
	NullInt16   sql.NullInt16
	NullInt32   sql.NullInt32
	NullInt64   sql.NullInt64
}

func TestTable(t *testing.T) {
	want := &table{
		Name: "foo_bar",
		Columns: []*column{
			{Name: "int8", Type: "TINYINT"},
			{Name: "int16", Type: "SMALLINT"},
			{Name: "int32", Type: "INTEGER"},
			{Name: "int64", Type: "BIGINT"},
			{Name: "uint8", Type: "TINYINT", Unsigned: true},
			{Name: "uint16", Type: "SMALLINT", Unsigned: true},
			{Name: "uint32", Type: "INTEGER", Unsigned: true},
			{Name: "uint64", Type: "BIGINT", Unsigned: true},
			{Name: "string", Type: "VARCHAR", Size: 191},
			{Name: "bool", Type: "TINYINT", Size: 1},
			{Name: "my_int", Type: "BIGINT"},
			{Name: "custom_type", Type: "TIMESTAMP"},
			{Name: "p_int8", Type: "TINYINT"},
			{Name: "p_p_int8", Type: "TINYINT"},
			{Name: "fuga", Type: "INTEGER"},
			{Name: "time", Type: "DATETIME"},
			{Name: "null_time", Type: "DATETIME"},
			{Name: "null_string", Type: "VARCHAR", Size: 191},
			{Name: "null_bool", Type: "TINYINT", Size: 1},
			{Name: "null_byte", Type: "VARBINARY", Size: 767},
			{Name: "null_float64", Type: "DOUBLE"},
			{Name: "null_int16", Type: "SMALLINT"},
			{Name: "null_int32", Type: "INTEGER"},
			{Name: "null_int64", Type: "BIGINT"},
		},
	}
	got, err := newTable(&FooBar{})
	if err != nil {
		t.Fatal(err)
	}
	if diff := cmp.Diff(want, got); diff != "" {
		t.Errorf("table structures are not match (-want/+got):\n%s", diff)
	}
}
