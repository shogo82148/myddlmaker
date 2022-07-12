package myddlmaker

type indexes interface {
	Indexes() []Index
}

type Index interface {
	// dummy method disallow implementation of this interface.
	myddlmaker()
}

type index struct {
	Name    string
	Columns []string
}

func (*index) myddlmaker() {}

func NewIndex(name string, col ...string) Index {
	return &index{
		Name:    name,
		Columns: col,
	}
}
