package api

import (
	"context"
	"etl/internal/config"
	"etl/internal/models"
	"fmt"
	"io"
	"net/http"
	"time"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
}

func NewClient(config config.APIConfig) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 30 * time.Second,
		},
		baseURL: config.ODataBaseURL,
	}
}

// fetch a specific document by identifier
// DocumentResponse contains the XML data and the URL it was fetched from
type DocumentResponse struct {
	XMLData []byte
	URL     string
}

func (c *Client) FetchDocument(ctx context.Context, kamerstukdossier models.Kamerstukdossier, volgnummer int) (*DocumentResponse, error) {
	if volgnummer == 0 {
		volgnummer = 1
	}

	docURL := c.buildDocumentURL(kamerstukdossier, volgnummer)
	xmlData, err := c.makeRequest(ctx, docURL)
	if err != nil {
		return nil, err
	}

	return &DocumentResponse{
		XMLData: xmlData,
		URL:     docURL,
	}, nil
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

// https://zoek.officielebekendmakingen.nl/kst-21501-02-3196.xml
// Getal: {nummer}-{toevoeging}-{motie.volgnummer}
func (c *Client) buildDocumentURL(kamerstukdossier models.Kamerstukdossier, volgnummer int) string {
	if kamerstukdossier.Toevoeging != nil && *kamerstukdossier.Toevoeging != "" {
		return fmt.Sprintf("https://zoek.officielebekendmakingen.nl/kst-%s-%s-%d.xml",
			kamerstukdossier.Nummer, *kamerstukdossier.Toevoeging, volgnummer)
	}
	return fmt.Sprintf("https://zoek.officielebekendmakingen.nl/kst-%s-%d.xml",
		kamerstukdossier.Nummer, volgnummer)
}
