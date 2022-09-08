package main

import (
	"log"

	"github.com/shogo82148/myddlmaker"
	schema "github.com/shogo82148/myddlmaker/testdata/example"
)

func main() {
	// create a new DDL maker.
	m, err := myddlmaker.New(&myddlmaker.Config{
		DB: &myddlmaker.DBConfig{
			Engine:  "InnoDB",
			Charset: "utf8mb4",
		},
		OutFilePath:   "schema.sql",
		OutGoFilePath: "schema_gen.go",
		PackageName:   "schema",
		Tag:           "myddlmaker",
	})
	if err != nil {
		log.Fatal(err)
	}

	m.AddStructs(&schema.User{})

	// generate an SQL file.
	if err := m.GenerateFile(); err != nil {
		log.Fatal(err)
	}

	// generate Go source code for basic SQL operations
	// such as INSERT, SELECT, and UPDATE.
	if err := m.GenerateGoFile(); err != nil {
		log.Fatal(err)
	}
}
