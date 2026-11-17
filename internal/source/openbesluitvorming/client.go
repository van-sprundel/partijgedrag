package openbesluitvorming

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"
)

const userAgent = "partijgedrag-rewrite/0.1"

var ErrEntityNotFound = errors.New("openbesluitvorming entity not found")
var ErrEntityRemoved = errors.New("openbesluitvorming entity removed")

type Client struct {
	baseURL    string
	httpClient *http.Client

	mu        sync.Mutex
	waitUntil time.Time
}

func NewClient(baseURL string) *Client {
	return &Client{
		baseURL: strings.TrimRight(baseURL, "/"),
		httpClient: &http.Client{
			Timeout: 120 * time.Second,
		},
	}
}

type Source struct {
	Key              string `json:"key"`
	SourceRef        string `json:"sourceRef"`
	Label            string `json:"label"`
	Supplier         string `json:"supplier"`
	OrganizationType string `json:"organizationType"`
	Implemented      bool   `json:"implemented"`
	IsAggregate      bool   `json:"isAggregate"`
}

type SourceStatus struct {
	SourceKey string `json:"sourceKey"`
	Label     string `json:"label"`
	CBSID     string `json:"cbsId"`
	State     string `json:"state"`
}

type ExportRecord struct {
	Seq         int64           `json:"seq"`
	Op          string          `json:"op"`
	Time        *time.Time      `json:"time"`
	EntityID    string          `json:"entity_id"`
	EntityType  string          `json:"entity_type"`
	SourceKey   string          `json:"source_key"`
	Supplier    string          `json:"supplier"`
	ContentHash string          `json:"content_hash"`
	Reason      string          `json:"reason"`
	Payload     json.RawMessage `json:"payload"`
}

type ExportPage struct {
	Records       []ExportRecord
	NextCursor    string
	HasMore       bool
	ChangesCursor string
}

type MotionPayload struct {
	Name            string          `json:"name"`
	MotionType      string          `json:"motion_type"`
	Status          string          `json:"status"`
	Result          string          `json:"result"`
	Date            string          `json:"date"`
	LastDiscussedAt string          `json:"last_discussed_at"`
	Proposers       []string        `json:"proposers"`
	CoProposers     []string        `json:"co_proposers"`
	Parties         []string        `json:"parties"`
	Votes           []MotionVote    `json:"votes"`
	Tally           *Tally          `json:"tally"`
	VoteSummary     string          `json:"vote_summary"`
	Meeting         json.RawMessage `json:"meeting"`
	Attachment      json.RawMessage `json:"attachment"`
}

func (payload MotionPayload) HasAttachment() bool {
	var refs []json.RawMessage
	if err := json.Unmarshal(payload.Attachment, &refs); err != nil {
		return false
	}
	return len(refs) > 0
}

type MotionVote struct {
	Option    string `json:"option"`
	Voter     string `json:"voter"`
	VoterName string `json:"voter_name"`
	Group     string `json:"group"`
	GroupName string `json:"group_name"`
}

type Tally struct {
	InFavour int `json:"in_favour"`
	Against  int `json:"against"`
}

type Entity struct {
	EntityID     string `json:"entityId"`
	EntityType   string `json:"entityType"`
	Title        string `json:"title"`
	MarkdownText string `json:"markdownText"`
	PDFURL       string `json:"pdfUrl"`
	DownloadURL  string `json:"downloadUrl"`
	PDFEntityID  string `json:"pdfEntityId"`
}

func (client *Client) FetchSources(ctx context.Context) ([]Source, error) {
	var body struct {
		Sources []Source `json:"sources"`
	}
	if _, err := client.getJSON(ctx, client.baseURL+"/api/sources", &body); err != nil {
		return nil, err
	}
	return body.Sources, nil
}

func (client *Client) FetchStatus(ctx context.Context) ([]SourceStatus, error) {
	var body struct {
		Sources []SourceStatus `json:"sources"`
	}
	if _, err := client.getJSON(ctx, client.baseURL+"/api/status", &body); err != nil {
		return nil, err
	}
	return body.Sources, nil
}

func (client *Client) FetchSnapshot(ctx context.Context, source string, cursor string, limit int) (ExportPage, error) {
	return client.fetchExport(ctx, "snapshot", source, cursor, limit)
}

func (client *Client) FetchChanges(ctx context.Context, source string, cursor string, limit int) (ExportPage, error) {
	return client.fetchExport(ctx, "changes", source, cursor, limit)
}

func (client *Client) fetchExport(ctx context.Context, kind string, source string, cursor string, limit int) (ExportPage, error) {
	query := url.Values{}
	query.Set("source", source)
	if cursor != "" {
		query.Set("cursor", cursor)
	}
	if limit > 0 {
		query.Set("limit", strconv.Itoa(limit))
	}
	requestURL := fmt.Sprintf("%s/api/export/%s?%s", client.baseURL, kind, query.Encode())

	body, header, err := client.get(ctx, requestURL, "application/x-ndjson")
	if err != nil {
		return ExportPage{}, err
	}

	page := ExportPage{
		NextCursor:    header.Get("X-Next-Cursor"),
		HasMore:       header.Get("X-Has-More") == "true",
		ChangesCursor: header.Get("X-Changes-Cursor"),
	}
	for index, line := range strings.Split(string(body), "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		var record ExportRecord
		if err := json.Unmarshal([]byte(line), &record); err != nil {
			return ExportPage{}, fmt.Errorf("decode %s record %d of %s: %w", kind, index, source, err)
		}
		page.Records = append(page.Records, record)
	}
	return page, nil
}

func (client *Client) FetchEntity(ctx context.Context, entityID string) (Entity, error) {
	requestURL := fmt.Sprintf("%s/api/entities/%s", client.baseURL, url.PathEscape(entityID))
	var entity Entity
	status, err := client.getJSON(ctx, requestURL, &entity)
	switch {
	case status == http.StatusNotFound:
		return Entity{}, ErrEntityNotFound
	case status == http.StatusGone:
		return Entity{}, ErrEntityRemoved
	case err != nil:
		return Entity{}, err
	}
	return entity, nil
}

func (client *Client) getJSON(ctx context.Context, requestURL string, target any) (int, error) {
	body, _, err := client.get(ctx, requestURL, "application/json")
	if err != nil {
		var statusErr *StatusError
		if errors.As(err, &statusErr) {
			return statusErr.Status, err
		}
		return 0, err
	}
	if err := json.Unmarshal(body, target); err != nil {
		return http.StatusOK, fmt.Errorf("decode %s: %w", requestURL, err)
	}
	return http.StatusOK, nil
}

type StatusError struct {
	Status int
	Code   string
	URL    string
	Body   string
}

func (err *StatusError) Error() string {
	return fmt.Sprintf("openbesluitvorming returned %d (%s) for %s: %s", err.Status, err.Code, err.URL, err.Body)
}

const maxAttempts = 6

func (client *Client) get(ctx context.Context, requestURL string, accept string) ([]byte, http.Header, error) {
	var lastErr error
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		if err := client.pace(ctx); err != nil {
			return nil, nil, err
		}

		body, header, retryAfter, err := client.getOnce(ctx, requestURL, accept)
		if err == nil {
			return body, header, nil
		}
		var statusErr *StatusError
		if errors.As(err, &statusErr) {
			return nil, nil, err
		}
		lastErr = err

		if retryAfter <= 0 {
			retryAfter = time.Duration(attempt) * 2 * time.Second
		}
		client.delay(retryAfter)
	}
	return nil, nil, lastErr
}

func (client *Client) getOnce(ctx context.Context, requestURL string, accept string) ([]byte, http.Header, time.Duration, error) {
	request, err := http.NewRequestWithContext(ctx, http.MethodGet, requestURL, nil)
	if err != nil {
		return nil, nil, 0, err
	}
	request.Header.Set("Accept", accept)
	request.Header.Set("User-Agent", userAgent)

	response, err := client.httpClient.Do(request)
	if err != nil {
		return nil, nil, 0, err
	}
	defer response.Body.Close()

	client.observeRateLimit(response.Header)

	body, err := io.ReadAll(response.Body)
	if err != nil {
		return nil, nil, 0, err
	}

	switch {
	case response.StatusCode == http.StatusOK:
		return body, response.Header, 0, nil
	case response.StatusCode == http.StatusTooManyRequests,
		response.StatusCode == http.StatusServiceUnavailable,
		response.StatusCode >= 500:
		retryAfter := parseSeconds(response.Header.Get("Retry-After"))
		return nil, nil, retryAfter, fmt.Errorf("openbesluitvorming returned %d for %s", response.StatusCode, requestURL)
	default:
		var decoded struct {
			Code string `json:"code"`
		}
		_ = json.Unmarshal(body, &decoded)
		snippet := strings.TrimSpace(string(body))
		if len(snippet) > 500 {
			snippet = snippet[:500]
		}
		return nil, nil, 0, &StatusError{Status: response.StatusCode, Code: decoded.Code, URL: requestURL, Body: snippet}
	}
}

func (client *Client) observeRateLimit(header http.Header) {
	remaining, err := strconv.Atoi(header.Get("RateLimit-Remaining"))
	if err != nil || remaining > 2 {
		return
	}
	reset := parseSeconds(header.Get("RateLimit-Reset"))
	wait := 3 * time.Second
	if reset > 0 && reset < wait {
		wait = reset
	}
	client.delay(wait)
}

func (client *Client) delay(wait time.Duration) {
	client.mu.Lock()
	defer client.mu.Unlock()
	until := time.Now().Add(wait)
	if until.After(client.waitUntil) {
		client.waitUntil = until
	}
}

func (client *Client) pace(ctx context.Context) error {
	client.mu.Lock()
	wait := time.Until(client.waitUntil)
	client.mu.Unlock()
	if wait <= 0 {
		return nil
	}
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-time.After(wait):
		return nil
	}
}

func parseSeconds(value string) time.Duration {
	seconds, err := strconv.Atoi(strings.TrimSpace(value))
	if err != nil || seconds < 0 {
		return 0
	}
	return time.Duration(seconds) * time.Second
}
