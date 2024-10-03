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
	UpdateSong(int, string, string, time.Time, string, string) error
	AddSong(string, string) error
	DeleteSong(int) error
	GetFilteredSongsDataWithPagination(string, string, time.Time, string, string, int, int) ([]models.GetSong, int, error)
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
			songDetails.ReleaseDate = convertToTimeAndFormat(releaseDate.Time)
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

func convertToTimeAndFormat(dateStr time.Time) string {
	// parsedTime, _ := time.Parse(time.RFC3339, dateStr)

	formattedDate := dateStr.Format("02.01.2006")
	return formattedDate
}
