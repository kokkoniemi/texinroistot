package db

import (
	"context"
	"database/sql"

	"github.com/kokkoniemi/texinroistot/internal/config"
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
