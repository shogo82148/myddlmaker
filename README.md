# myddlmaker

![Build Status](https://github.com/shogo82148/myddlmaker/workflows/Go/badge.svg)
[![Go Reference](https://pkg.go.dev/badge/github.com/shogo82148/myddlmaker.svg)](https://pkg.go.dev/github.com/shogo82148/myddlmaker)

myddlmaker is generate DDL (Data Definition Language) from Go structs.
It is a fork of [kayac/ddl-maker](https://github.com/kayac/ddl-maker) that focuses to use with MySQL.

## SYNOPSIS

Firstly, write your table definitions as Go structures.
Here is an example: [schema.go](./testdata/example/schema.go)

```go
package schema

import (
	"time"

	"github.com/shogo82148/myddlmaker"
)

//go:generate go run -tags myddlmaker gen/main.go

type User struct {
	ID        uint64 `ddl:",auto"`
	Name      string
	CreatedAt time.Time
}

func (*User) PrimaryKey() *myddlmaker.PrimaryKey {
	return myddlmaker.NewPrimaryKey("id")
}
```

Next, configure your DDL maker: [gen/main.go](./testdata/example/gen/main.go)

```go
package main

import (
	"log"

	"github.com/shogo82148/myddlmaker"
	schema "github.com/shogo82148/myddlmaker/testdata/example"
)

func main() {
	// create a new DDL maker.
	m, err := myddlmaker.New(&myddlmaker.Config{
		DB: &DBConfig{
			Engine:  "InnoDB",
			Charset: "utf8mb4",
			Collate: "utf8mb4_bin",
		},
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
```

Run `go generate`.

```console
$ go generate ./...
```

You can get the following SQL queries:

```sql
SET foreign_key_checks=0;

DROP TABLE IF EXISTS `user`;

CREATE TABLE `user` (
    `id` BIGINT UNSIGNED NOT NULL AUTO_INCREMENT,
    `name` VARCHAR(191) NOT NULL,
    `created_at` DATETIME NOT NULL,
    PRIMARY KEY (`id`)
) ENGINE = InnoDB DEFAULT CHARACTER SET = utf8mb4;

SET foreign_key_checks=1;
```

And more, the DDL maker generates Go source code for basic SQL operations such as INSERT, SELECT, and UPDATE.

```go
// Code generated by https://github.com/shogo82148/myddlmaker; DO NOT EDIT.

//go:build !myddlmaker

package schema

import (
	"context"
	"database/sql"
)

type execer interface {
	ExecContext(ctx context.Context, query string, args ...any) (sql.Result, error)
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
}

type queryer interface {
	QueryContext(ctx context.Context, query string, args ...any) (*sql.Rows, error)
	QueryRowContext(ctx context.Context, query string, args ...any) *sql.Row
	PrepareContext(ctx context.Context, query string) (*sql.Stmt, error)
}

func InsertUser(ctx context.Context, execer execer, values ...*User) error {
	const q = "INSERT INTO `user` (`name`, `created_at`) VALUES (?, ?)"
    // (snip)
	return nil
}

func SelectUser(ctx context.Context, queryer queryer, primaryKeys *User) (*User, error) {
	var v User
	row := queryer.QueryRowContext(ctx, "SELECT `id`, `name`, `created_at` FROM `user` WHERE `id` = ?", primaryKeys.ID)
	if err := row.Scan(&v.ID, &v.Name, &v.CreatedAt); err != nil {
		return nil, err
	}
	return &v, nil
}

func SelectAllUser(ctx context.Context, queryer queryer) ([]*User, error) {
	var ret []*User
	rows, err := queryer.QueryContext(ctx, "SELECT `id`, `name`, `created_at` FROM `user` ORDER BY `id`")
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	for rows.Next() {
		var v User
		if err := rows.Scan(&v.ID, &v.Name, &v.CreatedAt); err != nil {
			return nil, err
		}
		ret = append(ret, &v)
	}
	if err := rows.Err(); err != nil {
		return nil, err
	}
	return ret, nil
}

func UpdateUser(ctx context.Context, execer execer, values ...*User) error {
	stmt, err := execer.PrepareContext(ctx, "UPDATE `user` SET `name` = ?, `created_at` = ? WHERE `id` = ?")
	if err != nil {
		return err
	}
	defer stmt.Close()
	for _, value := range values {
		if _, err := stmt.ExecContext(ctx, value.Name, value.CreatedAt, value.ID); err != nil {
			return err
		}
	}
	return nil
}
```

You can use these generated functions in your application.

```go
db, _ := sql.Open("mysql", "user:password@/dbname")

// INSERT INTO `user` (`name`, `created_at`) VALUES ("Alice", NOW());
schema.InsertUser(context.TODO(), db, &schema.User{
	Name:      "Alice",
	CreatedAt: time.Now(),
})

// SELECT * FROM `user` WHERE `id` = 1;
user, err := schema.SelectUser(context.TODO(), db, &schema.User{
	ID: 1,
})

// UPDATE `user` SET `name` = "Bob", `created_at` = NOW() WHERE `id` = 1;
schema.UpdateUser(context.TODO(), db, &schema.User{
	ID: 1,
	Name: "Bob",
	CreatedAt: time.Now(),
})
```

## MySQL Types and Go Types

|         Golang Type          |    MySQL Column     |
| :--------------------------: | :-----------------: |
|            `int8`            |      `TINYINT`      |
|           `int16`            |     `SMALLINT`      |
|           `int32`            |      `INTEGER`      |
|   `int64`, `sql.NullInt64`   |      `BIGINT`       |
|   `uint8`, `sql.NullByte`    | `TINYINT UNSIGNED`  |
|           `uint16`           | `SMALLINT UNSIGNED` |
|           `uint32`           | `INTEGER UNSIGNED`  |
|           `uint64`           |  `BIGINT UNSIGNED`  |
|          `float32`           |       `FLOAT`       |
| `float64`, `sql.NullFloat64` |      `DOUBLE`       |
|  `string`, `sql.NullString`  |      `VARCHAR`      |
|    `bool`, `sql.NullBool`    |    `TINYINT(1)`     |
|           `[]byte`           |      `VARCHAR`      |
|         `time.Time`          |    `DATETIME(6)`    |
|      `json.RawMessage`       |       `JSON`        |

## Go Struct Tag Options

|      Tag Value      |                SQL Fragment                 |
| :-----------------: | :-----------------------------------------: |
|       `null`        |        `NULL` (default: `NOT NULL`)         |
|       `auto`        |              `AUTO INCREMENT`               |
|     `invisible`     |                 `INVISIBLE`                 |
|     `unsigned`      |                  `UNSIGNED`                 |
|    `size=<size>`    | `VARCHAR(<size>)`, `DATETIME(<size>)`, etc. |
|    `type=<type>`    |             override field type             |
|    `srid=<srid>`    |                override SRID                |
|  `default=<value>`  |              `DEFAULT <value>`              |
| `charset=<charset>` |          `CHARACTER SET <charset>`          |
| `collate=<collate>` |             `COLLATE <collate>`             |
| `comment=<comment>` |             `COMMENT <comment>`             |

## Primary Index

Implement the `PrimaryKey` method to define the primary index.

```go
func (*User) PrimaryKey() *myddlmaker.PrimaryKey {
    // PRIMARY KEY (`id1`, `id2`)
    return myddlmaker.NewPrimaryKey("id1", "id2")
}
```

## Indexes

Implement the `Indexes` method to define the indexes.

```go
func (*User) Indexes() []*myddlmaker.Index {
    return []*myddlmaker.Index{
        // INDEX `idx_name` (`name`)
        myddlmaker.NewIndex("idx_name", "name"),

        // INDEX `idx_name` (`name`) COMMENT 'some comment'
        myddlmaker.NewIndex("idx_name", "name").Comment("some comment"),

        // INDEX `idx_name` (`name`) INVISIBLE
        myddlmaker.NewIndex("idx_name", "name").Invisible(),
    }
}
```

## Unique Indexes

Implement the `UniqueIndexes` method to define the unique indexes.

```go
func (*User) UniqueIndexes() []*myddlmaker.UniqueIndex {
    return []*myddlmaker.UniqueIndex{
        // UNIQUE INDEX `idx_name` (`name`)
        myddlmaker.NewUniqueIndex("idx_name", "name"),

        // UNIQUE INDEX `idx_name` (`name`) COMMENT 'some comment'
        myddlmaker.NewUniqueIndex("idx_name", "name").Comment("some comment"),

        // UNIQUE INDEX `idx_name` (`name`) INVISIBLE
        myddlmaker.NewUniqueIndex("idx_name", "name").Invisible(),
    }
}
```

## Foreign Key Constraints

Implement the `ForeignKeys` method to define the foreign key constraints.

```go
func (*User) ForeignKeys() []*myddlmaker.ForeignKey {
    return []*myddlmaker.ForeignKey{
        // CONSTRAINT `name_of_constraint`
        //     FOREIGN KEY (`column1`, `column2`)
        //     REFERENCES `another_table` (`id1`, `id2`)
        myddlmaker.NewForeignKey(
            "name_of_constraint",
            []string{"column1", "column2"},
            "another_table",
            []string{"id1", "id2"},
        ),

        // CONSTRAINT `name_of_constraint`
        //     FOREIGN KEY (`column1`, `column2`)
        //     REFERENCES `another_table` (`id1`, `id2`)
        //     ON DELETE CASCADE
        myddlmaker.NewForeignKey(
            "name_of_constraint",
            []string{"column1", "column2"},
            "another_table",
            []string{"id1", "id2"},
        ).OnDelete(myddlmaker.ForeignKeyOptionCascade),
    }
}
```

## Spatial Indexes

Implement the `SpatialIndexes` method to define the spatial indexes.

```go
func (*User) SpatialIndexes() []*myddlmaker.SpatialIndex {
    return []*myddlmaker.SpatialIndex{
        // SPATIAL INDEX `idx_name` (`name`)
        myddlmaker.NewSpatialIndex("idx_name", "name"),

        // SPATIAL INDEX `idx_name` (`name`) COMMENT 'some comment'
        myddlmaker.NewSpatialIndex("idx_name", "name").Comment("some comment"),

        // SPATIAL INDEX `idx_name` (`name`) INVISIBLE
        myddlmaker.NewSpatialIndex("idx_name", "name").Invisible(),
    }
}
```

## Full Text Indexes

Implement the `FullTextIndexes` method to define the full-text indexes.

```go
func (*User) FullTextIndexes() []*myddlmaker.FullTextIndex {
    return []*myddlmaker.FullTextIndex{
        // FULLTEXT INDEX `idx_name` (`name`)
        myddlmaker.NewFullTextIndex("idx_name", "name"),

        // FULLTEXT INDEX `idx_name` (`name`) COMMENT 'some comment'
        myddlmaker.NewFullTextIndex("idx_name", "name").Comment("some comment"),

        // FULLTEXT INDEX `idx_name` (`name`) INVISIBLE
        myddlmaker.NewFullTextIndex("idx_name", "name").Invisible(),
    }
}
```
