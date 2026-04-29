package tweedekamer

import (
	"net/url"
	"strings"
	"testing"
	"time"
)

func TestChangedMotionsURLUsesValidLeanQuery(t *testing.T) {
	client := NewClient("https://example.test/OData/v4/2.0")
	since := time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC)

	queryURL := client.changedMotionsURL(since, 100)
	parsed, err := url.Parse(queryURL)
	if err != nil {
		t.Fatal(err)
	}

	if parsed.Path != "/OData/v4/2.0/Zaak" {
		t.Fatalf("unexpected path %q", parsed.Path)
	}

	query := parsed.Query()
	if got := query.Get("$filter"); got != "Soort eq 'Motie' and ApiGewijzigdOp ge 2024-01-02T03:04:05Z" {
		t.Fatalf("unexpected filter %q", got)
	}
	if got := query.Get("$orderby"); got != "ApiGewijzigdOp asc,Id asc" {
		t.Fatalf("unexpected orderby %q", got)
	}
	if got := query.Get("$top"); got != "100" {
		t.Fatalf("unexpected top %q", got)
	}

	selectFields := strings.Split(query.Get("$select"), ",")
	disallowed := map[string]bool{
		"Datum": true,
	}
	required := map[string]bool{
		"Id":             false,
		"Soort":          false,
		"Onderwerp":      false,
		"GestartOp":      false,
		"ApiGewijzigdOp": false,
		"Verwijderd":     false,
	}

	for _, field := range selectFields {
		if disallowed[field] {
			t.Fatalf("query selects invalid field %q", field)
		}
		if _, ok := required[field]; ok {
			required[field] = true
		}
	}

	for field, seen := range required {
		if !seen {
			t.Fatalf("query is missing required field %q", field)
		}
	}
}

func TestMotionDecisionsURLUsesNavigationEndpoint(t *testing.T) {
	client := NewClient("https://example.test/OData/v4/2.0")

	queryURL := client.motionDecisionsURL("motion-id")
	parsed, err := url.Parse(queryURL)
	if err != nil {
		t.Fatal(err)
	}

	if parsed.Path != "/OData/v4/2.0/Zaak(motion-id)/Besluit" {
		t.Fatalf("unexpected path %q", parsed.Path)
	}

	selectFields := strings.Split(parsed.Query().Get("$select"), ",")
	required := map[string]bool{
		"Id":                            false,
		"StemmingsSoort":                false,
		"BesluitSoort":                  false,
		"BesluitTekst":                  false,
		"AgendapuntZaakBesluitVolgorde": false,
		"ApiGewijzigdOp":                false,
		"Verwijderd":                    false,
	}

	for _, field := range selectFields {
		if _, ok := required[field]; ok {
			required[field] = true
		}
	}

	for field, seen := range required {
		if !seen {
			t.Fatalf("query is missing required field %q", field)
		}
	}
}

func TestDecisionVotesURLUsesNavigationEndpoint(t *testing.T) {
	client := NewClient("https://example.test/OData/v4/2.0")

	queryURL := client.decisionVotesURL("decision-id")
	parsed, err := url.Parse(queryURL)
	if err != nil {
		t.Fatal(err)
	}

	if parsed.Path != "/OData/v4/2.0/Besluit(decision-id)/Stemming" {
		t.Fatalf("unexpected path %q", parsed.Path)
	}

	selectFields := strings.Split(parsed.Query().Get("$select"), ",")
	required := map[string]bool{
		"Id":             false,
		"Besluit_Id":     false,
		"Soort":          false,
		"ActorFractie":   false,
		"Fractie_Id":     false,
		"Vergissing":     false,
		"ApiGewijzigdOp": false,
		"Verwijderd":     false,
	}

	for _, field := range selectFields {
		if _, ok := required[field]; ok {
			required[field] = true
		}
	}

	for field, seen := range required {
		if !seen {
			t.Fatalf("query is missing required field %q", field)
		}
	}
}
