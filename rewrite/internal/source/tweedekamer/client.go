package tweedekamer

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type Time struct {
	time.Time
}

func (value *Time) UnmarshalJSON(data []byte) error {
	if string(data) == "null" {
		return nil
	}

	var raw string
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	if raw == "" {
		return nil
	}

	for _, layout := range []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02",
	} {
		parsed, err := time.Parse(layout, raw)
		if err == nil {
			value.Time = parsed
			return nil
		}
	}

	return fmt.Errorf("unsupported time format %q", raw)
}

type Client struct {
	baseURL    string
	httpClient *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
	}
}

type MotionRecord struct {
	ID             string  `json:"Id"`
	Nummer         *string `json:"Nummer"`
	Onderwerp      *string `json:"Onderwerp"`
	Soort          *string `json:"Soort"`
	Titel          *string `json:"Titel"`
	Citeertitel    *string `json:"Citeertitel"`
	Status         *string `json:"Status"`
	GestartOp      *Time   `json:"GestartOp"`
	Vergaderjaar   *string `json:"Vergaderjaar"`
	GewijzigdOp    *Time   `json:"GewijzigdOp"`
	ApiGewijzigdOp *Time   `json:"ApiGewijzigdOp"`
	Verwijderd     *bool   `json:"Verwijderd"`
	Raw            json.RawMessage
}

func (record *MotionRecord) UnmarshalJSON(data []byte) error {
	type alias MotionRecord
	var decoded alias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}

	*record = MotionRecord(decoded)
	record.Raw = append(record.Raw[:0], data...)
	return nil
}

type ChangedMotionsPage struct {
	Records []MotionRecord
	NextURL string
}

func (client *Client) FetchChangedMotions(ctx context.Context, since time.Time, top int, nextURL string) (ChangedMotionsPage, error) {
	requestURL := nextURL
	if requestURL == "" {
		requestURL = client.changedMotionsURL(since, top)
	}

	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return ChangedMotionsPage{}, err
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", "partijgedrag-rewrite/0.1")

	response, err := client.httpClient.Do(request)
	if err != nil {
		return ChangedMotionsPage{}, err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return ChangedMotionsPage{}, fmt.Errorf("tweede kamer odata returned %d for %s", response.StatusCode, requestURL)
	}

	var body struct {
		Value   []MotionRecord `json:"value"`
		NextURL string         `json:"@odata.nextLink"`
	}
	if err := json.NewDecoder(response.Body).Decode(&body); err != nil {
		return ChangedMotionsPage{}, err
	}

	return ChangedMotionsPage{
		Records: body.Value,
		NextURL: body.NextURL,
	}, nil
}

func (client *Client) changedMotionsURL(since time.Time, top int) string {
	u, _ := url.Parse(client.baseURL + "/Zaak")
	query := u.Query()
	query.Set("$filter", fmt.Sprintf("Soort eq 'Motie' and ApiGewijzigdOp ge %s", formatODataDate(since)))
	query.Set("$select", strings.Join([]string{
		"Id",
		"Nummer",
		"Onderwerp",
		"Soort",
		"Titel",
		"Citeertitel",
		"Status",
		"GestartOp",
		"Vergaderjaar",
		"GewijzigdOp",
		"ApiGewijzigdOp",
		"Verwijderd",
	}, ","))
	query.Set("$orderby", "ApiGewijzigdOp asc,Id asc")
	query.Set("$top", fmt.Sprintf("%d", top))
	query.Set("$count", "false")
	u.RawQuery = query.Encode()
	return u.String()
}

func formatODataDate(value time.Time) string {
	return value.UTC().Format("2006-01-02T15:04:05Z")
}
