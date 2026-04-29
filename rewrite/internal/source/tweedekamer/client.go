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
		time.RFC3339Nano,
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

type DecisionRecord struct {
	ID                            string  `json:"Id"`
	AgendapuntID                  *string `json:"Agendapunt_Id"`
	StemmingsSoort                *string `json:"StemmingsSoort"`
	BesluitSoort                  *string `json:"BesluitSoort"`
	BesluitTekst                  *string `json:"BesluitTekst"`
	Opmerking                     *string `json:"Opmerking"`
	Status                        *string `json:"Status"`
	AgendapuntZaakBesluitVolgorde *int    `json:"AgendapuntZaakBesluitVolgorde"`
	GewijzigdOp                   *Time   `json:"GewijzigdOp"`
	ApiGewijzigdOp                *Time   `json:"ApiGewijzigdOp"`
	Verwijderd                    *bool   `json:"Verwijderd"`
	Raw                           json.RawMessage
}

func (record *DecisionRecord) UnmarshalJSON(data []byte) error {
	type alias DecisionRecord
	var decoded alias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}

	*record = DecisionRecord(decoded)
	record.Raw = append(record.Raw[:0], data...)
	return nil
}

type VoteRecord struct {
	ID              string  `json:"Id"`
	BesluitID       *string `json:"Besluit_Id"`
	Soort           *string `json:"Soort"`
	FractieGrootte  *int    `json:"FractieGrootte"`
	ActorNaam       *string `json:"ActorNaam"`
	ActorFractie    *string `json:"ActorFractie"`
	Vergissing      *bool   `json:"Vergissing"`
	SidActorLid     *string `json:"SidActorLid"`
	SidActorFractie *string `json:"SidActorFractie"`
	PersoonID       *string `json:"Persoon_Id"`
	FractieID       *string `json:"Fractie_Id"`
	GewijzigdOp     *Time   `json:"GewijzigdOp"`
	ApiGewijzigdOp  *Time   `json:"ApiGewijzigdOp"`
	Verwijderd      *bool   `json:"Verwijderd"`
	Raw             json.RawMessage
}

func (record *VoteRecord) UnmarshalJSON(data []byte) error {
	type alias VoteRecord
	var decoded alias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}

	*record = VoteRecord(decoded)
	record.Raw = append(record.Raw[:0], data...)
	return nil
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

func (client *Client) FetchMotionDecisions(ctx context.Context, motionSourceID string) ([]DecisionRecord, error) {
	requestURL := client.motionDecisionsURL(motionSourceID)
	var records []DecisionRecord

	for requestURL != "" {
		var body struct {
			Value   []DecisionRecord `json:"value"`
			NextURL string           `json:"@odata.nextLink"`
		}
		if err := client.fetchJSON(ctx, requestURL, &body); err != nil {
			return nil, err
		}

		records = append(records, body.Value...)
		requestURL = body.NextURL
	}

	return records, nil
}

func (client *Client) FetchDecisionVotes(ctx context.Context, decisionSourceID string) ([]VoteRecord, error) {
	requestURL := client.decisionVotesURL(decisionSourceID)
	var records []VoteRecord

	for requestURL != "" {
		var body struct {
			Value   []VoteRecord `json:"value"`
			NextURL string       `json:"@odata.nextLink"`
		}
		if err := client.fetchJSON(ctx, requestURL, &body); err != nil {
			return nil, err
		}

		records = append(records, body.Value...)
		requestURL = body.NextURL
	}

	return records, nil
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

func (client *Client) motionDecisionsURL(motionSourceID string) string {
	u, _ := url.Parse(fmt.Sprintf("%s/Zaak(%s)/Besluit", client.baseURL, motionSourceID))
	query := u.Query()
	query.Set("$select", strings.Join([]string{
		"Id",
		"Agendapunt_Id",
		"StemmingsSoort",
		"BesluitSoort",
		"BesluitTekst",
		"Opmerking",
		"Status",
		"AgendapuntZaakBesluitVolgorde",
		"GewijzigdOp",
		"ApiGewijzigdOp",
		"Verwijderd",
	}, ","))
	query.Set("$orderby", "AgendapuntZaakBesluitVolgorde asc,ApiGewijzigdOp asc")
	u.RawQuery = query.Encode()
	return u.String()
}

func (client *Client) decisionVotesURL(decisionSourceID string) string {
	u, _ := url.Parse(fmt.Sprintf("%s/Besluit(%s)/Stemming", client.baseURL, decisionSourceID))
	query := u.Query()
	query.Set("$select", strings.Join([]string{
		"Id",
		"Besluit_Id",
		"Soort",
		"FractieGrootte",
		"ActorNaam",
		"ActorFractie",
		"Vergissing",
		"SidActorLid",
		"SidActorFractie",
		"Persoon_Id",
		"Fractie_Id",
		"GewijzigdOp",
		"ApiGewijzigdOp",
		"Verwijderd",
	}, ","))
	query.Set("$orderby", "ApiGewijzigdOp asc,Id asc")
	u.RawQuery = query.Encode()
	return u.String()
}

func (client *Client) fetchJSON(ctx context.Context, requestURL string, target any) error {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return err
	}
	request.Header.Set("Accept", "application/json")
	request.Header.Set("User-Agent", "partijgedrag-rewrite/0.1")

	response, err := client.httpClient.Do(request)
	if err != nil {
		return err
	}
	defer response.Body.Close()

	if response.StatusCode != http.StatusOK {
		return fmt.Errorf("tweede kamer odata returned %d for %s", response.StatusCode, requestURL)
	}

	return json.NewDecoder(response.Body).Decode(target)
}

func formatODataDate(value time.Time) string {
	return value.UTC().Format("2006-01-02T15:04:05Z")
}
