package database

import (
	"database/sql"
	"fmt"
	"log"
	models "musicShopBackend/musicModels"
	"os"
	"strings"
	"time"

	_ "github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/github"
	_ "github.com/lib/pq"
)

type Storage interface {
	UpdateSong(int, string, string, time.Time, string, string) error
	AddSong(string, string) error
	SeedDataIfEmpty() error
	DeleteSong(int) error
	CreateTableIfNotExists() error
	GetFilteredSongsDataWithPagination(string, string, time.Time, string, string, int, int) ([]models.GetSong, int, error)
	GetSong(string, string, int, int) (string, error)
}

type PostgresStore struct {
	db *sql.DB
}

func (s *PostgresStore) MigrateDB() {

	s.CreateTableIfNotExists()
	s.SeedDataIfEmpty()

}

func NewPostgresStorage() (*PostgresStore, error) {
	user := os.Getenv("DB_USER")
	password := os.Getenv("DB_PASSWORD")
	dbname := os.Getenv("DB_NAME")
	port := os.Getenv("DB_PORT")
	host := os.Getenv("DB_HOST")
	connStr := fmt.Sprintf("host=%s user=%s port=%s dbname=%s password=%s sslmode=disable", host, user, port, dbname, password)
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

func (s *PostgresStore) GetSong(songName, groupName string, page, versesPerPage int) (string, error) {
	query := "SELECT song_text FROM songs WHERE 1=1"
	var args []interface{}
	argCount := 1

	if songName != "" {
		query += fmt.Sprintf(" AND song_name ILIKE '%%' || $%d || '%%'", argCount)
		args = append(args, songName)
		argCount++
	}

	if groupName != "" {
		query += fmt.Sprintf(" AND group_name ILIKE '%%' || $%d || '%%'", argCount)
		args = append(args, groupName)
		argCount++
	}

	var songText string
	err := s.db.QueryRow(query, args...).Scan(&songText)
	if err != nil {
		return "", err
	}

	verses := strings.Split(songText, "\n")

	start := (page - 1) * versesPerPage
	end := start + versesPerPage

	if start >= len(verses) {
		return "", fmt.Errorf("no verses available for this page")
	}

	if end > len(verses) {
		end = len(verses)
	}

	return strings.Join(verses[start:end], "\n"), nil
}

func formatToStringForDBRequest(date time.Time) string {
	return date.Format("2006-01-02")
}

func scanIntoSong(rows *sql.Rows) ([]models.GetSong, error) {
	var songs []models.GetSong
	for rows.Next() {
		var songDetails models.GetSong
		var id int
		var releaseDate sql.NullTime
		var songText, songLink, song, group sql.NullString
		var create, update *time.Time
		if err := rows.Scan(
			&id, &song, &group, &releaseDate, &songText, &songLink, &create, &update,
		); err != nil {
			return nil, err
		}

		songDetails.Song = song.String
		songDetails.Group = group.String
		if releaseDate.Valid {
			songDetails.ReleaseDate = formatToStringForDBRequest(releaseDate.Time)
		} else {
			songDetails.ReleaseDate = ""
		}
		songDetails.SongText = songText.String
		songDetails.SongLink = songLink.String

		songs = append(songs, songDetails)
	}
	return songs, nil
}

func (s *PostgresStore) UpdateSong(songID int, songName, groupName string, releaseDate time.Time, songText, songLink string) error {
	query := `update songs 
set song_name = $2, 
    group_name = $3, 
    release_date = $4, 
    song_text = $5, 
    song_link = $6,
    updated_at = $7 
where song_id = $1;`
	_, err := s.db.Exec(query, songID, songName, groupName, formatToStringForDBRequest(releaseDate), songText, songLink, time.Now())
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

func (s *PostgresStore) DeleteSong(songID int) error {
	query := `delete from songs where song_id = $1`
	result, err := s.db.Exec(query, songID)
	if err != nil {
		return err
	}
	checkIfDeleted, _ := result.RowsAffected()
	if checkIfDeleted == 0 {
		return fmt.Errorf("no song found")
	}
	return nil
}

func (s *PostgresStore) GetFilteredSongsDataWithPagination(songName, groupName string, dateParsed time.Time, songText, songLink string, entriesPerPage, offset int) ([]models.GetSong, int, error) {

	query := `SELECT * FROM songs WHERE 1=1`
	var args []interface{}
	argCount := 1

	if songName != "" {
		query += fmt.Sprintf(" AND song_name ILIKE '%%' || $%d || '%%'", argCount)
		args = append(args, songName)
		argCount++
	}

	if groupName != "" {
		query += fmt.Sprintf(" AND group_name ILIKE '%%' || $%d || '%%'", argCount)
		args = append(args, groupName)
		argCount++
	}

	if !dateParsed.IsZero() {
		query += fmt.Sprintf(" AND release_date = $%d", argCount)
		args = append(args, dateParsed)
		argCount++
	}

	if songText != "" {
		query += fmt.Sprintf(" AND song_text ILIKE '%%' || $%d || '%%'", argCount)
		args = append(args, songText)
		argCount++
	}

	if songLink != "" {
		query += fmt.Sprintf(" AND song_link ILIKE '%%' || $%d || '%%'", argCount)
		args = append(args, songLink)
		argCount++
	}

	query += fmt.Sprintf(" LIMIT $%d OFFSET $%d", argCount, argCount+1)
	args = append(args, entriesPerPage, offset)

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, 0, err
	}
	defer rows.Close()

	songs, err := scanIntoSong(rows)
	if err != nil {
		return nil, 0, err
	}
	var entriesTotal int
	countQuery := `SELECT count(*) FROM songs`
	if err := s.db.QueryRow(countQuery).Scan(&entriesTotal); err != nil {
		return nil, 0, err
	}

	return songs, entriesTotal, nil
}

func (s *PostgresStore) CreateTableIfNotExists() error {
	query := `
	CREATE EXTENSION IF NOT EXISTS pgcrypto;
	CREATE TABLE IF NOT EXISTS songs (
		song_id SERIAL PRIMARY KEY,
		song_name VARCHAR(255) NOT NULL,
		group_name VARCHAR(255) NOT NULL,
		release_date DATE,
		song_text TEXT,
		song_link VARCHAR(255),
		created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
		updated_at TIMESTAMP DEFAULT NULL
	);`

	_, err := s.db.Exec(query)
	if err != nil {
		return fmt.Errorf("failed to create table: %w", err)
	}
	log.Println("Table 'songs' is ready.")
	return nil
}

func (s *PostgresStore) SeedDataIfEmpty() error {
	var count int
	err := s.db.QueryRow(`SELECT COUNT(*) FROM songs`).Scan(&count)
	if err != nil {
		return fmt.Errorf("failed to check if table is empty: %w", err)
	}

	if count == 0 {
		query := `
		INSERT INTO songs (song_name, group_name, release_date, song_text, song_link)
		VALUES 
		('Bohemian Rhapsody', 'Queen', '1975-10-31', 'Is this the real life?\nIs this just fantasy?\nCaught in a landslide,\nNo escape from reality.', 'https://example.com/bohemian-rhapsody'),
		('Imagine', 'John Lennon', '1971-10-11', 'Imagine there''s no heaven,\nIt''s easy if you try,\nNo hell below us,\nAbove us, only sky.', 'https://example.com/imagine'),
		('Hotel California', 'Eagles', '1976-12-08', 'On a dark desert highway,\nCool wind in my hair,\nWarm smell of colitas,\nRising up through the air.', 'https://example.com/hotel-california'),
		('Stairway to Heaven', 'Led Zeppelin', '1971-11-08', 'There''s a lady who''s sure,\nAll that glitters is gold,\nAnd she''s buying a stairway to heaven.', 'https://example.com/stairway-to-heaven'),
		('Smells Like Teen Spirit', 'Nirvana', '1991-09-10', 'Load up on guns, bring your friends,\nIt''s fun to lose and to pretend.', 'https://example.com/smells-like-teen-spirit'),
		('Billie Jean', 'Michael Jackson', '1983-01-02', 'She was more like a beauty queen,\nFrom a movie scene.', 'https://example.com/billie-jean'),
		('Wonderwall', 'Oasis', '1995-10-30', 'Today is gonna be the day,\nThat they''re gonna throw it back to you.', 'https://example.com/wonderwall'),
		('Hey Jude', 'The Beatles', '1968-08-26', 'Hey Jude, don''t make it bad,\nTake a sad song and make it better.', 'https://example.com/hey-jude'),
		('Like a Rolling Stone', 'Bob Dylan', '1965-07-20', 'Once upon a time you dressed so fine,\nThrew the bums a dime, in your prime, didn’t you?', 'https://example.com/like-a-rolling-stone'),
		('Purple Haze', 'Jimi Hendrix', '1967-03-17', 'Purple haze all in my brain,\nLately things just don''t seem the same.', 'https://example.com/purple-haze');`

		_, err = s.db.Exec(query)
		if err != nil {
			return fmt.Errorf("failed to seed data: %w", err)
		}
		log.Println("Data seeded successfully.")
	} else {
		log.Println("Table 'songs' already has data. No seeding required.")
	}

	return nil
}
