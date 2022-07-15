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

type Index struct {
	name    string
	columns []string
}

func NewIndex(name string, col ...string) *Index {
	return &Index{
		name:    name,
		columns: col,
	}
}

type UniqueIndex struct {
	name    string
	columns []string
}

func NewUniqueIndex(name string, col ...string) *UniqueIndex {
	return &UniqueIndex{
		name:    name,
		columns: col,
	}
}

type ForeignKey struct {
	name       string
	columns    []string
	table      string
	references []string
	onUpdate   ForeignKeyOption
	onDelete   ForeignKeyOption
}

type ForeignKeyOption string

const (
	ForeignKeyOptionCascade    ForeignKeyOption = "CASCADE"
	ForeignKeyOptionSetNull    ForeignKeyOption = "SET NULL"
	ForeignKeyOptionRestrict   ForeignKeyOption = "RESTRICT"
	ForeignKeyOptionNoAction   ForeignKeyOption = "NO ACTION"
	ForeignKeyOptionSetDefault ForeignKeyOption = "SET DEFAULT"
)

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

func (fk *ForeignKey) OnUpdate(opt ForeignKeyOption) *ForeignKey {
	key := *fk // shallow copy
	key.onUpdate = opt
	return &key
}

func (fk *ForeignKey) OnDelete(opt ForeignKeyOption) *ForeignKey {
	key := *fk // shallow copy
	key.onDelete = opt
	return &key
}
