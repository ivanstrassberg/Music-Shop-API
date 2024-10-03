package api

import (
	"encoding/json"
	"fmt"
	"log"
	database "musicShopBackend/database"
	models "musicShopBackend/musicModels"
	"net/http"
	"strconv"
	"time"
)

type APIServer struct {
	listenAddr string
	storage    database.Storage
}

func NewAPIServer(listedAddr string, store database.Storage) *APIServer {
	return &APIServer{
		listenAddr: listedAddr,
		storage:    store,
	}
}

func (s *APIServer) Run() {
	mux := http.NewServeMux()
	mux.HandleFunc("GET /info", makeHTTPHandleFunc(s.handleGetSongsWithPagination))
	// mux.HandleFunc("GET /info/{group}/{song}", makeHTTPHandleFunc(s.handleGetSong))
	mux.HandleFunc("POST /info", makeHTTPHandleFunc(s.handlePostSong))
	mux.HandleFunc("PUT /info", makeHTTPHandleFunc(s.handleUpdateSong))
	mux.HandleFunc("DELETE /info/{id}", makeHTTPHandleFunc(s.handleDeleteSong))
	log.Println("Server started on port", s.listenAddr)
	if err := http.ListenAndServe(s.listenAddr, mux); err != nil {
		fmt.Errorf("Failed to start a server on port %s", s.listenAddr)
	}

}

// func (s *APIServer) handleGetSong(w http.ResponseWriter, r *http.Request) error {
// 	group, song := r.PathValue("group"), r.PathValue("song")
// 	// non ASCII Characters handling just for fun
// 	decodedGroup, err := url.QueryUnescape(group)
// 	if err != nil {
// 		WriteJSON(w, http.StatusBadRequest, "bad request")
// 		return nil
// 	}

// 	decodedSong, err := url.QueryUnescape(song)
// 	if err != nil {
// 		WriteJSON(w, http.StatusBadRequest, "bad request")
// 		return nil
// 	}

// 	songDetail, err := s.storage.GetSong(decodedGroup, decodedSong)
// 	if err != nil {
// 		WriteJSON(w, http.StatusInternalServerError, "internal server error", err)
// 		return nil
// 	}
// 	if songDetail.ReleaseDate == "" {
// 		WriteJSON(w, http.StatusNotFound, "not found")
// 		return nil
// 	}
// 	WriteJSON(w, http.StatusOK, songDetail)
// 	return nil
// }

func (s *APIServer) handleUpdateSong(w http.ResponseWriter, r *http.Request) error {
	updateSong := new(models.UpdateSong)
	if err := json.NewDecoder(r.Body).Decode(updateSong); err != nil {
		WriteJSON(w, http.StatusBadRequest, "bad request")
		return nil
	}

	parsedDate, err := parseDate(updateSong.ReleaseDate)
	if err != nil {
		WriteJSON(w, http.StatusBadRequest, "bad request")
	}

	if err := s.storage.UpdateSong(
		updateSong.SongID, updateSong.Song,
		updateSong.Group, parsedDate, updateSong.SongText,
		updateSong.SongLink); err != nil {
		WriteJSON(w, http.StatusInternalServerError, "internal server error")
		return nil
	}

	WriteJSON(w, http.StatusOK, "song updated")
	return nil
}

func (s *APIServer) handlePostSong(w http.ResponseWriter, r *http.Request) error {
	addSong := new(models.AddSong)
	if err := json.NewDecoder(r.Body).Decode(addSong); err != nil {
		WriteJSON(w, http.StatusBadRequest, "bad request")
		return nil
	}
	if err := s.storage.AddSong(addSong.Song, addSong.Group); err != nil {
		WriteJSON(w, http.StatusInternalServerError, "internal server error")
		return nil
	}
	WriteJSON(w, http.StatusOK, "song created")
	return nil
}

func (s *APIServer) handleDeleteSong(w http.ResponseWriter, r *http.Request) error {
	strSongID := r.PathValue("id")
	songID, err := strconv.Atoi(strSongID)
	if err != nil {
		WriteJSON(w, http.StatusBadRequest, "bad request")
		return nil
	}
	if err := s.storage.DeleteSong(songID); err != nil {
		WriteJSON(w, http.StatusInternalServerError, "internal server error")
		return nil
	}
	WriteJSON(w, http.StatusNoContent, "deleted")
	return nil
}

func (s *APIServer) handleGetSongsWithPagination(w http.ResponseWriter, r *http.Request) error {
	songName := r.URL.Query().Get("song_name")
	groupName := r.URL.Query().Get("group_name")
	releaseDate := r.URL.Query().Get("release_date")
	songText := r.URL.Query().Get("song_text")
	songLink := r.URL.Query().Get("song_link")
	pageStr := r.URL.Query().Get("pages")
	entriesStr := r.URL.Query().Get("entries")

	page, err := strconv.Atoi(pageStr)
	if err != nil || page < 1 {
		page = 1
	}

	entriesPerPage, err := strconv.Atoi(entriesStr)
	if err != nil || entriesPerPage < 1 {
		entriesPerPage = 10
	}
	offset := (page - 1) * entriesPerPage
	if songName == "" {
		songName = ""
	}
	if groupName == "" {
		groupName = ""
	}
	if songText == "" {
		songText = ""
	}
	if songLink == "" {
		songLink = ""
	}

	dateParsed, err := parseDate(releaseDate)
	if err != nil {
		WriteJSON(w, http.StatusBadRequest, "bad request")
		return err
	}
	songs, entriesTotal, err := s.storage.GetFilteredSongsDataWithPagination(songName, groupName, dateParsed, songText, songLink, entriesPerPage, offset)
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, "internal server error")
		return err
	}
	pagesTotal := (entriesTotal + entriesPerPage - 1) / entriesPerPage
	response := (models.PaginatedFilteredResults{
		Page:           page,
		PagesTotal:     pagesTotal,
		EntriesTotal:   entriesTotal,
		EntriesPerPage: entriesPerPage,
		Songs:          songs,
	})
	WriteJSON(w, http.StatusOK, response)
	return nil
}

// Auxiliary functions.

func makeHTTPHandleFunc(f apiFunc) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := f(w, r); err != nil {
			WriteJSON(w, http.StatusBadRequest, models.APIError{Error: err.Error()})
		}
	}
}

func WriteJSON(w http.ResponseWriter, status int, msg any, err ...error) {
	w.Header().Add("Content-Type", "application/json")
	w.WriteHeader(status)
	json.NewEncoder(w).Encode(msg)
	// log.Println(err)
}

type apiFunc func(http.ResponseWriter, *http.Request) error

func parseDate(dateToParse string) (time.Time, error) {
	if dateToParse == "" {
		return time.Time{}, nil
	}
	inputLayout := "02.01.2006"
	parsedDate, err := time.Parse(inputLayout, dateToParse)
	if err != nil {
		return time.Time{}, err
	}
	return parsedDate, nil
}
