package api

import (
	"encoding/json"
	"fmt"
	"log"
	database "musicShopBackend/database"
	models "musicShopBackend/musicModels"
	"net/http"
	"net/url"
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
	mux.HandleFunc("GET /info/{group}/{song}", makeHTTPHandleFunc(s.handleGetSong))
	mux.HandleFunc("POST /info", makeHTTPHandleFunc(s.handlePostSong))
	mux.HandleFunc("PUT /info", makeHTTPHandleFunc(s.handleUpdateSong)) // patch maybe?
	log.Println("Server started on port", s.listenAddr)
	if err := http.ListenAndServe(s.listenAddr, mux); err != nil {
		fmt.Errorf("Failed to start a server on port %s", s.listenAddr)
	}

}

func (s *APIServer) handleGetSong(w http.ResponseWriter, r *http.Request) error {
	group, song := r.PathValue("group"), r.PathValue("song")
	// non ASCII Characters handling just for fun
	decodedGroup, err := url.QueryUnescape(group)
	if err != nil {
		WriteJSON(w, http.StatusBadRequest, "bad request")
		return nil
	}

	decodedSong, err := url.QueryUnescape(song)
	if err != nil {
		WriteJSON(w, http.StatusBadRequest, "bad request")
		return nil
	}

	songDetail, err := s.storage.GetSong(decodedGroup, decodedSong)
	if err != nil {
		WriteJSON(w, http.StatusInternalServerError, "internal server error", err)
		return nil
	}
	if songDetail.ReleaseDate == "" {
		WriteJSON(w, http.StatusNotFound, "not found")
		return nil
	}
	WriteJSON(w, http.StatusOK, songDetail)
	return nil
}

func (s *APIServer) handleUpdateSong(w http.ResponseWriter, r *http.Request) error {
	updateSong := new(models.UpdateSong)
	if err := json.NewDecoder(r.Body).Decode(updateSong); err != nil {
		WriteJSON(w, http.StatusBadRequest, "bad request")
		return nil
	}

	inputLayout := "02.01.2006"
	parsedDate, err := time.Parse(inputLayout, updateSong.ReleaseDate)
	if err != nil {
		WriteJSON(w, http.StatusBadRequest, "invalid date format, expected DD.MM.YYYY")
		return err
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
	log.Println(err)
}

type apiFunc func(http.ResponseWriter, *http.Request) error
