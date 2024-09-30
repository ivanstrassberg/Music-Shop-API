package database

import (
	"database/sql"
	"fmt"
	models "musicShopBackend/musicModels"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/github"
	_ "github.com/lib/pq"
)

type Storage interface {
	AddSong(string, string) error
}

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStorage() (*PostgresStore, error) {
	connStr := "user=postgres port=5433 dbname=musicshop password=root sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {
		fmt.Errorf("Failed to establish connection to DB")
		return nil, err
	}
	err = db.Ping()
	if err != nil {
		fmt.Errorf("Ping unsuccessful")
		return nil, err
	}
	return &PostgresStore{
		db: db,
	}, nil
}

func (s *PostgresStore) AddSong(group, song string) (*models.SongDetail, error) {
	details := new(models.SongDetail)
	query := `select (release_date, song_text, song_link) from song_list where song_group=$1 and song_name=$2`
	rows, err := s.db.Query(query, group, song)
	if err != nil {
		return nil, err
	}
	return details, nil
}

func scanIntoSong(rows *sql.Rows) {

}
