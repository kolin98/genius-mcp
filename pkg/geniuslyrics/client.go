package geniuslyrics

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/PuerkitoBio/goquery"
	"github.com/dop251/goja"
)

type Client interface {
	GetLyrics(geniusPath string) (string, error)
}

type RequestConstructor func(path string) (*http.Request, error)

type client struct {
	requestConstructor RequestConstructor
	httpClient         *http.Client
	vm                 *goja.Runtime
}

func NewClient(
	requestConstructor RequestConstructor,
	httpClient *http.Client,
	vm *goja.Runtime,
) Client {
	return &client{
		requestConstructor: requestConstructor,
		httpClient:         httpClient,
		vm:                 vm,
	}
}

func NewDefaultClient() Client {
	return NewClient(
		DefaultRequestConstructor,
		&http.Client{},
		goja.New(),
	)
}

func DefaultRequestConstructor(path string) (*http.Request, error) {
	url, err := url.JoinPath("https://genius.com", path)
	if err != nil {
		return nil, fmt.Errorf("failed to join path: %w", err)
	}

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to construct request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/91.0.4472.124 Safari/537.36")
	req.Header.Set("Referer", "https://genius.com")

	return req, nil
}

func (c *client) GetLyrics(geniusPath string) (string, error) {
	req, err := c.requestConstructor(geniusPath)
	if err != nil {
		return "", err
	}

	html, err := c.getHTML(req)
	if err != nil {
		return "", err
	}

	stateScript, err := c.getStateScript(html)
	if err != nil {
		return "", err
	}

	lyrics, err := c.getLyricsFromScript(stateScript)
	if err != nil {
		return "", err
	}

	return lyrics, nil
}

func (c *client) getHTML(req *http.Request) (string, error) {
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("got error response: %d: %w", resp.StatusCode, err)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read body: %w", err)
	}

	return string(body), nil
}

func (c *client) getStateScript(html string) (string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return "", fmt.Errorf("failed to create document: %w", err)
	}

	var stateScript string
	doc.Find("script").Each(func(i int, s *goquery.Selection) {
		if strings.Contains(s.Text(), "__PRELOADED_STATE__") {
			stateScript = s.Text()
		}
	})

	if stateScript == "" {
		return "", fmt.Errorf("preloaded state script not found")
	}

	return stateScript, nil
}

func (c *client) getLyricsFromScript(stateScript string) (string, error) {
	stateScript = strings.ReplaceAll(stateScript, "window.", "")

	_, err := c.vm.RunString(stateScript)
	if err != nil {
		return "", fmt.Errorf("failed to execute state script: %w", err)
	}

	v, err := c.vm.RunString(`__PRELOADED_STATE__.songPage.lyricsData.body.html`)
	if err != nil {
		return "", fmt.Errorf("failed to execute lyrics data script: %w", err)
	}

	lyricsHTML, ok := v.Export().(string)
	if !ok {
		return "", fmt.Errorf("lyrics data is not a string")
	}
	lyrics, err := c.parseLyrics(lyricsHTML)
	if err != nil {
		return "", err
	}

	return lyrics, nil
}

func (c *client) parseLyrics(html string) (string, error) {
	doc, err := goquery.NewDocumentFromReader(strings.NewReader(html))
	if err != nil {
		return "", fmt.Errorf("failed to create document from lyrics data: %w", err)
	}

	lyrics := make([]string, 0)
	doc.Each(func(i int, s *goquery.Selection) {
		if s.Text() != "" {
			lyrics = append(lyrics, s.Text())
		}
	})
	lyricsJoined := strings.Join(lyrics, "\n")

	return lyricsJoined, nil
}
