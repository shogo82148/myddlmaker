package myddlmaker

import (
	"context"
	"reflect"
	"testing"
)

func TestJSON(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	db, ok := setupDatabase(ctx, t)
	if !ok {
		return
	}

	ddl := "CREATE TABLE `foo` (`id` INTEGER, `object` JSON, PRIMARY KEY (`id`))"
	if _, err := db.ExecContext(ctx, ddl); err != nil {
		t.Fatal(err)
	}

	type Object = JSON[struct {
		A string `json:"a"`
		B int    `json:"b"`
	}]

	// insert an Object as a JSON value
	obj0 := Object{{
		A: "string value",
		B: 42,
	}}
	_, err := db.ExecContext(ctx, "INSERT INTO `foo` (`id`, `object`) VALUES (?, ?)", 1, obj0)
	if err != nil {
		t.Fatal(err)
	}

	// get the JSON value
	var obj1 JSON[Object]
	row := db.QueryRowContext(ctx, "SELECT `object` FROM `foo` WHERE `id` = ?", 1)
	if err := row.Scan(&obj1); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(obj0, obj1) {
		t.Errorf("result not match: got %#v, want %#v", obj0, obj1)
	}
}
