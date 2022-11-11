package myddlmaker

import (
	"bytes"
	"database/sql"
	"database/sql/driver"
	"encoding/json"
	"fmt"
	"io"
	"strings"
)

var _ driver.Valuer = JSON[int]{}
var _ sql.Scanner = (*JSON[int])(nil)

type JSON[T any] [1]T

// Value implements [database/sql/driver.Valuer] interface.
func (v JSON[T]) Value() (driver.Value, error) {
	return json.Marshal(v[0])
}

// Scan implements [database/sql.Scanner] interface.
func (v *JSON[T]) Scan(src any) error {
	var r io.Reader
	switch src := src.(type) {
	case []byte:
		r = bytes.NewReader(src)
	case string:
		r = strings.NewReader(src)
	default:
		return fmt.Errorf("myddlmaker: unsupported type: %T", src)
	}

	dec := json.NewDecoder(r)
	return dec.Decode(&v[0])
}
