package web

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"partijgedrag/internal/analysis"
)

func TestNewParsesTemplates(t *testing.T) {
	server, err := New(nil, false)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	for _, name := range []string{"home", "motions", "motion", "party_likeness", "party_comparison", "party_focus", "coalition_analysis", "coalition_motions", "voting_compass", "compass_results", "data_quality"} {
		if server.templates[name] == nil {
			t.Fatalf("template %q was not parsed", name)
		}
	}
}

func TestPartyComparisonURL(t *testing.T) {
	got := partyComparisonURL("rutte-iv", 10, "vvd-id", "nsc-id", "disagree", "", 25, 0)
	want := "/party-likeness/compare?party1=vvd-id&party2=nsc-id&period=rutte-iv"
	if got != want {
		t.Fatalf("partyComparisonURL() = %q, want %q", got, want)
	}

	got = partyComparisonURL("rutte-iv", 25, "vvd-id", "nsc-id", "agree", "zorg-en-gezondheid", 50, 50)
	want = "/party-likeness/compare?category=zorg-en-gezondheid&limit=50&minCommon=25&offset=50&party1=vvd-id&party2=nsc-id&period=rutte-iv&relation=agree"
	if got != want {
		t.Fatalf("partyComparisonURL() = %q, want %q", got, want)
	}
}

func TestCompareURLSkipsCellsWithoutData(t *testing.T) {
	matrix := likenessMatrix([]analysis.PartyLikeness{{
		Party1SourceID: "vvd-id",
		Party1Name:     "VVD",
		Party2SourceID: "nsc-id",
		Party2Name:     "NSC",
		CommonMotions:  120,
		SameVotes:      94,
		Similarity:     78.33,
	}})

	if got := compareURL(matrix, "vvd-id", "vvd-id", "rutte-iv", 10); got != "" {
		t.Fatalf("diagonal cell should not link, got %q", got)
	}
	if got := compareURL(matrix, "vvd-id", "unknown-id", "rutte-iv", 10); got != "" {
		t.Fatalf("cell below minCommon should not link, got %q", got)
	}
	// The matrix is mirrored, so a pair links from either side.
	for _, pair := range [][2]string{{"vvd-id", "nsc-id"}, {"nsc-id", "vvd-id"}} {
		if got := compareURL(matrix, pair[0], pair[1], "rutte-iv", 10); got == "" {
			t.Fatalf("cell %v should link to the comparison page", pair)
		}
	}
}

func TestComparisonRelationViewsFollowCategoryFilter(t *testing.T) {
	comparison := analysis.PartyComparison{
		CommonMotions:  100,
		SameVotes:      78,
		DifferentVotes: 22,
		Categories: []analysis.ComparisonCategory{{
			CategoryKey:    "migratie",
			Name:           "Migratie",
			CommonMotions:  30,
			SameVotes:      18,
			DifferentVotes: 12,
		}},
	}
	link := func(relation string, category string, offset int) string { return relation }

	views := comparisonRelationViews(comparison, "", "disagree", link)
	if views[0].Count != 22 || views[1].Count != 78 || views[2].Count != 100 {
		t.Fatalf("unfiltered counts = %d/%d/%d, want 22/78/100", views[0].Count, views[1].Count, views[2].Count)
	}
	if !views[0].Selected || views[1].Selected {
		t.Fatal("expected only the disagree toggle to be selected")
	}

	views = comparisonRelationViews(comparison, "migratie", "agree", link)
	if views[0].Count != 12 || views[1].Count != 18 || views[2].Count != 30 {
		t.Fatalf("filtered counts = %d/%d/%d, want 12/18/30", views[0].Count, views[1].Count, views[2].Count)
	}
	if !views[1].Selected {
		t.Fatal("expected the agree toggle to be selected")
	}
}

func TestComparisonRelationViewsHandleUnknownCategory(t *testing.T) {
	views := comparisonRelationViews(analysis.PartyComparison{CommonMotions: 100, SameVotes: 78, DifferentVotes: 22}, "geen-categorie", "disagree", func(string, string, int) string { return "" })
	for _, view := range views {
		if view.Count != 0 {
			t.Fatalf("unknown category should count nothing, got %s = %d", view.Key, view.Count)
		}
	}
}

func TestMotionsURL(t *testing.T) {
	got := motionsURL("zorg wonen", true, "zorg-en-gezondheid", 50, 100)
	want := "/motions?category=zorg-en-gezondheid&limit=50&offset=100&search=zorg+wonen&withVotes=true"
	if got != want {
		t.Fatalf("motionsURL() = %q, want %q", got, want)
	}
}

func TestCoalitionMotionsURL(t *testing.T) {
	got := coalitionMotionsURL("rutte-iv", "party-id", "ChristenUnie", "against", 25, 50, "10")
	want := "/coalition-analysis/motions?limit=25&minCommon=10&offset=50&partyName=ChristenUnie&partySourceId=party-id&period=rutte-iv&relation=against"
	if got != want {
		t.Fatalf("coalitionMotionsURL() = %q, want %q", got, want)
	}
}

func TestStaticCacheControl(t *testing.T) {
	server, err := New(nil, false)
	if err != nil {
		t.Fatalf("New() error: %v", err)
	}

	mux := http.NewServeMux()
	server.Register(mux)

	req := httptest.NewRequest("GET", "/static/styles.css", nil)
	rec := httptest.NewRecorder()
	mux.ServeHTTP(rec, req)

	got := rec.Header().Get("Cache-Control")
	want := "public, max-age=31536000, immutable"
	if got != want {
		t.Fatalf("expected Cache-Control %q for static files, got %q", want, got)
	}
}
