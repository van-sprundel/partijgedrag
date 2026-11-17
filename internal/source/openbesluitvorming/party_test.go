package openbesluitvorming

import "testing"

func TestPartyKeyFoldsSpellings(t *testing.T) {
	groups := []struct {
		municipality string
		names        []string
	}{
		{"Utrecht", []string{"Student & Starter", "Student en Starter"}},
		{"Utrecht", []string{"GroenLinks", "Groenlinks", "Fractie GroenLinks", "GL"}},
		{"Utrecht", []string{"ChristenUnie", "Christen Unie", "CU", "Fractie CU"}},
		{"Utrecht", []string{"Partij voor de Vrijheid", "PVV"}},
		{"Nijmegen", []string{"CDA", "Fractie CDA", "Raadsleden CDA", "CDA-fractie", "raadsleden cda"}},
		{"Westland", []string{"CDA Westland", "CDA"}},
		{"Nijmegen", []string{"GroenLinks-PvdA", "PvdA-GroenLinks", "GL-PvdA"}},
		{"Nijmegen", []string{"Fractie 50+", "50PLUS", "50 Plus"}},
	}
	for _, group := range groups {
		want := PartyKey(group.names[0], group.municipality)
		if want == "" {
			t.Fatalf("PartyKey(%q) is empty", group.names[0])
		}
		for _, name := range group.names[1:] {
			if got := PartyKey(name, group.municipality); got != want {
				t.Errorf("PartyKey(%q, %q) = %q, want %q (same as %q)", name, group.municipality, got, want, group.names[0])
			}
		}
	}
}

func TestPartyKeyKeepsDistinctFracties(t *testing.T) {
	distinct := [][2]string{
		{"Westland Verstandig", "Westlandse VVD"},
		{"Stadspartij Nijmegen", "Fractie 1Nijmegen"},
		{"SP", "SGP"},
		{"PvdA", "GroenLinks-PvdA"},
		{"CU-SGP Westland", "ChristenUnie"},
	}
	for _, pair := range distinct {
		if PartyKey(pair[0], "Westland") == PartyKey(pair[1], "Westland") {
			t.Errorf("%q and %q collapsed to %q", pair[0], pair[1], PartyKey(pair[0], "Westland"))
		}
	}
}

func TestPartyKeyRejectsNonFractieGroups(t *testing.T) {
	for _, name := range []string{"", "Raadsleden", "Fractie", "  ", "de fractie van"} {
		if got := PartyKey(name, "Utrecht"); got != "" {
			t.Errorf("PartyKey(%q) = %q, want empty", name, got)
		}
	}
}

func TestPartyKeyKeepsMunicipalityInLocalNames(t *testing.T) {
	if got := PartyKey("Westland Verstandig", "Westland"); got != "westlandverstandig" {
		t.Errorf("PartyKey(Westland Verstandig) = %q", got)
	}
	if got := PartyKey("Utrecht", "Utrecht"); got != "utrecht" {
		t.Errorf("PartyKey(Utrecht) = %q", got)
	}
}
