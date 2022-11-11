package myddlmaker

import (
	"database/sql"
	"encoding/json"
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
// Implement the Table interface to customize the table name.
type Table interface {
	Table() string
}

type table struct {
	name            string
	rawName         string
	columns         []*column
	primaryKey      *PrimaryKey
	indexes         []*Index
	uniqueIndexes   []*UniqueIndex
	foreignKeys     []*ForeignKey
	fullTextIndexes []*FullTextIndex
	spatialIndexes  []*SpatialIndex
}

func newTable(s any) (*table, error) {
	val := reflect.ValueOf(s)
	typ := indirect(val.Type())
	iface := val.Interface()
	if typ.Kind() != reflect.Struct {
		return nil, fmt.Errorf("myddlmaker: expected struct: %s", typ.Kind())
	}

	var tbl table
	tbl.rawName = typ.Name()
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
	if idx, ok := iface.(fullTextIndexes); ok {
		tbl.fullTextIndexes = idx.FullTextIndexes()
	}
	if idx, ok := iface.(spatialIndex); ok {
		tbl.spatialIndexes = idx.SpatialIndexes()
	}

	return &tbl, nil
}

type column struct {
	// name is the name in SQL queries
	name string

	// rawName is the name in Go codes.
	rawName string

	// typ is the type name in SQL queries
	typ string

	// rawType is the type name in Go codes.
	rawType reflect.Type

	size int

	// autoIncr marks the column an auto increment column.
	autoIncr bool

	unsigned bool

	// invisible marks invisible columns.
	// https://dev.mysql.com/doc/refman/8.0/en/invisible-columns.html
	invisible bool

	// null enables to accept NULL values.
	null bool

	// def is the default value of the column.
	def string

	// comment is a comment
	comment string

	// charset is character set
	charset string

	// collate is collation name
	collate string

	// srid is the id of spatial reference systems
	srid int
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
var jsonRawMessageType = reflect.TypeOf(json.RawMessage{})

func newColumn(f reflect.StructField) (*column, error) {
	var invalidType bool

	typ := indirect(f.Type)
	col := &column{
		rawType: typ,
	}
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
	case reflect.Slice:
		if typ == jsonRawMessageType {
			col.typ = "JSON"
		} else if typ.Elem().Kind() == reflect.Uint8 {
			col.typ = "VARBINARY"
			col.size = 767
		} else {
			invalidType = true
		}
	case reflect.Array:
		if typ.Elem().Kind() == reflect.Uint8 {
			col.typ = "BINARY"
			col.size = typ.Len()
		} else {
			invalidType = true
		}
	case reflect.Struct:
		switch typ {
		case timeType:
			col.typ = "DATETIME"
			col.size = 6
		case nullTimeType:
			col.typ = "DATETIME"
			col.size = 6
		case nullStringType:
			col.typ = "VARCHAR"
			col.size = 191
		case nullBoolType:
			col.typ = "TINYINT"
			col.size = 1
		case nullByteType:
			col.typ = "TINYINT"
			col.unsigned = true
		case nullFloat64Type:
			col.typ = "DOUBLE"
		case nullInt16Type:
			col.typ = "SMALLINT"
		case nullInt32Type:
			col.typ = "INTEGER"
		case nullInt64Type:
			col.typ = "BIGINT"
		default:
			invalidType = true
		}
	default:
		invalidType = true
	}

	// parse the tag of the field.
	col.rawName = f.Name
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
		switch opt {
		case "null":
			col.null = true
		case "auto":
			col.autoIncr = true
		case "invisible":
			col.invisible = true
		default:
			name, val, _ := strings.Cut(opt, "=")
			switch name {
			case "size":
				v, err := strconv.ParseInt(val, 10, 0)
				if err != nil {
					return nil, fmt.Errorf("myddlmaker: failed to parse size param in tag: %w", err)
				}
				col.size = int(v)
			case "srid":
				v, err := strconv.ParseInt(val, 10, 0)
				if err != nil {
					return nil, fmt.Errorf("myddlmaker: failed to parse srid param in tag: %w", err)
				}
				col.srid = int(v)
			case "type":
				if strings.HasPrefix(val, "DECIMAL(") && !strings.HasSuffix(val, ")") {
					var comb string
					comb, remain, _ = strings.Cut(remain, ",")
					val = val + "," + comb
				}
				if strings.HasPrefix(val, "NUMERIC(") && !strings.HasSuffix(val, ")") {
					var comb string
					comb, remain, _ = strings.Cut(remain, ",")
					val = val + "," + comb
				}
				col.typ = val
				col.unsigned = false
				col.size = 0
				invalidType = false
			case "default":
				col.def = val
			case "charset":
				col.charset = val
			case "collate":
				col.collate = val
			case "comment":
				col.comment = val
			}
		}
	}

	if invalidType {
		return nil, fmt.Errorf("myddlmaker: unknown type: %s", typ.String())
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
