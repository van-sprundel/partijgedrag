package odata

import (
	"context"
	"encoding/json"
	"etl/internal/config"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"
)

type Client struct {
	httpClient *http.Client
	baseURL    string
}

func NewClient(config config.APIConfig) *Client {
	return &Client{
		httpClient: &http.Client{
			Timeout: 60 * time.Second,
		},
		baseURL: config.ODataBaseURL,
	}
}

type QueryOptions struct {
	Filter    string
	Expand    string
	Select    string
	Top       int
	Skip      int
	SkipToken string
	OrderBy   string
	Count     bool
}

func (c *Client) BuildQuery(entitySet string, options QueryOptions) string {
	baseURL := fmt.Sprintf("%s/%s", c.baseURL, entitySet)
	params := url.Values{}

	if options.Filter != "" {
		params.Add("$filter", options.Filter)
	}
	if options.Expand != "" {
		params.Add("$expand", options.Expand)
	}
	if options.Select != "" {
		params.Add("$select", options.Select)
	}
	if options.Top > 0 {
		params.Add("$top", fmt.Sprintf("%d", options.Top))
	}
	if options.Skip > 0 {
		params.Add("$skip", fmt.Sprintf("%d", options.Skip))
	}
	if options.SkipToken != "" {
		params.Add("$skiptoken", options.SkipToken)
	}
	if options.OrderBy != "" {
		params.Add("$orderby", options.OrderBy)
	}
	if options.Count {
		params.Add("$count", "true")
	}

	if len(params) > 0 {
		return baseURL + "?" + params.Encode()
	}
	return baseURL
}

func (c *Client) ExecuteQuery(ctx context.Context, entitySet string, options QueryOptions) ([]byte, error) {
	queryURL := c.BuildQuery(entitySet, options)
	return c.MakeRequest(ctx, queryURL)
}

func (c *Client) GetMotiesWithVotes(ctx context.Context, skip int, top int) ([]byte, error) {
	options := QueryOptions{
		Filter: "verwijderd eq false and Soort eq 'Motie'",
		Expand: "Besluit($filter=Verwijderd eq false;$expand=Stemming($filter=Verwijderd eq false;$expand=Persoon,Fractie)),ZaakActor($filter=relatie eq 'Indiener'),Kamerstukdossier($filter=HoogsteVolgnummer gt 0;$select=Id,Nummer,HoogsteVolgnummer)",
		// dont set top to enable proper pagination with nextlink
		Skip: skip,
	}

	return c.ExecuteQuery(ctx, "Zaak", options)
}

// BuildDocumentURL constructs the URL for fetching XML documents from officielebekendmakingen.nl
func (c *Client) BuildDocumentURL(nummer string, volgnummer int) string {
	return fmt.Sprintf("https://zoek.officielebekendmakingen.nl/kst-%s-%d.xml", nummer, volgnummer)
}

// FetchDocument fetches an XML document from officielebekendmakingen.nl
func (c *Client) FetchDocument(ctx context.Context, nummer string, volgnummer int) ([]byte, error) {
	url := c.BuildDocumentURL(nummer, volgnummer)
	return c.MakeRequest(ctx, url)
}

func (c *Client) MakeRequest(ctx context.Context, requestURL string) ([]byte, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", requestURL, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("User-Agent", "odata-importer/1.0")

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

type ODataResponse struct {
	Context  string      `json:"@odata.context"`
	Count    int         `json:"@odata.count,omitempty"`
	NextLink string      `json:"@odata.nextLink,omitempty"`
	Value    interface{} `json:"value"`
}

func ParseODataResponse(data []byte) (*ODataResponse, error) {
	var response ODataResponse
	if err := json.Unmarshal(data, &response); err != nil {
		return nil, fmt.Errorf("parsing OData response: %w", err)
	}
	return &response, nil
}

func (r *ODataResponse) HasNextPage() bool {
	return r.NextLink != ""
}

func (r *ODataResponse) GetNextSkipToken() string {
	if r.NextLink == "" {
		return ""
	}

	u, err := url.Parse(r.NextLink)
	if err != nil {
		return ""
	}

	return u.Query().Get("$skiptoken")
}

func (r *ODataResponse) GetNextSkip() int {
	if r.NextLink == "" {
		return 0
	}

	u, err := url.Parse(r.NextLink)
	if err != nil {
		return 0
	}

	skipValue := u.Query().Get("$skip")
	if skipValue == "" {
		return 0
	}

	var skip int
	fmt.Sscanf(skipValue, "%d", &skip)
	return skip
}
