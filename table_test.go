package myddlmaker

import "testing"

type FooBar struct{}

func TestTable(t *testing.T) {
	table := newTable(&FooBar{})
	if table.name != "foo_bar" {
		t.Errorf("unexpected table name: want %q, got %q", "foo_bar", table.name)
	}
}
