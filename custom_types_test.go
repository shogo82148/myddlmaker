package myddlmaker

import (
	"context"
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

	type Object struct {
		A string `json:"a"`
		B int    `json:"b"`
	}

	var obj JSON[Object]
	obj.V = Object{
		A: "string value",
		B: 42,
	}
	_, err := db.ExecContext(ctx, "INSERT INTO `foo` (`id`, `object`) VALUES (?, ?)", 1, obj)
	if err != nil {
		t.Fatal(err)
	}
}
