package api

import (
	"encoding/json"
	"fmt"
	"log"
	database "musicShopBackend/database"
	models "musicShopBackend/musicModels"
	"net/http"
	"net/url"
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
	mux.HandleFunc("GET /info/{group}/{song}", makeHTTPHandleFunc(s.handleInfo))

	log.Println("Server started on port", s.listenAddr)
	if err := http.ListenAndServe(s.listenAddr, mux); err != nil {
		fmt.Errorf("Failed to start a server on port %s", s.listenAddr)
	}

}

func (s *APIServer) handleInfo(w http.ResponseWriter, r *http.Request) error {
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
	WriteJSON(w, http.StatusOK, songDetail)
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
