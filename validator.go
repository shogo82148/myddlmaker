package myddlmaker

import (
	"fmt"
	"log"
)

type validator struct {
	tables   []*table
	errCount int

	// key: table name
	// value: table
	tableMap map[string]*table

	// key: table name, column name
	// value: column
	columnMap map[[2]string]*column
}

func newValidator(tables []*table) *validator {
	return &validator{
		tables: tables,
	}
}

func (v *validator) Validate() error {
	v.createTableMap()
	return nil
}

func (v *validator) SaveError(msg string) {
	v.errCount++
	log.Println(msg)
}

func (v *validator) SaveErrorf(format string, args ...any) {
	v.errCount++
	log.Printf(format, args...)
}

func (v *validator) Err() error {
	if v.errCount == 0 {
		return nil
	}
	return fmt.Errorf("myddlmaker: %d error(s) found", v.errCount)
}

func (v *validator) createTableMap() {
	tables := make(map[string]*table, len(v.tables))
	columns := make(map[[2]string]*column)
	for _, table := range v.tables {
		// validate uniqueness of table names
		if _, ok := tables[table.name]; ok {
			v.SaveErrorf("table %q already exists", table.name)
			continue
		}

		tables[table.name] = table

		for _, col := range table.columns {
			name := [2]string{table.name, col.name}

			// // validate uniqueness of column names
			if _, ok := columns[name]; ok {
				v.SaveErrorf("column %q.%q already exists", table.name, col.name)
				continue
			}

			columns[name] = col
		}
	}
	v.tableMap = tables
	v.columnMap = columns
}
