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
	GetSong(string, string) (*models.GetSongDetail, error)
}

type PostgresStore struct {
	db *sql.DB
}

func NewPostgresStorage() (*PostgresStore, error) {
	connStr := "user=postgres port=5433 dbname=musicshop password=root sslmode=disable"
	db, err := sql.Open("postgres", connStr)
	if err != nil {

		return nil, fmt.Errorf("failed to establish connection to DB")
	}
	err = db.Ping()
	if err != nil {

		return nil, fmt.Errorf("ping unsuccessful")
	}
	return &PostgresStore{
		db: db,
	}, nil
}

func (s *PostgresStore) GetSong(group, song string) (*models.GetSongDetail, error) {
	details := new(models.GetSongDetail)
	query := `select (release_date, song_text, song_link) from song_list where song_group=$1 and song_name=$2`
	rows, err := s.db.Query(query, group, song)
	if err != nil {
		return nil, err
	}
	details, err = scanIntoSong(rows)

	if err != nil {
		return nil, err
	}

	return details, nil
}

func scanIntoSong(rows *sql.Rows) (*models.GetSongDetail, error) {
	songDetails := new(models.GetSongDetail)
	if err := rows.Scan(&songDetails.ReleaseDate, &songDetails.Text, &songDetails.Link); err != nil {
		return nil, err
	}
	return songDetails, nil
}
