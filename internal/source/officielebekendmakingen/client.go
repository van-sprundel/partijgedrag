package officielebekendmakingen

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// ErrNotFound is returned when no publication exists at the derived URL.
var ErrNotFound = errors.New("officielebekendmakingen: document not found")

type Client struct {
	httpClient *http.Client
}

func NewClient() *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

// DocumentURL builds the kamerstuk XML URL for a motion document,
// e.g. https://zoek.officielebekendmakingen.nl/kst-36045-299.xml
// or with a dossier toevoeging https://zoek.officielebekendmakingen.nl/kst-21501-02-3196.xml
func DocumentURL(dossierNummer string, dossierToevoeging string, volgnummer int) string {
	if dossierToevoeging != "" {
		return fmt.Sprintf("https://zoek.officielebekendmakingen.nl/kst-%s-%s-%d.xml", dossierNummer, dossierToevoeging, volgnummer)
	}
	return fmt.Sprintf("https://zoek.officielebekendmakingen.nl/kst-%s-%d.xml", dossierNummer, volgnummer)
}

func (client *Client) FetchDocument(ctx context.Context, documentURL string) ([]byte, error) {
	var lastErr error
	for attempt := 1; attempt <= 3; attempt++ {
		body, err := client.fetchOnce(ctx, documentURL)
		if err == nil || errors.Is(err, ErrNotFound) {
			return body, err
		}
		lastErr = err

		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(time.Duration(attempt) * time.Second):
		}
	}
	return nil, lastErr
}

func (client *Client) fetchOnce(ctx context.Context, documentURL string) ([]byte, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, documentURL, nil)
	if err != nil {
		return nil, err
	}
	request.Header.Set("User-Agent", "partijgedrag-rewrite/0.1")

	response, err := client.httpClient.Do(request)
	if err != nil {
		return nil, err
	}
	defer response.Body.Close()

	if response.StatusCode == http.StatusNotFound {
		return nil, ErrNotFound
	}
	if response.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(response.Body, 2048))
		return nil, fmt.Errorf("officielebekendmakingen returned %d for %s: %s", response.StatusCode, documentURL, strings.TrimSpace(string(body)))
	}

	return io.ReadAll(response.Body)
}
