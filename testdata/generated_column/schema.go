package schema

import (
	"github.com/shogo82148/myddlmaker"
)

type Item struct {
	ID      int32 `ddl:",auto"`
	Price   int32
	Count   int32
	Total   int32  `ddl:",generated=price * count"`
	Summary string `ddl:",generated=concat(price, ' x ', count),stored"`
	Label   string `ddl:",generated=concat('(', price, ',', count),stored,comment=string literals with parentheses and commas"`
}

func (*Item) PrimaryKey() *myddlmaker.PrimaryKey {
	return myddlmaker.NewPrimaryKey("id")
}
