package db

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/kokkoniemi/texinroistot/internal/config"
	"github.com/lib/pq"
	_ "github.com/lib/pq"
)

var (
	pgdb *sql.DB
)

func GetDB() (*sql.DB, error) {
	if pgdb != nil {
		return pgdb, nil
	}

	conn, err := sql.Open("postgres", config.DBConnectionString)
	if err != nil {
		return nil, err
	}

	if err = conn.Ping(); err != nil {
		return nil, err
	}

	pgdb = conn
	return pgdb, nil
}

func Execute(q string, args ...any) (sql.Result, error) {
	db, err := GetDB()
	if err != nil {
		return nil, err
	}

	return db.ExecContext(context.Background(), q, args...)
}

func Query(q string, args ...any) (*sql.Rows, error) {
	db, err := GetDB()
	if err != nil {
		return nil, err
	}

	return db.QueryContext(context.Background(), q, args...)
}

func StartTransaction() (*sql.Tx, error) {
	db, err := GetDB()
	if err != nil {
		return nil, err
	}

	return db.Begin()
}

type bulkInsertParams struct {
	Table   string
	Columns []string
	Values  [][]interface{}
}

func BulkInsertTxn(params bulkInsertParams) (int64, error) {
	db, err := GetDB()
	if err != nil {
		return 0, err
	}

	txn, err := db.Begin()
	if err != nil {
		return 0, err
	}

	stmt, err := txn.Prepare(pq.CopyIn(
		params.Table,
		params.Columns...,
	))

	if err != nil {
		return 0, err
	}

	for _, v := range params.Values {
		_, err = stmt.Exec(v...)
		if err != nil {
			return 0, err
		}
	}

	res, err := stmt.Exec()
	if err != nil {
		return 0, err
	}

	err = stmt.Close()
	if err != nil {
		return 0, err
	}

	err = txn.Commit()
	if err != nil {
		return 0, err
	}

	rows, err := res.RowsAffected()
	if err != nil {
		return 0, err
	}

	if int(rows) != len(params.Values) {
		return 0, fmt.Errorf("something went wrong inserting to table '%s'", params.Table)
	}

	return rows, nil
}
