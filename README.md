# scany (KF Edition)

## Moifications

All credit where credit is due, this project is based off scany by Georgy Savva ( https://github.com/georgysavva/scany ), The aim of this version is to make it pgx compatible only, making this lighter weight also adding features for inserting with a struct (with tag definition)

## Overview

Go favors simplicity, and it's pretty common to work with a database via driver directly without any ORM.
It provides great control and efficiency in your queries, but here is a problem: 
you need to manually iterate over database rows and scan data from all columns into a corresponding destination.
It can be error-prone verbose and just tedious. 
scany aims to solve this problem, 
it allows developers to scan complex data from a database into Go structs and other composite types 
with just one function call and don't bother with rows iteration.

scany isn't limited to any specific database. It integrates with `database/sql`, 
so any database with `database/sql` driver is supported. 
It also works with [pgx](https://github.com/jackc/pgx) library native interface. 
Apart from the out of the box support, scany can be easily extended to work with almost any database library.

Note that, scany isn't an ORM. First of all, it works only in one direction: 
it scans data into Go objects from the database, but it can't build database queries based on those objects.
Secondly, it doesn't know anything about relations between objects e.g: one to many, many to many.

## Features

* Custom database column name via struct tag
* Reusing structs via nesting or embedding 
* NULLs and custom types support
* Omitted struct fields
* Apart from structs, support for other destination types: maps, slices and etc.

## Install

```
go get github.com/KirksFletcher/scany
```


## `pgx` native interface
### Select example

```go
package main

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"

	"github.com/KirksFletcher/pgxscan"
)

type User struct {
	ID    string
	Name  string
	Email string
	Age   int
}

func main() {
	ctx := context.Background()
	db, _ := pgxpool.Connect(ctx, "example-connection-url")

	var users []*User
	pgxscan.Select(ctx, db, &users, `SELECT id, name, email, age FROM users`)
	// users variable now contains data from all rows.
}
```

### Insert example

```go
package main

import (
	"context"

	"github.com/jackc/pgx/v4/pgxpool"
	"github.com/KirksFletcher/pgxscan"
)

type User struct {
	ID    string `pgx:"id"` //will be inserted as id
	Name  string `pgx:"user_name"` //will be inserted as user_name
	Email string `pgx:"user_email"` //will be inserted as user_email
	UserAge   int     //no pgx tag will auto snakecase:- will be inserted as user_age
}

func main() {
	ctx := context.Background()
	db, _ := pgxpool.Connect(ctx, "example-connection-url")

    user := User{
        ID:     "user_1",
        Name:   "User Name",
        Email:  "user@email.com",
        Age:    40,
        }	

	pgxscan.Insert(ctx, db, user, "my_user_table", ` ADDITIONAL QUERIES TO BE APPENDED OR BLANK`)
	
}
```

Use [`pgxscan`](https://pkg.go.dev/github.com/KirksFletcher/pgxscan) 
package to work with `pgx` library native interface. 

## How to use with other database libraries

Use [`dbscan`](https://pkg.go.dev/github.com/KirksFletcher/dbscan) package that works with an abstract database, 
and can be integrated with any library that has a concept of rows. 
This particular package implements core scany features and contains all the logic.

## Comparisson with [sqlx](https://github.com/jmoiron/sqlx)

* sqlx only works with `database/sql` standard library. scany isn't limited only to `database/sql`, it also supports [pgx](https://github.com/jackc/pgx) native interface and can be extended to work with any database library independent of `database/sql`
* In terms of scanning and mapping abilities, scany provides all [features](https://github.com/KirksFletcher#features) of sqlx
* scany has a simpler API and much fewer concepts, so it's easier to start working with

## Supported Go versions 

scany supports Go 1.13 and higher.

## Roadmap   

* Add ability to set custom function to translate struct field to column name, 
instead of the default to snake case function 
* Allow to use a custom separator for embedded structs prefix, instead of the default "."

## Tests

The only thing you need to run tests locally is an internet connection, 
it's required to download and cache the database binary.
Just type `go test ./...` inside scany root directory and let the code do the rest. 

## Contributing 

Every feature request or question is appreciated. Don't hesitate, just post an issue or PR.

## License

This project is licensed under the terms of the MIT license.
