package myddlmaker

import (
	"fmt"
	"reflect"
	"strings"
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

func newColumn(f reflect.StructField) (*column, error) {
	col := &column{}

	name, _, _ := strings.Cut(f.Tag.Get(StructTagName), ",")
	if name == "" {
		name = camelToSnake(f.Name)
	}
	col.name = name

	switch direct(f.Type).Kind() {
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
