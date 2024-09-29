package database

import (
	"database/sql"
	"fmt"

	_ "github.com/lib/pq"
)

type Storage interface{}

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStorage() (*sql.DB, error) {
	connStr := "user=postgres port=5433 dbname=musicshop password=root sslmode=disable"
	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Errorf("Failed to establish connection to DB")
		return nil, err
	}
	err = conn.Ping()
	if err != nil {
		fmt.Errorf("Ping unsuccessful")
		return nil, err
	}
	return conn, nil
}
