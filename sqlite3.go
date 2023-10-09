// Package sqlite3 wraps go-sqlite3 driver as an adapter for rel.
//
// Usage:
//
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
	"log"
	"strings"

	"github.com/go-rel/rel"
	"github.com/go-rel/sql"
	"github.com/go-rel/sql/builder"
)

// SQLite3 adapter
type SQLite3 struct {
	sql.SQL
}

// Name of database type this adapter implements.
const Name string = "sqlite3"

// New sqlite3 adapter using existing connection.
func New(database *db.DB) rel.Adapter {
	var (
		bufferFactory     = builder.BufferFactory{ArgumentPlaceholder: "?", BoolTrueValue: "1", BoolFalseValue: "0", Quoter: builder.Quote{IDPrefix: "\"", IDSuffix: "\"", IDSuffixEscapeChar: "\"", ValueQuote: "'", ValueQuoteEscapeChar: "'"}}
		filterBuilder     = builder.Filter{}
		queryBuilder      = builder.Query{BufferFactory: bufferFactory, Filter: filterBuilder}
		OnConflictBuilder = builder.OnConflict{Statement: "ON CONFLICT", IgnoreStatement: "DO NOTHING", UpdateStatement: "DO UPDATE SET", TableQualifier: "excluded", SupportKey: true}
		InsertBuilder     = builder.Insert{BufferFactory: bufferFactory, InsertDefaultValues: true, OnConflict: OnConflictBuilder}
		insertAllBuilder  = builder.InsertAll{BufferFactory: bufferFactory, OnConflict: OnConflictBuilder}
		updateBuilder     = builder.Update{BufferFactory: bufferFactory, Query: queryBuilder, Filter: filterBuilder}
		deleteBuilder     = builder.Delete{BufferFactory: bufferFactory, Query: queryBuilder, Filter: filterBuilder}
		ddlBufferFactory  = builder.BufferFactory{InlineValues: true, BoolTrueValue: "1", BoolFalseValue: "0", Quoter: builder.Quote{IDPrefix: "\"", IDSuffix: "\"", IDSuffixEscapeChar: "\"", ValueQuote: "'", ValueQuoteEscapeChar: "'"}}
		ddlQueryBuilder   = builder.Query{BufferFactory: ddlBufferFactory, Filter: filterBuilder}
		tableBuilder      = builder.Table{BufferFactory: ddlBufferFactory, ColumnMapper: columnMapper, ColumnOptionsMapper: columnOptionsMapper, DefinitionFilter: definitionFilter, DropKeyMapper: sql.DropKeyMapper}
		indexBuilder      = builder.Index{BufferFactory: ddlBufferFactory, Query: ddlQueryBuilder, Filter: filterBuilder, SupportFilter: true}
	)

	return &SQLite3{
		SQL: sql.SQL{
			QueryBuilder:     queryBuilder,
			InsertBuilder:    InsertBuilder,
			InsertAllBuilder: insertAllBuilder,
			UpdateBuilder:    updateBuilder,
			DeleteBuilder:    deleteBuilder,
			TableBuilder:     tableBuilder,
			IndexBuilder:     indexBuilder,
			Increment:        -1,
			ErrorMapper:      errorMapper,
			DB:               database,
		},
	}
}

// Open sqlite3 connection using dsn.
func Open(dsn string) (rel.Adapter, error) {
	database, err := db.Open("sqlite3", dsn)
	return New(database), err
}

// MustOpen sqlite3 connection using dsn.
func MustOpen(dsn string) rel.Adapter {
	database, err := db.Open("sqlite3", dsn)
	if err != nil {
		panic(err)
	}
	return New(database)
}

// Name of database adapter.
func (SQLite3) Name() string {
	return Name
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
		typ  string
		m, n int
	)

	switch column.Type {
	case rel.ID:
		return "INTEGER", 0, 0
	case rel.BigID:
		return "INTEGER", 0, 0
	case rel.Int:
		typ = "INTEGER"
		m = column.Limit
	default:
		typ, m, n = sql.ColumnMapper(column)
	}

	if column.Unsigned {
		typ = "UNSIGNED " + typ
	}

	return typ, m, n
}

func columnOptionsMapper(column *rel.Column) string {
	var buffer strings.Builder

	if column.Primary {
		buffer.WriteString(" PRIMARY KEY")
		if column.Type == rel.ID || column.Type == rel.BigID {
			buffer.WriteString(" AUTOINCREMENT")
		}
	}

	if column.Required {
		buffer.WriteString(" NOT NULL")
	}

	if column.Unique {
		buffer.WriteString(" UNIQUE")
	}

	buf := buffer.String()
	if buf != "" {
		buf = buf[1:]
	}

	return buf
}

func definitionFilter(table rel.Table, def rel.TableDefinition) bool {
	if table.Op == rel.SchemaAlter {
		// https://www.sqlite.org/omitted.html
		// > Only the RENAME TABLE, ADD COLUMN, RENAME COLUMN, and DROP COLUMN variants of the ALTER TABLE command are supported.
		_, ok := def.(rel.Key)
		if ok {
			log.Print("[REL] SQLite3 adapter does not support adding keys when modifying tables")

			return false
		}
	}

	return true
}
