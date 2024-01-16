package schema

import (
	"github.com/shogo82148/myddlmaker"
)

type User struct {
	ID int32 `ddl:",auto"`
}

func (*User) PrimaryKey() *myddlmaker.PrimaryKey {
	return myddlmaker.NewPrimaryKey("id")
}
