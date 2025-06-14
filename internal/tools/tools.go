package tools

import (
	"github.com/kolin98/genius-mcp/internal/handlers"
	"github.com/kolin98/genius-mcp/pkg/geniusapi"
	"github.com/kolin98/genius-mcp/pkg/geniuslyrics"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

type Tool interface {
	Register(server *server.MCPServer)
}

type tool struct {
	tool    mcp.Tool
	handler server.ToolHandlerFunc
}

func (t *tool) Register(server *server.MCPServer) {
	server.AddTool(t.tool, t.handler)
}

func GetTools(apiClient geniusapi.Client, lyricsClient geniuslyrics.Client) []Tool {
	return []Tool{
		&tool{
			tool: mcp.NewTool("find_song",
				mcp.WithDescription(`
					Find a song by its title, artist, album, description or lyrics.
					Returns a list of songs that match the search query, from most
					relevant to least relevant.
					Among others it provides the song ID and path that can be used
					in other tools.
				`),
				mcp.WithString("query",
					mcp.Required(),
					mcp.Description("Search query for the song"),
				),
			),
			handler: handlers.NewFindSongHandler(apiClient),
		},
		&tool{
			tool: mcp.NewTool("get_song_details",
				mcp.WithDescription(`
					Get details about a song by its ID.
					Returns the song details, including the title, artist, album,
					release date, and detailed description.
				`),
				mcp.WithNumber("id",
					mcp.Required(),
					mcp.Description("The ID of the song to get details about. The id can be retrieved from the find_song tool."),
				),
			),
			handler: handlers.NewSongDetailsHandler(apiClient),
		},
		&tool{
			tool: mcp.NewTool("get_song_lyrics",
				mcp.WithDescription(`
					Get the lyrics of a song by its path.
					Returns the lyrics of the song in string format.
				`),
				mcp.WithString("path",
					mcp.Required(),
					mcp.Description("The path of the song to get lyrics for. The path can be retrieved from the find_song or get_song_details tool."),
				),
			),
			handler: handlers.NewLyricsHandler(lyricsClient),
		},
	}
}
