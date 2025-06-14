package models

import (
	"time"

	"github.com/kolin98/genius-mcp/pkg/geniusapi"
)

func MapSong(song geniusapi.Song) Song {
	// October 28, 2002
	releaseDate, err := time.Parse("January 2, 2006", song.ReleaseDate)
	if err != nil {
		releaseDate = time.Time{}
	}

	return Song{
		ID:          song.ID,
		Path:        song.Path,
		Artist:      song.Artist,
		Title:       song.Title,
		ReleaseDate: releaseDate,
	}
}

func MapSongFull(song geniusapi.SongFull) Song {
	mappedSong := MapSong(song.Song)
	mappedSong.Description = song.Description.Plain

	return mappedSong
}
