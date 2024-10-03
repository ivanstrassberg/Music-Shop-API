package musicModels

import "time"

type APIError struct {
	Error string `json:"error"`
}

type GetSongAdmin struct {
	SongID      int        `json:"song_id"`
	Song        string     `json:"song"`
	Group       string     `json:"group"`
	ReleaseDate string     `json:"release_date"`
	SongText    string     `json:"song_text"`
	SongLink    string     `json:"song_link"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
}

type PaginatedFilteredResults struct {
	Page           int       `json:"page"`
	PagesTotal     int       `json:"pages_total"`
	EntriesTotal   int       `json:"entries_total"`
	EntriesPerPage int       `json:"entries_per_page"`
	Songs          []GetSong `json:"songs"`
}

type GetSong struct {
	Song        string `json:"song"`
	Group       string `json:"group"`
	ReleaseDate string `json:"release_date"`
	SongText    string `json:"song_text"`
	SongLink    string `json:"song_link"`
}

type UpdateSong struct {
	SongID      int    `json:"song_id"`
	Song        string `json:"song"`
	Group       string `json:"group"`
	ReleaseDate string `json:"release_date"`
	SongText    string `json:"song_text"`
	SongLink    string `json:"song_link"`
}

type AddSong struct {
	Group string `json:"group"`
	Song  string `json:"song"`
}
