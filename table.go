package myddlmaker

import "reflect"

type table struct {
	name string
}

func newTable(s any) *table {
	val := reflect.Indirect(reflect.ValueOf(s))
	typ := val.Type()
	return &table{
		name: camelToSnake(typ.Name()),
	}
}
