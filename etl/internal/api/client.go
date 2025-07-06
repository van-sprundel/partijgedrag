package api

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

const (
	BaseURL = "https://gegevensmagazijn.tweedekamer.nl/SyncFeed/2.0/Feed"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: BaseURL,
	}
}

func (c *Client) FetchFeed(ctx context.Context, category string, skiptoken string) ([]byte, error) {
	params := url.Values{}
	params.Add("category", category)

	if skiptoken != "" {
		params.Add("skiptoken", skiptoken)
	}

	feedURL := c.baseURL + "?" + params.Encode()
	return c.makeRequest(ctx, feedURL)
}

// fetch a specific document by identifier (kst-[nummer][-volgnummer])
func (c *Client) FetchDocument(ctx context.Context, nummer string, toevoeging string, volgnummer int) ([]byte, error) {
	if volgnummer == 0 {
		volgnummer = 1
	}

	docURL := c.buildDocumentURL(nummer, toevoeging, volgnummer)
	return c.makeRequest(ctx, docURL)
}

func (c *Client) makeRequest(ctx context.Context, url string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("User-Agent", "xml-importer/1.0")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("making request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	return body, nil
}

func (c *Client) buildDocumentURL(nummer string, toevoeging string, volgnummer int) string {
	if toevoeging != "" {
		return fmt.Sprintf("https://zoek.officielebekendmakingen.nl/kst-%s-%s-%d.xml",
			nummer, toevoeging, volgnummer)
	}
	return fmt.Sprintf("https://zoek.officielebekendmakingen.nl/kst-%s-%d.xml",
		nummer, volgnummer)
}
