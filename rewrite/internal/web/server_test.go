package web

import "testing"

func TestNewParsesTemplates(t *testing.T) {
	server, err := New(nil)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	for _, name := range []string{"home", "motions", "motion", "party_likeness", "coalition_analysis", "coalition_motions"} {
		if server.templates[name] == nil {
			t.Fatalf("template %q was not parsed", name)
		}
	}
}

func TestMotionsURL(t *testing.T) {
	got := motionsURL("zorg wonen", true, 50, 100)
	want := "/motions?limit=50&offset=100&search=zorg+wonen&withVotes=true"
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
