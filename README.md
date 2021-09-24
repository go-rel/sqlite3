# sqlite3

[![GoDoc](https://godoc.org/github.com/go-rel/sqlite3?status.svg)](https://pkg.go.dev/github.com/go-rel/sqlite3)
[![Tesst](https://github.com/go-rel/sqlite3/actions/workflows/test.yml/badge.svg?branch=main)](https://github.com/go-rel/sqlite3/actions/workflows/test.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/go-rel/sqlite3)](https://goreportcard.com/report/github.com/go-rel/sqlite3)
[![codecov](https://codecov.io/gh/go-rel/sqlite3/branch/main/graph/badge.svg?token=GX2dOCV7Cq)](https://codecov.io/gh/go-rel/sqlite3)
[![Gitter chat](https://badges.gitter.im/go-rel/rel.png)](https://gitter.im/go-rel/rel)

SQLite3 adapter for REL.

## Example 

```go
package main

import (
	"context"

	_ "github.com/mattn/go-sqlite3"
	"github.com/go-rel/mssql"
	"github.com/go-rel/rel"
)

func main() {
	// open sqlite3 connection.
	adapter, err := sqlite3.Open("dev.db")
	if err != nil {
		panic(err)
	}
	defer adapter.Close()

	// initialize rel's repo.
	repo := rel.New(adapter)
	repo.Ping(context.TODO())
}
```