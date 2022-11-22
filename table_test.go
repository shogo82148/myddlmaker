package myddlmaker

import (
	"database/sql"
	"encoding/json"
	"testing"
	"time"

	"github.com/google/go-cmp/cmp"
	"github.com/google/go-cmp/cmp/cmpopts"
)

type myInt int64
type customType struct{}
type nullUint32 struct {
	u32   uint32
	valid bool
}

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

	// binary types
	Bytes     []byte
	ByteArray [4]byte
	JSONValue json.RawMessage

	// custom type
	MyInt      myInt
	CustomType customType `ddl:",type=TIMESTAMP"`
	Decimal    float64    `ddl:",type=DECIMAL"`
	Numeric    float64    `ddl:",type=NUMERIC"`

	// custom type that contains comma
	DecimalWithPrecisionAndScale float64 `ddl:",type=DECIMAL(9,6)"`
	NumericWithPrecisionAndScale float64 `ddl:",type=NUMERIC(9,6)"`
	DecimalWithPrecision         float64 `ddl:",type=DECIMAL(9)"`
	NumericWithPrecision         float64 `ddl:",type=NUMERIC(9)"`

	// custom type that is unsigned integer
	NullUint32 nullUint32 `ddl:",type=INTEGER,unsigned,null"`

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

	// Auto Increment
	Auto int64 `ddl:",auto"`

	// Default Value
	DefaultValue int64 `ddl:",default=123"`
}

func TestTable(t *testing.T) {
	want := &table{
		name:    "foo_bar",
		rawName: "FooBar",
		columns: []*column{
			{name: "int8", rawName: "Int8", typ: "TINYINT"},
			{name: "int16", rawName: "Int16", typ: "SMALLINT"},
			{name: "int32", rawName: "Int32", typ: "INTEGER"},
			{name: "int64", rawName: "Int64", typ: "BIGINT"},
			{name: "uint8", rawName: "Uint8", typ: "TINYINT", unsigned: true},
			{name: "uint16", rawName: "Uint16", typ: "SMALLINT", unsigned: true},
			{name: "uint32", rawName: "Uint32", typ: "INTEGER", unsigned: true},
			{name: "uint64", rawName: "Uint64", typ: "BIGINT", unsigned: true},
			{name: "string", rawName: "String", typ: "VARCHAR", size: 191},
			{name: "bool", rawName: "Bool", typ: "TINYINT", size: 1},
			{name: "bytes", rawName: "Bytes", typ: "VARBINARY", size: 767},
			{name: "byte_array", rawName: "ByteArray", typ: "BINARY", size: 4},
			{name: "json_value", rawName: "JSONValue", typ: "JSON"},
			{name: "my_int", rawName: "MyInt", typ: "BIGINT"},
			{name: "custom_type", rawName: "CustomType", typ: "TIMESTAMP"},
			{name: "decimal", rawName: "Decimal", typ: "DECIMAL"},
			{name: "numeric", rawName: "Numeric", typ: "NUMERIC"},
			{name: "decimal_with_precision_and_scale", rawName: "DecimalWithPrecisionAndScale", typ: "DECIMAL(9,6)"},
			{name: "numeric_with_precision_and_scale", rawName: "NumericWithPrecisionAndScale", typ: "NUMERIC(9,6)"},
			{name: "decimal_with_precision", rawName: "DecimalWithPrecision", typ: "DECIMAL(9)"},
			{name: "numeric_with_precision", rawName: "NumericWithPrecision", typ: "NUMERIC(9)"},
			{name: "null_uint32", rawName: "NullUint32", typ: "INTEGER", unsigned: true, null: true},
			{name: "p_int8", rawName: "PInt8", typ: "TINYINT"},
			{name: "p_p_int8", rawName: "PPInt8", typ: "TINYINT"},
			{name: "fuga", rawName: "Hoge", typ: "INTEGER"},
			{name: "time", rawName: "Time", typ: "DATETIME", size: 6},
			{name: "null_time", rawName: "NullTime", typ: "DATETIME", size: 6},
			{name: "null_string", rawName: "NullString", typ: "VARCHAR", size: 191},
			{name: "null_bool", rawName: "NullBool", typ: "TINYINT", size: 1},
			{name: "null_byte", rawName: "NullByte", typ: "TINYINT", unsigned: true},
			{name: "null_float64", rawName: "NullFloat64", typ: "DOUBLE"},
			{name: "null_int16", rawName: "NullInt16", typ: "SMALLINT"},
			{name: "null_int32", rawName: "NullInt32", typ: "INTEGER"},
			{name: "null_int64", rawName: "NullInt64", typ: "BIGINT"},
			{name: "auto", rawName: "Auto", typ: "BIGINT", autoIncr: true},
			{name: "default_value", rawName: "DefaultValue", typ: "BIGINT", def: "123"},
		},
	}
	got, err := newTable(&FooBar{})
	if err != nil {
		t.Fatal(err)
	}
	opt1 := cmp.AllowUnexported(table{}, column{}, PrimaryKey{}, Index{}, UniqueIndex{}, ForeignKey{})
	opt2 := cmpopts.IgnoreFields(column{}, "rawType")
	if diff := cmp.Diff(want, got, opt1, opt2); diff != "" {
		t.Errorf("table structures are not match (-want/+got):\n%s", diff)
	}
}

func TestTable_UnknownType(t *testing.T) {
	type FooBar struct {
		// The DDL maker doesn't know about customType.
		// It causes some errors.
		Foo customType
	}

	_, err := newTable(&FooBar{})
	if err == nil {
		t.Error("want some errors, got nil")
	}
}

func TestCutComma(t *testing.T) {
	tests := []struct {
		in     string
		before string
		after  string
		found  bool
	}{
		{
			in:     "6,type=INTEGER",
			before: "6",
			after:  "type=INTEGER",
			found:  true,
		},
		{
			in:     "6,type=INTEGER,auto",
			before: "6",
			after:  "type=INTEGER,auto",
			found:  true,
		},
		{
			in:     "DECIMAL(9,6)",
			before: "DECIMAL(9,6)",
			after:  "",
			found:  false,
		},
		{
			in:     "DECIMAL(9,6),null",
			before: "DECIMAL(9,6)",
			after:  "null",
			found:  true,
		},
		{
			in:     "DECIMAL((9),(6))",
			before: "DECIMAL((9),(6))",
			after:  "",
			found:  false,
		},
		{
			in:     "DECIMAL((9),(6)),null",
			before: "DECIMAL((9),(6))",
			after:  "null",
			found:  true,
		},
		{
			in:     "null",
			before: "null",
			after:  "",
			found:  false,
		},
		{
			in:     "DECIMAL(9,6),null",
			before: "DECIMAL(9,6)",
			after:  "null",
			found:  true,
		},
		{
			in:     "DECIMAL(9,6)),null",
			before: "DECIMAL(9,6))",
			after:  "null",
			found:  true,
		},
	}

	for i, tt := range tests {
		before, after, found := cutComma(tt.in)
		if before != tt.before {
			t.Errorf("%d: unexpected before: got %q, want %q", i, before, tt.before)
		}
		if after != tt.after {
			t.Errorf("%d: unexpected after: got %q, want %q", i, after, tt.after)
		}
		if found != tt.found {
			t.Errorf("%d: unexpected found: got %t, want %t", i, found, tt.found)
		}
	}
}
