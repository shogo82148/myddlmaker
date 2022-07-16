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

// Table customize the table name.
type Table interface {
	Table() string
}

type table struct {
	name          string
	columns       []*column
	primaryKey    *PrimaryKey
	indexes       []*Index
	uniqueIndexes []*UniqueIndex
	foreignKeys   []*ForeignKey
}

func newTable(s any) (*table, error) {
	val := reflect.ValueOf(s)
	typ := indirect(val.Type())
	iface := val.Interface()
	if typ.Kind() != reflect.Struct {
		return nil, fmt.Errorf("myddlmaker: expected struct: %s", typ.Kind())
	}

	var tbl table
	if t, ok := iface.(Table); ok {
		tbl.name = t.Table()
	} else {
		tbl.name = camelToSnake(typ.Name())
	}

	fields := reflect.VisibleFields(typ)
	tbl.columns = make([]*column, 0, len(fields))
	for _, f := range fields {
		col, err := newColumn(f)
		if err != nil {
			if !errors.Is(err, errSkipColumn) {
				return nil, err
			}
		} else {
			tbl.columns = append(tbl.columns, col)
		}
	}

	if pk, ok := iface.(primaryKey); ok {
		tbl.primaryKey = pk.PrimaryKey()
	}
	if idx, ok := iface.(indexes); ok {
		tbl.indexes = idx.Indexes()
	}
	if idx, ok := iface.(uniqueIndexes); ok {
		tbl.uniqueIndexes = idx.UniqueIndexes()
	}
	if idx, ok := iface.(foreignKeys); ok {
		tbl.foreignKeys = idx.ForeignKeys()
	}

	return &tbl, nil
}

type column struct {
	name     string
	typ      string
	size     int
	autoIncr bool
	unsigned bool
	null     bool
	def      string
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
		col.typ = "TINYINT"
		col.size = 1
	case reflect.Int8:
		col.typ = "TINYINT"
	case reflect.Int16:
		col.typ = "SMALLINT"
	case reflect.Int32:
		col.typ = "INTEGER"
	case reflect.Int64:
		col.typ = "BIGINT"
	case reflect.Uint8:
		col.typ = "TINYINT"
		col.unsigned = true
	case reflect.Uint16:
		col.typ = "SMALLINT"
		col.unsigned = true
	case reflect.Uint32:
		col.typ = "INTEGER"
		col.unsigned = true
	case reflect.Uint64:
		col.typ = "BIGINT"
		col.unsigned = true
	case reflect.Float32:
		col.typ = "FLOAT"
	case reflect.Float64:
		col.typ = "DOUBLE"
	case reflect.String:
		col.typ = "VARCHAR"
		col.size = 191
	case reflect.Struct:
		switch typ {
		case timeType:
			col.typ = "DATETIME"
		case nullTimeType:
			col.typ = "DATETIME"
		case nullStringType:
			col.typ = "VARCHAR"
			col.size = 191
		case nullBoolType:
			col.typ = "TINYINT"
			col.size = 1
		case nullByteType:
			col.typ = "VARBINARY"
			col.size = 767
		case nullFloat64Type:
			col.typ = "DOUBLE"
		case nullInt16Type:
			col.typ = "SMALLINT"
		case nullInt32Type:
			col.typ = "INTEGER"
		case nullInt64Type:
			col.typ = "BIGINT"
		}
	}

	// parse the tag of the field.
	name, remain, _ := strings.Cut(f.Tag.Get(StructTagName), ",")
	if name == "" {
		name = camelToSnake(f.Name)
	} else if name == "-" {
		return nil, errSkipColumn
	}
	col.name = name
	for len(remain) > 0 {
		var opt string
		opt, remain, _ = strings.Cut(remain, ",")
		switch {
		case opt == "null":
			col.null = true
		case opt == "auto":
			col.autoIncr = true
		case strings.HasPrefix(opt, "size="):
			v, err := strconv.ParseInt(opt[len("size="):], 10, 0)
			if err != nil {
				return nil, fmt.Errorf("myddlmaker: failed to parse size param in tag: %w", err)
			}
			col.size = int(v)
		case strings.HasPrefix(opt, "type="):
			col.typ = opt[len("type="):]
			col.unsigned = false
			col.size = 0
		case strings.HasPrefix(opt, "default="):
			col.def = opt[len("default="):]
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
