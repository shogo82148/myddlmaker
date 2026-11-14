package schema

import (
	"github.com/shogo82148/myddlmaker"
)

type User struct {
	ID       int32 `ddl:",auto"`
	Name     string
	Age      int32
	MinScore int32
	MaxScore int32
	Memo     string
}

func (*User) PrimaryKey() *myddlmaker.PrimaryKey {
	return myddlmaker.NewPrimaryKey("id")
}

func (*User) Checks() []*myddlmaker.Check {
	return []*myddlmaker.Check{
		myddlmaker.NewCheck("chk_user_name", "char_length(`name`) > 0"),
		myddlmaker.NewCheck("chk_user_age", "`age` BETWEEN 0 AND 150"),
		myddlmaker.NewCheck("chk_user_score", "`min_score` <= `max_score`"),
		myddlmaker.NewCheck("chk_user_memo", "`memo` <> 'forbidden'").NotEnforced(),
	}
}
