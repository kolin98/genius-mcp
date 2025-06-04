package geniusapi

import "encoding/json"

type Response struct {
	Response json.RawMessage `json:"response"`
}

type Song struct {
	ID          int    `json:"id"`
	Path        string `json:"path"`
	Artist      string `json:"artist_names"`
	Title       string `json:"title"`
	ReleaseDate string `json:"release_date_for_display"`
}

type SongFull struct {
	Song
	Description struct {
		Plain string `json:"plain"`
	} `json:"description"`
}

type SearchHit struct {
	Type   string          `json:"type"`
	Result json.RawMessage `json:"result"`
}

type SearchHits struct {
	Hits []SearchHit `json:"hits"`
}
