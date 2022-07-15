package myddlmaker

import (
	"database/sql"
	"errors"
	"fmt"
	"reflect"
	"strconv"
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
	Name          string
	Columns       []*column
	PrimaryKey    *PrimaryKey
	Indexes       []*Index
	UniqueIndexes []*UniqueIndex
	ForeignKeys   []*ForeignKey
}

func newTable(s any) (*table, error) {
	val := reflect.ValueOf(s)
	typ := indirect(val.Type())
	if typ.Kind() != reflect.Struct {
		return nil, fmt.Errorf("myddlmaker: expected struct: %s", typ.Kind())
	}

	var tbl table
	tbl.Name = camelToSnake(typ.Name())

	fields := reflect.VisibleFields(typ)
	tbl.Columns = make([]*column, 0, len(fields))
	for _, f := range fields {
		col, err := newColumn(f)
		if err != nil {
			if !errors.Is(err, errSkipColumn) {
				return nil, err
			}
		} else {
			tbl.Columns = append(tbl.Columns, col)
		}
	}

	iface := val.Interface()
	if pk, ok := iface.(primaryKey); ok {
		tbl.PrimaryKey = pk.PrimaryKey()
	}
	if idx, ok := iface.(indexes); ok {
		tbl.Indexes = idx.Indexes()
	}
	if idx, ok := iface.(uniqueIndexes); ok {
		tbl.UniqueIndexes = idx.UniqueIndexes()
	}
	if idx, ok := iface.(foreignKeys); ok {
		tbl.ForeignKeys = idx.ForeignKeys()
	}

	return &tbl, nil
}

type column struct {
	Name     string
	Type     string
	Size     int
	Unsigned bool
	Null     bool
}

var errSkipColumn = errors.New("myddlmaker: skip this column")
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

	typ := indirect(f.Type)
	switch typ.Kind() {
	case reflect.Bool:
		col.Type = "TINYINT"
		col.Size = 1
	case reflect.Int8:
		col.Type = "TINYINT"
	case reflect.Int16:
		col.Type = "SMALLINT"
	case reflect.Int32:
		col.Type = "INTEGER"
	case reflect.Int64:
		col.Type = "BIGINT"
	case reflect.Uint8:
		col.Type = "TINYINT"
		col.Unsigned = true
	case reflect.Uint16:
		col.Type = "SMALLINT"
		col.Unsigned = true
	case reflect.Uint32:
		col.Type = "INTEGER"
		col.Unsigned = true
	case reflect.Uint64:
		col.Type = "BIGINT"
		col.Unsigned = true
	case reflect.Float32:
		col.Type = "FLOAT"
	case reflect.Float64:
		col.Type = "DOUBLE"
	case reflect.String:
		col.Type = "VARCHAR"
		col.Size = 191
	case reflect.Struct:
		switch typ {
		case timeType:
			col.Type = "DATETIME"
		case nullTimeType:
			col.Type = "DATETIME"
		case nullStringType:
			col.Type = "VARCHAR"
			col.Size = 191
		case nullBoolType:
			col.Type = "TINYINT"
			col.Size = 1
		case nullByteType:
			col.Type = "VARBINARY"
			col.Size = 767
		case nullFloat64Type:
			col.Type = "DOUBLE"
		case nullInt16Type:
			col.Type = "SMALLINT"
		case nullInt32Type:
			col.Type = "INTEGER"
		case nullInt64Type:
			col.Type = "BIGINT"
		}
	}

	// parse the tag of the field.
	name, remain, _ := strings.Cut(f.Tag.Get(StructTagName), ",")
	if name == "" {
		name = camelToSnake(f.Name)
	} else if name == "-" {
		return nil, errSkipColumn
	}
	col.Name = name
	for len(remain) > 0 {
		var opt string
		opt, remain, _ = strings.Cut(remain, ",")
		switch {
		case opt == "null":
			col.Null = true
		case strings.HasPrefix(opt, "size="):
			v, err := strconv.ParseInt(opt[len("size="):], 10, 0)
			if err != nil {
				return nil, fmt.Errorf("myddlmaker: failed to parse size param in tag: %w", err)
			}
			col.Size = int(v)
		case strings.HasPrefix(opt, "type="):
			col.Type = opt[len("type="):]
			col.Unsigned = false
			col.Size = 0
		}
	}

	return col, nil
}

func indirect(typ reflect.Type) reflect.Type {
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
