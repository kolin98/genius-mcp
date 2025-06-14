package main

import (
	"context"
	"log/slog"
	"net/url"
	"os"

	"github.com/kolin98/genius-mcp/internal/config"
	"github.com/kolin98/genius-mcp/internal/tools"
	"github.com/kolin98/genius-mcp/pkg/geniusapi"
	"github.com/kolin98/genius-mcp/pkg/geniuslyrics"
	"github.com/mark3labs/mcp-go/mcp"
	"github.com/mark3labs/mcp-go/server"
)

func main() {
	cfg, err := config.LoadConfig()
	if err != nil {
		slog.Error("Failed to load config", "error", err)
		os.Exit(1)
	}

	s := server.NewMCPServer(
		"Genius Lyrics",
		"0.1.0",
		server.WithToolCapabilities(false),
		server.WithLogging(),
	)

	callbackURL, _ := url.JoinPath(cfg.Host, "/callback") // nolint: errcheck
	geniusAPIClient := geniusapi.NewDefaultClient(
		cfg.GeniusClientID,
		cfg.GeniusClientSecret,
		callbackURL,
	)
	err = geniusAPIClient.Initialize()
	if err != nil {
		slog.Error("Failed to initialize Genius API client", "error", err)
		os.Exit(1)
	}

	serverTools := tools.GetTools(
		geniusAPIClient,
		geniuslyrics.NewDefaultClient(),
	)

	for _, tool := range serverTools {
		tool.Register(s)
	}

	s.AddNotificationHandler("notification", func(ctx context.Context, notification mcp.JSONRPCNotification) {
		slog.Info("Notification received", "method", notification.Method)
	})

	sseServer := server.NewSSEServer(
		s,
		server.WithBaseURL("http://localhost:5000"),
	)

	if err := sseServer.Start(":5000"); err != nil {
		slog.Error("Failed to start server", "error", err)
		os.Exit(1)
	}
}
