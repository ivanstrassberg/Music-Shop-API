package musicModels

import "time"

type APIError struct {
	Error string `json:"error"`
}

type Song struct {
	UUID        string     `json:"uuid"`
	SongName    string     `json:"song"`
	GroupName   string     `json:"group"`
	ReleaseDate string     `json:"release_date"`
	SongText    string     `json:"song_text"`
	SongLink    string     `json:"song_link"`
	CreatedAt   time.Time  `json:"created_at"`
	UpdatedAt   *time.Time `json:"updated_at"`
}

type GetSongDetail struct {
	ReleaseDate string `json:"releaseDate"`
	Text        string `json:"text"`
	Link        string `json:"link"`
}

type UpdateSong struct {
	SongKey     string `json:"song_name_key"`
	GroupKey    string `json:"group_name_key"`
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
