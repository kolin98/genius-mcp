package models

import "time"

type Song struct {
	ID          int       `json:"id"`
	Path        string    `json:"path"`
	Artist      string    `json:"artist_names"`
	Title       string    `json:"title"`
	ReleaseDate time.Time `json:"release_date"`
	Description string    `json:"description,omitempty"`
}
