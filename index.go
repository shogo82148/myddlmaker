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
	name      string
	columns   []string
	comment   string
	invisible bool
}

func NewIndex(name string, col ...string) *Index {
	return &Index{
		name:    name,
		columns: col,
	}
}

func (idx *Index) Comment(comment string) *Index {
	tmp := *idx // shallow copy
	tmp.comment = comment
	return &tmp
}

func (idx *Index) Invisible() *Index {
	tmp := *idx // shallow copy
	tmp.invisible = true
	return &tmp
}

type UniqueIndex struct {
	name      string
	columns   []string
	comment   string
	invisible bool
}

func NewUniqueIndex(name string, col ...string) *UniqueIndex {
	return &UniqueIndex{
		name:    name,
		columns: col,
	}
}

func (idx *UniqueIndex) Comment(comment string) *UniqueIndex {
	tmp := *idx // shallow copy
	tmp.comment = comment
	return &tmp
}

func (idx *UniqueIndex) Invisible() *UniqueIndex {
	tmp := *idx // shallow copy
	tmp.invisible = true
	return &tmp
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

type fullTextIndexes interface {
	FullTextIndexes() []*FullTextIndex
}

// https://dev.mysql.com/doc/refman/8.0/en/innodb-fulltext-index.html
type FullTextIndex struct {
	name    string
	column  string
	comment string
	parser  string
}

func NewFullTextIndex(name string, column string) *FullTextIndex {
	return &FullTextIndex{
		name:   name,
		column: column,
	}
}

func (idx *FullTextIndex) Comment(comment string) *FullTextIndex {
	tmp := *idx // shallow copy
	tmp.comment = comment
	return &tmp
}

func (idx *FullTextIndex) WithParser(parser string) *FullTextIndex {
	tmp := *idx // shallow copy
	tmp.parser = parser
	return &tmp
}
