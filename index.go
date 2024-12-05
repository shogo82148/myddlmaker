package myddlmaker

type indexes interface {
	Indexes() []*Index
}

type uniqueIndexes interface {
	UniqueIndexes() []*UniqueIndex
}

type foreignKeys interface {
	ForeignKeys() []*ForeignKey
}

// Index is an index of a table.
// Implement the Indexes method to define the indexes.
//
//	func (*User) Indexes() []*myddlmaker.Index {
//	    return []*myddlmaker.Index{
//	        // INDEX `idx_name` (`name`)
//	        myddlmaker.NewIndex("idx_name", "name"),
//	    }
//	}
type Index struct {
	name      string
	columns   []string
	comment   string
	invisible bool
	order     map[string]string // key: column, value: ASC or DESC
}

// NewIndex returns a new index.
func NewIndex(name string, col ...string) *Index {
	if name == "" {
		panic("name is missing")
	}
	if len(col) == 0 {
		panic("col is missing")
	}
	order := make(map[string]string, len(col))

	return &Index{
		name:    name,
		columns: col,
		order:   order,
	}
}

// Comment returns a copy of idx with the comment.
func (idx *Index) Comment(comment string) *Index {
	tmp := *idx // shallow copy
	tmp.comment = comment
	return &tmp
}

// Invisible returns a copy of idx, but it is invisible from MySQL planner.
func (idx *Index) Invisible() *Index {
	tmp := *idx // shallow copy
	tmp.invisible = true
	return &tmp
}

// ASC returns a copy of idx with the ASC option.
// If you set both ASC and DESC, the last one will be used.
// If you specify the non-existent column, panic will be raised.
func (idx *Index) ASC(column string) *Index {
	tmp := *idx // shallow copy
	tmp.order = make(map[string]string, len(tmp.columns))
	for k, v := range idx.order {
		tmp.order[k] = v
	}
	if !containsColumn(idx, column) {
		panic("column is missing")
	}
	tmp.order[column] = "ASC"
	return &tmp
}

// DESC returns a copy of idx with the DESC option.
// If you set both ASC and DESC, the last one will be used.
// If you specify the non-existent column, panic will be raised.
func (idx *Index) DESC(column string) *Index {
	tmp := *idx // shallow copy
	tmp.order = make(map[string]string, len(tmp.columns))
	for k, v := range idx.order {
		tmp.order[k] = v
	}
	if !containsColumn(idx, column) {
		panic("column is missing")
	}
	tmp.order[column] = "DESC"
	return &tmp
}

// containsColumn returns true if the index contains the column.
// slices.Contains is supported in Go 1.21. So, we can't use it.
func containsColumn(idx *Index, column string) bool {
	for _, c := range idx.columns {
		if c == column {
			return true
		}
	}
	return false
}

// UniqueIndex is a unique index of a table.
// Implement the UniqueIndexes method to define the unique indexes.
//
//	func (*User) UniqueIndexes() []*myddlmaker.Index {
//		return []*myddlmaker.Index{
//			// UNIQUE INDEX `idx_name` (`name`)
//			myddlmaker.NewUniqueIndex("idx_name", "name"),
//		}
//	}
type UniqueIndex struct {
	name      string
	columns   []string
	comment   string
	invisible bool
}

// NewUniqueIndex returns a new unique index.
func NewUniqueIndex(name string, col ...string) *UniqueIndex {
	if name == "" {
		panic("name is missing")
	}
	if len(col) == 0 {
		panic("col is missing")
	}
	return &UniqueIndex{
		name:    name,
		columns: col,
	}
}

// Comment returns a copy of idx with the comment.
func (idx *UniqueIndex) Comment(comment string) *UniqueIndex {
	tmp := *idx // shallow copy
	tmp.comment = comment
	return &tmp
}

// Invisible returns a copy of idx, but it is invisible from MySQL planner.
func (idx *UniqueIndex) Invisible() *UniqueIndex {
	tmp := *idx // shallow copy
	tmp.invisible = true
	return &tmp
}

// ForeignKey is a foreign key constraint.
// Implement the ForeignKeys method to define the foreign key constraints.
//
//	func (*User) ForeignKeys() []*myddlmaker.ForeignKey {
//		return []*myddlmaker.ForeignKey{
//			// CONSTRAINT `name_of_constraint`
//			//     FOREIGN KEY (`column1`, `column2`)
//			//     REFERENCES `another_table` (`id1`, `id2`)
//			myddlmaker.NewForeignKey(
//				"name_of_constraint",
//				[]string{"column1", "column2"},
//				"another_table",
//				[]string{"id1", "id2"},
//			),
//		}
//	}
type ForeignKey struct {
	name       string
	columns    []string
	table      string
	references []string
	onUpdate   ForeignKeyOption
	onDelete   ForeignKeyOption
}

// ForeignKeyOption is an option of a referential action.
type ForeignKeyOption string

const (
	// ForeignKeyOptionCascade deletes or updates the row from the parent table
	// and automatically delete or update the matching rows in the child table.
	ForeignKeyOptionCascade ForeignKeyOption = "CASCADE"

	// ForeignKeyOptionSetNull deletes or updates the row from the parent table
	// and set the foreign key column or columns in the child table to NULL.
	ForeignKeyOptionSetNull ForeignKeyOption = "SET NULL"

	// ForeignKeyOptionRestrict rejects the delete or update operation for the parent table.
	ForeignKeyOptionRestrict ForeignKeyOption = "RESTRICT"

	// ForeignKeyOptionNoAction is same as ForeignKeyOptionRestrict in MySQL.
	// However, in some database system, it maybe not.
	// So we should not use this for better compatibility.
	// ForeignKeyOptionNoAction ForeignKeyOption = "NO ACTION"

	// ForeignKeyOptionSetDefault can't be used with InnoDB.
	// In many cases, there is no need to use it.
	// ForeignKeyOptionSetDefault ForeignKeyOption = "SET DEFAULT"
)

// NewForeignKey returns a new foreign key constraint.
func NewForeignKey(name string, columns []string, table string, references []string) *ForeignKey {
	if name == "" {
		panic("name is missing")
	}
	if table == "" {
		panic("table is missing")
	}
	if len(columns) == 0 {
		panic("columns is missing")
	}
	if len(references) == 0 {
		panic("references is missing")
	}
	if len(columns) != len(references) {
		panic("columns and references must have same length")
	}
	return &ForeignKey{
		name:       name,
		columns:    columns,
		table:      table,
		references: references,
	}
}

// OnUpdate returns a copy of fk with the referential action option opt
// specified by ON UPDATE cause.
func (fk *ForeignKey) OnUpdate(opt ForeignKeyOption) *ForeignKey {
	key := *fk // shallow copy
	key.onUpdate = opt
	return &key
}

// OnDelete returns a copy of fk with the referential action option opt
// specified by ON DELETE cause.
func (fk *ForeignKey) OnDelete(opt ForeignKeyOption) *ForeignKey {
	key := *fk // shallow copy
	key.onDelete = opt
	return &key
}

type fullTextIndexes interface {
	FullTextIndexes() []*FullTextIndex
}

// FullTextIndex is a full text index.
// https://dev.mysql.com/doc/refman/8.0/en/innodb-fulltext-index.html
// Implement the `FullTextIndexes` method to define the full-text indexes.
//
//	func (*User) FullTextIndexes() []*myddlmaker.FullTextIndex {
//		return []*myddlmaker.FullTextIndex{
//			// FULLTEXT INDEX `idx_name` (`name`)
//			myddlmaker.NewFullTextIndex("idx_name", "name"),
//		}
//	}
type FullTextIndex struct {
	name      string
	column    string
	invisible bool
	comment   string
	parser    string
}

// NewFullTextIndex returns a new full text index.
func NewFullTextIndex(name string, column string) *FullTextIndex {
	if name == "" {
		panic("name is missing")
	}
	if column == "" {
		panic("column is missing")
	}
	return &FullTextIndex{
		name:   name,
		column: column,
	}
}

// Invisible returns a copy of idx, but it is invisible from MySQL planner.
func (idx *FullTextIndex) Invisible() *FullTextIndex {
	tmp := *idx // shallow copy
	tmp.invisible = true
	return &tmp
}

// Comment returns a copy of idx with the comment.
func (idx *FullTextIndex) Comment(comment string) *FullTextIndex {
	tmp := *idx // shallow copy
	tmp.comment = comment
	return &tmp
}

// WithParser returns a copy of idx with the full-text plugin.
func (idx *FullTextIndex) WithParser(parser string) *FullTextIndex {
	tmp := *idx // shallow copy
	tmp.parser = parser
	return &tmp
}

type spatialIndex interface {
	SpatialIndexes() []*SpatialIndex
}

// SpatialIndex is a spatial index.
// Implement the SpatialIndexes method to define the spatial indexes.
//
//	func (*User) SpatialIndexes() []*myddlmaker.SpatialIndex {
//		return []*myddlmaker.SpatialIndex{
//			// SPATIAL INDEX `idx_name` (`name`)
//			myddlmaker.NewSpatialIndex("idx_name", "name"),
//		}
//	}
type SpatialIndex struct {
	name      string
	column    string
	invisible bool
	comment   string
}

// NewSpatialIndex returns a new spatial index.
func NewSpatialIndex(name string, column string) *SpatialIndex {
	if name == "" {
		panic("name is missing")
	}
	if column == "" {
		panic("column is missing")
	}
	return &SpatialIndex{
		name:   name,
		column: column,
	}
}

// Invisible returns a copy of idx, but it is invisible from MySQL planner.
func (idx *SpatialIndex) Invisible() *SpatialIndex {
	tmp := *idx // shallow copy
	tmp.invisible = true
	return &tmp
}

// Comment returns a copy of idx with the comment.
func (idx *SpatialIndex) Comment(comment string) *SpatialIndex {
	tmp := *idx // shallow copy
	tmp.comment = comment
	return &tmp
}
