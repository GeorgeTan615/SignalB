package database

import (
	"database/sql"
)

type DBClient struct {
	DB *sql.DB
}

func newDBClient(db *sql.DB) *DBClient {
	return &DBClient{
		DB: db,
	}
}

func (d *DBClient) Close() {
	d.DB.Close()
}
