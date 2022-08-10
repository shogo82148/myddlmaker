package schema

import (
	"github.com/shogo82148/myddlmaker"
)

type Foo1 struct {
	ID int32
}

func (*Foo1) PrimaryKey() *myddlmaker.PrimaryKey {
	return myddlmaker.NewPrimaryKey("id")
}
