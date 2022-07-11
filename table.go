package myddlmaker

import (
	"database/sql"
	"fmt"
	"reflect"
	"strings"
	"time"
)

var (
	// StructTagName is the key name of the tag string.
	StructTagName = "ddl"

	// IgnoreName is the string that myddlmaker ignores.
	IgnoreName = "-"
)

type table struct {
	name    string
	columns []*column
}

func newTable(s any) (*table, error) {
	val := reflect.Indirect(reflect.ValueOf(s))
	typ := direct(val.Type())
	if typ.Kind() != reflect.Struct {
		return nil, fmt.Errorf("myddlmaker: expected struct: %s", typ.Kind())
	}

	fields := reflect.VisibleFields(typ)
	columns := make([]*column, 0, len(fields))
	for _, f := range fields {
		col, err := newColumn(f)
		if err != nil {
			return nil, err
		}
		columns = append(columns, col)
	}

	return &table{
		name:    camelToSnake(typ.Name()),
		columns: columns,
	}, nil
}

type column struct {
	name string
	typ  string
}

var timeType = reflect.TypeOf(time.Time{})
var nullTimeType = reflect.TypeOf(sql.NullTime{})
var nullStringType = reflect.TypeOf(sql.NullString{})
var nullBoolType = reflect.TypeOf(sql.NullBool{})
var nullByteType = reflect.TypeOf(sql.NullByte{})
var nullFloat64Type = reflect.TypeOf(sql.NullFloat64{})
var nullInt16Type = reflect.TypeOf(sql.NullInt16{})
var nullInt32Type = reflect.TypeOf(sql.NullInt32{})
var nullInt64Type = reflect.TypeOf(sql.NullInt64{})

func newColumn(f reflect.StructField) (*column, error) {
	col := &column{}

	name, _, _ := strings.Cut(f.Tag.Get(StructTagName), ",")
	if name == "" {
		name = camelToSnake(f.Name)
	}
	col.name = name

	typ := direct(f.Type)
	switch typ.Kind() {
	case reflect.Bool:
		col.typ = "TINYINT(1)"
	case reflect.Int8:
		col.typ = "TINYINT"
	case reflect.Int16:
		col.typ = "SMALLINT"
	case reflect.Int32:
		col.typ = "INTEGER"
	case reflect.Int64:
		col.typ = "BIGINT"
	case reflect.Uint8:
		col.typ = "TINYINT unsigned"
	case reflect.Uint16:
		col.typ = "SMALLINT unsigned"
	case reflect.Uint32:
		col.typ = "INTEGER unsigned"
	case reflect.Uint64:
		col.typ = "BIGINT unsigned"
	case reflect.Float32:
		col.typ = "FLOAT"
	case reflect.Float64:
		col.typ = "DOUBLE"
	case reflect.String:
		col.typ = "VARCHAR(191)"
	case reflect.Struct:
		switch typ {
		case timeType:
			col.typ = "DATETIME"
		case nullTimeType:
			col.typ = "DATETIME"
		case nullStringType:
			col.typ = "VARCHAR(191)"
		case nullBoolType:
			col.typ = "TINYINT(1)"
		case nullByteType:
			col.typ = "VARBINARY(767)"
		case nullFloat64Type:
			col.typ = "DOUBLE"
		case nullInt16Type:
			col.typ = "SMALLINT"
		case nullInt32Type:
			col.typ = "INTEGER"
		case nullInt64Type:
			col.typ = "BIGINT"
		default:
			return nil, fmt.Errorf("myddlmaker: unknown type: %v", typ)
		}
	default:
		return nil, fmt.Errorf("myddlmaker: unknown type: %v", typ)
	}
	return col, nil
}

func direct(typ reflect.Type) reflect.Type {
	seen := map[reflect.Type]struct{}{
		typ: {},
	}
	for typ.Kind() == reflect.Pointer {
		elem := typ.Elem()
		if _, ok := seen[elem]; ok {
			return typ
		}
		typ = elem
		seen[typ] = struct{}{}
	}
	return typ
}
