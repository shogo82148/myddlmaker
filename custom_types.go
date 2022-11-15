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

// Get returns the value of v.
func (v JSON[T]) Get() T {
	return v[0]
}

// Set sets v = u.
func (v *JSON[T]) Set(u T) {
	v[0] = u
}

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

type jsonMarker interface {
	jsonMarker()
}

// jsonMarker is a marker for the reflect package.
// On Go 1.19, the reflect package can't handle generic types correctly.
// However, it can handle a interface implemented by generic types.
func (v JSON[T]) jsonMarker() { /* nothing to do */ }
