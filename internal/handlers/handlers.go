package handlers

import (
	"context"
	"encoding/json"

	"github.com/kolin98/genius-mcp/internal/models"
	"github.com/kolin98/genius-mcp/pkg/geniusapi"
	"github.com/kolin98/genius-mcp/pkg/geniuslyrics"
	"github.com/mark3labs/mcp-go/mcp"
)

func NewFindSongHandler(apiClient geniusapi.Client) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		query, err := request.RequireString("query")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		result, err := apiClient.Search(query)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		songs := make([]models.Song, len(result))
		for i, song := range result {
			songs[i] = models.MapSong(song)
		}

		songsJSON, err := json.Marshal(songs)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(string(songsJSON)), nil
	}
}

func NewSongDetailsHandler(apiClient geniusapi.Client) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		id, err := request.RequireInt("id")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		result, err := apiClient.GetSong(id)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		song := models.MapSongFull(*result)

		songJSON, err := json.Marshal(song)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(string(songJSON)), nil
	}
}

func NewLyricsHandler(lyricsClient geniuslyrics.Client) func(context.Context, mcp.CallToolRequest) (*mcp.CallToolResult, error) {
	return func(ctx context.Context, request mcp.CallToolRequest) (*mcp.CallToolResult, error) {
		path, err := request.RequireString("path")
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		lyrics, err := lyricsClient.GetLyrics(path)
		if err != nil {
			return mcp.NewToolResultError(err.Error()), nil
		}

		return mcp.NewToolResultText(lyrics), nil
	}
}
