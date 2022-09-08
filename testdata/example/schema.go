package schema

import (
	"time"

	"github.com/shogo82148/myddlmaker"
)

type User struct {
	ID        uint64 `ddl:",auto"`
	Name      string
	CreatedAt time.Time
}

func (*User) PrimaryKey() *myddlmaker.PrimaryKey {
	return myddlmaker.NewPrimaryKey("id")
}
