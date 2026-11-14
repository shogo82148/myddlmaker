package myddlmaker

import (
	"fmt"
	"log"
	"slices"
)

type validationError struct {
	errs []string
}

func (e *validationError) Error() string {
	return fmt.Sprintf("myddlmaker: %d error(s) found", len(e.errs))
}

type validator struct {
	SkipValidationFKIndex bool

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
		v.validateColumns(table)
		v.validateIndex(table)
		v.validateIndexName(table)
	}
	v.validateConstraints()
	v.validateForeignKeys()
	v.validateChecks()

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

func (v *validator) validateColumns(table *table) {
	for _, col := range table.columns {
		if col.generated == "" {
			if col.stored {
				v.SaveErrorf("table %q, column %q: stored requires generated", table.name, col.name)
			}
			continue
		}
		if col.autoIncr {
			v.SaveErrorf("table %q, column %q: generated column cannot be auto increment", table.name, col.name)
		}
		if col.def != "" {
			v.SaveErrorf("table %q, column %q: generated column cannot have a default value", table.name, col.name)
		}
		// InnoDB doesn't support virtual generated columns as primary keys.
		if !col.stored && slices.Contains(table.primaryKey.columns, col.name) {
			v.SaveErrorf("table %q, column %q: virtual generated column cannot be a primary key", table.name, col.name)
		}
	}
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

	// CHECK constraints have their own namespace per schema.
	seenChecks := map[string]struct{}{}
	for _, table := range v.tables {
		for _, chk := range table.checks {
			if _, ok := seenChecks[chk.name]; ok {
				v.SaveErrorf("table %q: duplicated name of check constraint: %q", table.name, chk.name)
				continue
			}
			seenChecks[chk.name] = struct{}{}
		}
	}
}

func (v *validator) validateChecks() {
	for _, table := range v.tables {
		for _, chk := range table.checks {
			v.validateCheck(table, chk)
		}
	}
}

// validateCheck validates the columns referenced by the CHECK constraint.
// Only the column names enclosed in back quotes are validated,
// because it is hard to distinguish unquoted column names from keywords and functions.
func (v *validator) validateCheck(table *table, chk *Check) {
	seen := map[string]struct{}{}
	for _, name := range quotedIdentifiers(chk.expr) {
		if _, ok := seen[name]; ok {
			continue
		}
		seen[name] = struct{}{}

		col, ok := v.columnMap[[2]string{table.name, name}]
		if !ok {
			v.SaveErrorf("table %q, check constraint %q: column %q not found", table.name, chk.name, name)
			continue
		}

		// https://dev.mysql.com/doc/refman/8.0/en/create-table-check-constraints.html
		// > AUTO_INCREMENT columns are not permitted.
		if col.autoIncr {
			v.SaveErrorf("table %q, check constraint %q: auto increment column %q cannot be used", table.name, chk.name, name)
		}

		// > Foreign key referential actions (ON UPDATE, ON DELETE) are prohibited on columns used in CHECK constraints.
		// In fact, MySQL rejects only the actions that modify the columns.
		for _, fk := range table.foreignKeys {
			if !modifiesColumns(fk) || !slices.Contains(fk.columns, name) {
				continue
			}
			v.SaveErrorf("table %q, check constraint %q: column %q cannot be used, because it is modified by the referential action of foreign key %q", table.name, chk.name, name, fk.name)
		}
	}
}

// modifiesColumns reports whether the referential actions of fk modify the columns of the child table.
// ON DELETE CASCADE deletes the rows instead of modifying the columns, so MySQL accepts it.
func modifiesColumns(fk *ForeignKey) bool {
	switch fk.onUpdate {
	case ForeignKeyOptionCascade, ForeignKeyOptionSetNull:
		return true
	}
	return fk.onDelete == ForeignKeyOptionSetNull
}

func (v *validator) validateForeignKeys() {
	for _, table := range v.tables {
		for _, fk := range table.foreignKeys {
			v.validateFKColumns(table, fk)
			v.validateFKRef(table, fk)
		}
	}
}

func (v *validator) validateFKColumns(table *table, fk *ForeignKey) {
	passed := true
	for _, col := range fk.columns {
		name := [2]string{table.name, col}
		if _, ok := v.columnMap[name]; !ok {
			v.SaveErrorf("table %q, foreign key %q: column %q not found", table.name, fk.name, col)
			passed = false
			continue
		}
	}

	if !v.SkipValidationFKIndex {
		if passed && !v.hasIndex(table, fk.columns) {
			v.SaveErrorf("table %q, foreign key %q: index required on table %q", table.name, fk.name, table.name)
		}
	}
}

func (v *validator) validateFKRef(table *table, fk *ForeignKey) {
	ref, ok := v.tableMap[fk.table]
	if !ok {
		v.SaveErrorf("table %q, foreign key %q: referenced table %q not found", table.name, fk.name, fk.table)
		return
	}

	passed := true
	for i, col := range fk.references {
		refcol, ok := v.columnMap[[2]string{ref.name, col}]
		if !ok {
			passed = false
			v.SaveErrorf("table %q, foreign key %q: referenced column %q.%q not found", table.name, fk.name, ref.name, col)
			continue
		}

		// type check
		mycol, ok := v.columnMap[[2]string{table.name, fk.columns[i]}]
		if !ok {
			// this error is already reported
			// just ignore it
			continue
		}
		if refcol.typ != mycol.typ || refcol.unsigned != mycol.unsigned {
			v.SaveErrorf("table %q, foreign key %q: column %q and referenced column %q.%q type mismatch", table.name, fk.name, mycol.name, ref.name, col)
		}
		if refcol.charset != mycol.charset {
			v.SaveErrorf("table %q, foreign key %q: column %q and referenced column %q.%q character set mismatch", table.name, fk.name, mycol.name, ref.name, col)
		}
		if refcol.collate != mycol.collate {
			v.SaveErrorf("table %q, foreign key %q: column %q and referenced column %q.%q collate mismatch", table.name, fk.name, mycol.name, ref.name, col)
		}
	}

	if !v.SkipValidationFKIndex {
		if passed && !v.hasIndex(ref, fk.references) {
			v.SaveErrorf("table %q, foreign key %q: index required on table %q", table.name, fk.name, ref.name)
		}
	}
}

func (v *validator) hasIndex(table *table, cols []string) bool {
	if v.hasPrefix(table.primaryKey.columns, cols) {
		return true
	}

	for _, idx := range table.indexes {
		if v.hasPrefix(idx.columns, cols) {
			return true
		}
	}

	for _, idx := range table.uniqueIndexes {
		if v.hasPrefix(idx.columns, cols) {
			return true
		}
	}

	return false
}

func (v *validator) hasPrefix(s []string, prefix []string) bool {
	if len(s) < len(prefix) {
		return false
	}
	for i := range prefix {
		if s[i] != prefix[i] {
			return false
		}
	}
	return true
}
