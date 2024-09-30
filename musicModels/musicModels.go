package musicModels

type APIError struct {
	Error string `json:"error"`
}

type GetSongDetail struct {
	ReleaseDate string `json:"releaseDate"`
	Text        string `json:"text"`
	Link        string `json:"link"`
}

// type
