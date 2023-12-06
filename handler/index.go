package handler

import (
	"encoding/json"
	"net/http"
)

type Medias struct {
	Width  int    `json:"width,omitempty"`
	Height int    `json:"height,omitempty"`
	Source string `json:"src"`
}

type IndexedMedia struct {
	URL     string   `json:"url"`
	Medias  []Medias `json:"medias"`
	Caption string   `json:"caption,omitempty"`
}

func HandlerIndex(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(map[string]interface{}{
	})
}
