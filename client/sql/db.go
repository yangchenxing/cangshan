package sql

import (
	gosql "database/sql"
	"fmt"
	"regexp"

	"github.com/yangchenxing/cangshan/application"
	"github.com/yangchenxing/cangshan/logging"
)

func init() {
	application.RegisterModulePrototype("SQLDB", new(DB))
}

var (
	lineSeperator = regexp.MustCompile("[\n\t ]+")
	ErrNoRows     = gosql.ErrNoRows
	ErrTxDone     = gosql.ErrTxDone
)

func normalizeSQLQuery(query string) string {
	return lineSeperator.ReplaceAllString(query, " ")
}

type Rows struct {
	*gosql.Rows
}
type Row struct {
	*gosql.Row
}
type Result gosql.Result

// DB is a wrapper of standard sql.DB. It output query to debug log
type DB struct {
	*gosql.DB
	Driver     string
	DataSource string
	Debug      bool
}

// Initialize the DB module for application
func (db *DB) Initialize() error {
	var err error
	if db.DB, err = gosql.Open(db.Driver, db.DataSource); err != nil {
		return fmt.Errorf("open sql db fail: %s", err.Error())
	}
	return nil
}

// Begin a transaction
func (db *DB) Begin() (*Tx, error) {
	if db.Debug {
		logging.Debug("Begin SQL Transaction")
	}
	tx, err := db.DB.Begin()
	if err != nil {
		return nil, err
	}
	return &Tx{tx, db}, nil
}

// Exec executes a non-select query
func (db *DB) Exec(query string, args ...interface{}) (Result, error) {
	if db.Debug {
		logging.Debug("SQL: query=\"%s\", params=%v", normalizeSQLQuery(query), args)
	}
	return db.DB.Exec(query, args...)
}

// Prepare a query statement
func (db *DB) Prepare(query string) (*Stmt, error) {
	s, err := db.DB.Prepare(query)
	if err != nil {
		return nil, err
	}
	return &Stmt{s, query, db}, nil
}

// Query multiple rows
func (db *DB) Query(query string, args ...interface{}) (*Rows, error) {
	if db.Debug {
		logging.Debug("SQL: query=\"%s\", params=%v", normalizeSQLQuery(query), args)
	}
	rows, err := db.DB.Query(query, args...)
	return &Rows{rows}, err
}

// QueryRow query single row
func (db *DB) QueryRow(query string, args ...interface{}) *Row {
	if db.Debug {
		logging.Debug("SQL: query=\"%s\", params=%v", normalizeSQLQuery(query), args)
	}
	return &Row{db.DB.QueryRow(query, args...)}
}
