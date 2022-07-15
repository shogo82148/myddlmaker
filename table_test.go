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
		name: "foo_bar",
		columns: []*column{
			{name: "int8", typ: "TINYINT"},
			{name: "int16", typ: "SMALLINT"},
			{name: "int32", typ: "INTEGER"},
			{name: "int64", typ: "BIGINT"},
			{name: "uint8", typ: "TINYINT", unsigned: true},
			{name: "uint16", typ: "SMALLINT", unsigned: true},
			{name: "uint32", typ: "INTEGER", unsigned: true},
			{name: "uint64", typ: "BIGINT", unsigned: true},
			{name: "string", typ: "VARCHAR", size: 191},
			{name: "bool", typ: "TINYINT", size: 1},
			{name: "my_int", typ: "BIGINT"},
			{name: "custom_type", typ: "TIMESTAMP"},
			{name: "p_int8", typ: "TINYINT"},
			{name: "p_p_int8", typ: "TINYINT"},
			{name: "fuga", typ: "INTEGER"},
			{name: "time", typ: "DATETIME"},
			{name: "null_time", typ: "DATETIME"},
			{name: "null_string", typ: "VARCHAR", size: 191},
			{name: "null_bool", typ: "TINYINT", size: 1},
			{name: "null_byte", typ: "VARBINARY", size: 767},
			{name: "null_float64", typ: "DOUBLE"},
			{name: "null_int16", typ: "SMALLINT"},
			{name: "null_int32", typ: "INTEGER"},
			{name: "null_int64", typ: "BIGINT"},
		},
	}
	got, err := newTable(&FooBar{})
	if err != nil {
		t.Fatal(err)
	}
	opt := cmp.AllowUnexported(table{}, column{}, PrimaryKey{}, Index{}, UniqueIndex{}, ForeignKey{})
	if diff := cmp.Diff(want, got, opt); diff != "" {
		t.Errorf("table structures are not match (-want/+got):\n%s", diff)
	}
}
