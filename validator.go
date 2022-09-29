package myddlmaker

import (
	"fmt"
	"log"
)

type validationError struct {
	errs []string
}

func (e *validationError) Error() string {
	return fmt.Sprintf("myddlmaker: %d error(s) found", len(e.errs))
}

type validator struct {
	tables []*table
	errs   []string

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

	for _, table := range v.tables {
		v.validateIndex(table)
		v.validateIndexName(table)
	}
	v.validateConstraints()

	if err := v.Err(); err != nil {
		return err
	}
	return nil
}

func (v *validator) SaveError(msg string) {
	v.errs = append(v.errs, msg)
	log.Println(msg)
}

func (v *validator) SaveErrorf(format string, args ...any) {
	v.errs = append(v.errs, fmt.Sprintf(format, args...))
	log.Printf(format, args...)
}

func (v *validator) Err() error {
	if len(v.errs) == 0 {
		return nil
	}
	return &validationError{
		errs: v.errs,
	}
}

func (v *validator) createTableMap() {
	tables := make(map[string]*table, len(v.tables))
	columns := make(map[[2]string]*column)
	for _, table := range v.tables {
		// validate uniqueness of table names
		if _, ok := tables[table.name]; ok {
			v.SaveErrorf("duplicated name of table: %q", table.name)
			continue
		}

		tables[table.name] = table

		for _, col := range table.columns {
			name := [2]string{table.name, col.name}

			// validate uniqueness of column names
			if _, ok := columns[name]; ok {
				v.SaveErrorf("table %q: duplicated name of column: %q", table.name, col.name)
				continue
			}

			columns[name] = col
		}
	}
	v.tableMap = tables
	v.columnMap = columns
}

func (v *validator) validateIndex(table *table) {
	// check existence of the column in the primary key
	for _, col := range table.primaryKey.columns {
		name := [2]string{table.name, col}
		if _, ok := v.columnMap[name]; !ok {
			v.SaveErrorf("table %q, primary key: column %q not found", table.name, col)
			continue
		}
	}

	for _, idx := range table.indexes {
		// check existence of the column in the index
		for _, col := range idx.columns {
			name := [2]string{table.name, col}
			if _, ok := v.columnMap[name]; !ok {
				v.SaveErrorf("table %q, index %q: column %q not found", table.name, idx.name, col)
				continue
			}
		}
	}

	for _, idx := range table.uniqueIndexes {
		// check existence of the column in the unique index
		for _, col := range idx.columns {
			name := [2]string{table.name, col}
			if _, ok := v.columnMap[name]; !ok {
				v.SaveErrorf("table %q, unique index %q: column %q not found", table.name, idx.name, col)
				continue
			}
		}
	}
}

func (v *validator) validateIndexName(table *table) {
	seen := map[string]struct{}{}

	for _, idx := range table.indexes {
		if _, ok := seen[idx.name]; ok {
			v.SaveErrorf("table %q: duplicated name of index: %q", table.name, idx.name)
			continue
		}
		seen[idx.name] = struct{}{}
	}

	for _, idx := range table.uniqueIndexes {
		if _, ok := seen[idx.name]; ok {
			v.SaveErrorf("table %q: duplicated name of index: %q", table.name, idx.name)
			continue
		}
		seen[idx.name] = struct{}{}
	}

	for _, idx := range table.fullTextIndexes {
		if _, ok := seen[idx.name]; ok {
			v.SaveErrorf("table %q: duplicated name of index: %q", table.name, idx.name)
			continue
		}
		seen[idx.name] = struct{}{}
	}

	for _, idx := range table.spatialIndexes {
		if _, ok := seen[idx.name]; ok {
			v.SaveErrorf("table %q: duplicated name of index: %q", table.name, idx.name)
			continue
		}
		seen[idx.name] = struct{}{}
	}
}

func (v *validator) validateConstraints() {
	seen := map[string]struct{}{}

	for _, table := range v.tables {
		for _, fk := range table.foreignKeys {
			if _, ok := seen[fk.name]; ok {
				v.SaveErrorf("table %q: duplicated name of foreign key constraint: %q", table.name, fk.name)
				continue
			}
			seen[fk.name] = struct{}{}
		}
	}
}
