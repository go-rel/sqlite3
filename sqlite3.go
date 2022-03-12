// Package sqlite3 wraps go-sqlite3 driver as an adapter for rel.
//
// Usage:
//	// open sqlite3 connection.
//	adapter, err := sqlite3.Open("dev.db")
//	if err != nil {
//		panic(err)
//	}
//	defer adapter.Close()
//
//	// initialize rel's repo.
//	repo := rel.New(adapter)
package sqlite3

import (
	db "database/sql"
	"strings"

	"github.com/go-rel/rel"
	"github.com/go-rel/sql"
	"github.com/go-rel/sql/builder"
)

// New sqlite3 adapter using existing connection.
func New(database *db.DB) rel.Adapter {
	var (
		bufferFactory     = builder.BufferFactory{ArgumentPlaceholder: "?", BoolTrueValue: "1", BoolFalseValue: "0", Quoter: builder.Quote{IDPrefix: "\"", IDSuffix: "\"", IDSuffixEscapeChar: "\"", ValueQuote: "'", ValueQuoteEscapeChar: "'"}}
		filterBuilder     = builder.Filter{}
		queryBuilder      = builder.Query{BufferFactory: bufferFactory, Filter: filterBuilder}
		OnConflictBuilder = builder.OnConflict{Statement: "ON CONFLICT", IgnoreStatement: "DO NOTHING", UpdateStatement: "DO UPDATE SET", TableQualifier: "EXCLUDED", SupportKey: true}
		InsertBuilder     = builder.Insert{BufferFactory: bufferFactory, InsertDefaultValues: true, OnConflict: OnConflictBuilder}
		insertAllBuilder  = builder.InsertAll{BufferFactory: bufferFactory, OnConflict: OnConflictBuilder}
		updateBuilder     = builder.Update{BufferFactory: bufferFactory, Query: queryBuilder, Filter: filterBuilder}
		deleteBuilder     = builder.Delete{BufferFactory: bufferFactory, Query: queryBuilder, Filter: filterBuilder}
		ddlBufferFactory  = builder.BufferFactory{InlineValues: true, BoolTrueValue: "1", BoolFalseValue: "0", Quoter: builder.Quote{IDPrefix: "\"", IDSuffix: "\"", IDSuffixEscapeChar: "\"", ValueQuote: "'", ValueQuoteEscapeChar: "'"}}
		ddlQueryBuilder   = builder.Query{BufferFactory: ddlBufferFactory, Filter: filterBuilder}
		tableBuilder      = builder.Table{BufferFactory: ddlBufferFactory, ColumnMapper: columnMapper}
		indexBuilder      = builder.Index{BufferFactory: ddlBufferFactory, Query: ddlQueryBuilder, Filter: filterBuilder, SupportFilter: true}
	)

	return &sql.SQL{
		QueryBuilder:     queryBuilder,
		InsertBuilder:    InsertBuilder,
		InsertAllBuilder: insertAllBuilder,
		UpdateBuilder:    updateBuilder,
		DeleteBuilder:    deleteBuilder,
		TableBuilder:     tableBuilder,
		IndexBuilder:     indexBuilder,
		IncrementFunc:    incrementFunc,
		ErrorMapper:      errorMapper,
		DB:               database,
	}
}

// Open sqlite3 connection using dsn.
func Open(dsn string) (rel.Adapter, error) {
	var database, err = db.Open("sqlite3", dsn)
	return New(database), err
}

func incrementFunc(adapter sql.SQL) int {
	// decrement
	return -1
}

func errorMapper(err error) error {
	if err == nil {
		return nil
	}

	var (
		msg         = err.Error()
		failedSep   = " failed: "
		failedIndex = strings.Index(msg, failedSep)
		failedLen   = 9 // len(failedSep)
	)

	if failedIndex < 0 {
		failedIndex = 0
	}

	switch msg[:failedIndex] {
	case "UNIQUE constraint":
		return rel.ConstraintError{
			Key:  msg[failedIndex+failedLen:],
			Type: rel.UniqueConstraint,
			Err:  err,
		}
	case "CHECK constraint":
		return rel.ConstraintError{
			Key:  msg[failedIndex+failedLen:],
			Type: rel.CheckConstraint,
			Err:  err,
		}
	default:
		return err
	}
}

func columnMapper(column *rel.Column) (string, int, int) {
	var (
		typ      string
		m, n     int
		unsigned = column.Unsigned
	)

	column.Unsigned = false

	switch column.Type {
	case rel.ID:
		typ = "INTEGER"
	case rel.BigID:
		typ = "BIGINT"
	case rel.Int:
		typ = "INTEGER"
		m = column.Limit
	default:
		typ, m, n = sql.ColumnMapper(column)
	}

	if unsigned {
		typ = "UNSIGNED " + typ
	}

	return typ, m, n
}
