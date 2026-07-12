package officielebekendmakingen

import (
	"os"
	"strings"
	"testing"
)

func TestExtractBulletPoints(t *testing.T) {
	xmlData, err := os.ReadFile("testdata/kst-36045-299.xml")
	if err != nil {
		t.Fatal(err)
	}

	result, err := ExtractBulletPoints(xmlData)
	if err != nil {
		t.Fatal(err)
	}
	if result == nil {
		t.Fatal("expected a motion document, got nil result")
	}

	if result.Title != "MOTIE VAN DE LEDEN DASSEN EN PIRI" {
		t.Errorf("unexpected title: %q", result.Title)
	}

	expected := []string{
		"constaterende dat het Matra-programma al decennialang bijdraagt aan de versterking van de democratische rechtsstaat, goed bestuur en maatschappelijke organisaties in (potentiële) kandidaat-lidstaten van de Europese Unie;",
		"overwegende dat het versterken van democratie en rechtsstaat in de Europese buurlanden van strategisch belang is voor de veiligheid, stabiliteit en toekomstige uitbreiding van de Europese Unie;",
		"overwegende dat er nog bezuinigingen ingepland staan op het Matra-fonds;",
		"verzoekt de regering niet te bezuinigen op het Matra-fonds",
	}
	if len(result.BulletPoints) != len(expected) {
		t.Fatalf("expected %d bullet points, got %d: %#v", len(expected), len(result.BulletPoints), result.BulletPoints)
	}
	for i, want := range expected {
		if result.BulletPoints[i] != want {
			t.Errorf("bullet %d:\n got %q\nwant %q", i, result.BulletPoints[i], want)
		}
	}
}

func TestExtractBulletPointsNonMotion(t *testing.T) {
	xmlData := []byte(`<?xml version="1.0" encoding="utf-8"?>
<officiele-publicatie>
  <kamerstuk>
    <stuk>
      <titel>BRIEF VAN DE MINISTER</titel>
      <algemeen><vrije-tekst><tekst><al>verzoekt van alles</al></tekst></vrije-tekst></algemeen>
    </stuk>
  </kamerstuk>
</officiele-publicatie>`)

	result, err := ExtractBulletPoints(xmlData)
	if err != nil {
		t.Fatal(err)
	}
	if result != nil {
		t.Fatalf("expected nil result for non-motion document, got %#v", result)
	}
}

func TestExtractBulletPointsInvalidXML(t *testing.T) {
	if _, err := ExtractBulletPoints([]byte("<html>not found</html>")); err == nil {
		// an HTML error page has no kamerstuk title, so it is treated as a non-motion
		return
	}
}

func TestDocumentURL(t *testing.T) {
	if got := DocumentURL("36045", "", 299); got != "https://zoek.officielebekendmakingen.nl/kst-36045-299.xml" {
		t.Errorf("unexpected url: %s", got)
	}
	if got := DocumentURL("21501", "02", 3196); got != "https://zoek.officielebekendmakingen.nl/kst-21501-02-3196.xml" {
		t.Errorf("unexpected url: %s", got)
	}
	if !strings.HasPrefix(DocumentURL("36045", "", 1), "https://zoek.officielebekendmakingen.nl/") {
		t.Error("unexpected host")
	}
}
