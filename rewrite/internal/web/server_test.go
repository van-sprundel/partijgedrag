package web

import "testing"

func TestNewParsesTemplates(t *testing.T) {
	server, err := New(nil)
	if err != nil {
		t.Fatalf("New() returned error: %v", err)
	}

	for _, name := range []string{"home", "motions", "motion"} {
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
