package database

import (
	"database/sql"
	"fmt"
)

type Storage struct {
	db *sql.DB
}

func (db *Storage) Init() error {
	connStr := "user=postgres port=5433 dbname=musicShop password=root sslmode=disable"
	conn, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Errorf("Failed to establish connection to DB")
		return err
	}
	err = conn.Ping()
	if err != nil {
		fmt.Errorf("Ping unsuccessful")
		return err
	}
	return nil
}
