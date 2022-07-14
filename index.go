package myddlmaker

type indexes interface {
	Indexes() []*Index
}

type uniqueIndexes interface {
	UniqueIndexes() []*UniqueIndex
}

type Index struct {
	Name    string
	Columns []string
}

func NewIndex(name string, col ...string) *Index {
	return &Index{
		Name:    name,
		Columns: col,
	}
}

type UniqueIndex struct {
	Name    string
	Columns []string
}

func NewUniqueIndex(name string, col ...string) *UniqueIndex {
	return &UniqueIndex{
		Name:    name,
		Columns: col,
	}
}
