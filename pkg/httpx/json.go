package httpx

import (
	"encoding/json"
	"log"
	"net/http"
)

func WriteJSON(w http.ResponseWriter, status int, payload any) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	if payload == nil {
		return
	}
	err := json.NewEncoder(w).Encode(payload)
	if err != nil {
		log.Printf("WriteJSON: unexpected error: %v", err)
		http.Error(w, "Internal server error", http.StatusInternalServerError)
	}
}
