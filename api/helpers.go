package api

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
)

var (
	ErrInternal = errors.New("internal error")
)

func logError(err error) {
	if err != nil {
		log.Printf("error: %v", err)
	}
}

func writeJSON(w http.ResponseWriter, data interface{}, statusCode int) error {
	w.Header().Add("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(statusCode)
	err := json.NewEncoder(w).Encode(data)
	return err
}

func writeError(w http.ResponseWriter, message string, statusCode int) error {
	log.Printf("response errored: %s", message)
	return writeJSON(
		w,
		map[string]string{"error": message},
		statusCode,
	)
}
