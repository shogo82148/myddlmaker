package myddlmaker

import (
	"fmt"
	"log"
	"reflect"
	"strings"

	"github.com/serenize/snaker"
	"github.com/shogo82148/myddlmaker/dialect"
)

// PrimaryKey describes primary keys of MySQL tables.
type PrimaryKey struct {
	columns []string
}

func AddPrimaryKey(columns ...string) *PrimaryKey {
	return &PrimaryKey{
		columns: columns,
	}
}

// Table is for type assertion
type Table interface {
	Table() string
}

// ForeignKey is for type assertion
type ForeignKey interface {
	ForeignKeys() dialect.ForeignKeys
}

// Index is for type assertion
type Index interface {
	Indexes() dialect.Indexes
}

func (dm *DDLMaker) parse() {
	for _, s := range dm.Structs {
		val := reflect.Indirect(reflect.ValueOf(s))
		rt := val.Type()

		var columns []dialect.Column
		for i := 0; i < rt.NumField(); i++ {
			rtField := rt.Field(i)
			column, err := parseField(rtField, dm.Dialect)
			if err != nil {
				if err == ErrIgnoreField {
					continue
				}
				log.Fatalln("error parse field", err.Error())
			}
			columns = append(columns, column)
		}

		table := parseTable(s, columns, dm.Dialect)
		dm.Tables = append(dm.Tables, table)
	}
}

func parseField(field reflect.StructField, d dialect.Dialect) (dialect.Column, error) {
	tagStr := strings.Replace(field.Tag.Get(TAGPREFIX), " ", "", -1)

	for _, tag := range strings.Split(tagStr, ",") {
		if tag == IGNORETAG {
			return nil, ErrIgnoreField
		}
	}

	var typeName string
	switch {
	case field.Type.PkgPath() != "":
		// ex) time.Time
		pkgName := field.Type.PkgPath()
		if strings.Contains(pkgName, "/") {
			pkgs := strings.Split(pkgName, "/")
			pkgName = pkgs[len(pkgs)-1]
		}
		typeName = fmt.Sprintf("%s.%s", pkgName, field.Type.Name())
	case field.Type.Kind() == reflect.Ptr:
		// pointer type
		typeName = fmt.Sprintf("*%s", field.Type.Elem())
	case field.Type.Kind() == reflect.Slice:
		// slice type
		typeName = fmt.Sprintf("[]%s", field.Type.Elem())
	default:
		typeName = field.Type.Name()
	}

	return newColumn(snaker.CamelToSnake(field.Name), typeName, tagStr, d), nil
}

func parseTable(s interface{}, columns []dialect.Column, d dialect.Dialect) dialect.Table {
	var tableName string
	var primaryKey *PrimaryKey
	var foreignKeys dialect.ForeignKeys
	var indexes dialect.Indexes

	if v, ok := s.(Table); ok {
		tableName = snaker.CamelToSnake(v.Table())
	} else {
		val := reflect.Indirect(reflect.ValueOf(s))
		tableName = snaker.CamelToSnake(val.Type().Name())
	}
	if v, ok := s.(interface{ PrimaryKey() *PrimaryKey }); ok {
		primaryKey = v.PrimaryKey()
	}
	if v, ok := s.(ForeignKey); ok {
		foreignKeys = v.ForeignKeys()
	}
	if v, ok := s.(Index); ok {
		indexes = v.Indexes()
	}

	return newTable(tableName, primaryKey, foreignKeys, columns, indexes, d)
}
