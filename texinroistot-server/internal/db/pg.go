package db

import (
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
