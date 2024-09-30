package database

import (
	"database/sql"
	"fmt"
	models "musicShopBackend/musicModels"
	"time"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/github"
	_ "github.com/lib/pq"
)

type Storage interface {
	GetSong(string, string) (*models.GetSongDetail, error)
	UpdateSong(string, string, string, string, time.Time, string, string) error
	AddSong(string, string) error
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
	query := `select (release_date, song_text, song_link) from songs where group_name=$1 and song_name=$2`
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
	for rows.Next() {
		if err := rows.Scan(&songDetails.ReleaseDate, &songDetails.Text, &songDetails.Link); err != nil {
			return nil, err
		}
	}

	return songDetails, nil
}

func (s *PostgresStore) UpdateSong(songKey, groupKey, songName, groupName string, releaseDate time.Time, songText, songLink string) error {
	query := `update songs 
set song_name = $3, 
    group_name = $4, 
    release_date = TO_DATE($5, 'DD.MM.YYYY'), 
    song_text = $6, 
    song_link = $7,
	udpated_at = $8
where song_name = $1 and group_name = $2;`
	_, err := s.db.Query(query, songKey, groupKey, songName, groupName, releaseDate, songText, songLink, time.Now())
	if err != nil {
		return err
	}
	return nil
}

func (s *PostgresStore) AddSong(songName, groupName string) error {
	query := `INSERT INTO songs (song_name, group_name)
VALUES ($1,$2);`
	_, err := s.db.Query(query, songName, groupName)
	if err != nil {
		return err
	}
	return nil
}
