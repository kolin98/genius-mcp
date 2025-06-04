package geniusapi

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"net/url"

	"github.com/google/uuid"
	"golang.org/x/oauth2"
)

type Client interface {
	Initialize() error
	Search(query string) ([]Song, error)
}

type client struct {
	apiBaseURL  string
	httpClient  *http.Client
	accessToken string
	oauthConfig *oauth2.Config
}

func NewClient(
	apiBaseURL string,
	httpClient *http.Client,
	accessToken string,
	oauthConfig *oauth2.Config,
) Client {
	return &client{
		httpClient:  httpClient,
		apiBaseURL:  apiBaseURL,
		accessToken: accessToken,
		oauthConfig: oauthConfig,
	}
}

func NewDefaultClient(clientID, clientSecret, redirectURL string) Client {
	oauthConfig := &oauth2.Config{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		RedirectURL:  redirectURL,
		Scopes:       []string{},
		Endpoint: oauth2.Endpoint{
			AuthURL:  "https://api.genius.com/oauth/authorize",
			TokenURL: "https://api.genius.com/oauth/token",
		},
	}

	return NewClient(
		"https://api.genius.com",
		&http.Client{},
		"",
		oauthConfig,
	)
}

func (c *client) Initialize() error {
	callbackURL, err := url.Parse(c.oauthConfig.RedirectURL)
	if err != nil {
		return fmt.Errorf("failed to parse redirect URL: %w", err)
	}

	codeChan := make(chan string)

	http.HandleFunc(callbackURL.Path, func(w http.ResponseWriter, r *http.Request) {
		fmt.Println("Callback received")
		code := r.URL.Query().Get("code")
		if code == "" {
			http.Error(w, "Code not found", http.StatusBadRequest)
			return
		}
		codeChan <- code
		fmt.Fprintf(w, "Authentication successful! You can close this window.")
	})

	go func() {
		if err := http.ListenAndServe(callbackURL.Host, nil); err != nil {
			log.Fatal(err)
		}
	}()

	authURL := c.oauthConfig.AuthCodeURL(uuid.New().String())
	fmt.Printf("Open this URL in your browser: %s\n", authURL)

	code := <-codeChan

	token, err := c.oauthConfig.Exchange(context.Background(), code)
	if err != nil {
		return fmt.Errorf("failed to exchange code for token: %w", err)
	}

	c.accessToken = token.AccessToken

	fmt.Printf("Access token: %s\n", c.accessToken)

	return nil
}

func (c *client) Search(query string) ([]Song, error) {
	if c.accessToken == "" {
		return nil, fmt.Errorf("client is not initialized")
	}

	urlStr, err := url.JoinPath(c.apiBaseURL, "search")
	if err != nil {
		return nil, fmt.Errorf("base url is malformed: %w", err)
	}

	url, _ := url.Parse(urlStr) // nolint: errcheck
	queryParams := url.Query()
	queryParams.Set("q", query)
	url.RawQuery = queryParams.Encode()

	req, err := http.NewRequest("GET", url.String(), nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", c.accessToken))

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("got error response: %d: %w", resp.StatusCode, err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("failed to read body: %w", err)
	}

	var response Response
	if err := json.Unmarshal(body, &response); err != nil {
		return nil, fmt.Errorf("failed to unmarshal response: %w", err)
	}

	var searchHits SearchHits
	if err := json.Unmarshal(response.Response, &searchHits); err != nil {
		return nil, fmt.Errorf("failed to unmarshal search hits: %w", err)
	}

	songs := make([]Song, 0)
	for _, hit := range searchHits.Hits {
		if hit.Type == "song" {
			var song Song
			if err := json.Unmarshal(hit.Result, &song); err != nil {
				return nil, fmt.Errorf("failed to unmarshal song: %w", err)
			}
			songs = append(songs, song)
		}
	}

	return songs, nil
}
