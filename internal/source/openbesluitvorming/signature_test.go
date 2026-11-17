package openbesluitvorming

import (
	"slices"
	"testing"
)

var (
	nijmegenGroups = []string{
		"Fractie 1Nijmegen", "Fractie 50+", "Fractie CDA", "Fractie D66", "Fractie Gewoon Nijmegen",
		"Fractie GewoonNijmegen.NU", "Fractie GroenLinks", "Fractie ODV", "Fractie Partij voor de Dieren",
		"Fractie PRO", "Fractie PvdA", "Fractie SP", "Fractie Stadspartij Nijmegen", "Fractie VVD",
		"Lijst De Vries-Kraaijeveld",
	}
	utrechtGroups = []string{
		"CDA", "ChristenUnie", "D66", "GroenLinks", "GroenLinks-PvdA", "PvdA", "Student en Starter",
		"Student & Starter", "VVD", "Volt", "groep Zwanenberg", "Stadsbelang Utrecht",
	}
)

func TestSignaturesCutJunkTail(t *testing.T) {
	cases := []struct {
		groups []string
		in     string
		want   string
	}{
		{
			nijmegenGroups,
			"Roept het college op bij Breng en de provincie de terugkeer van de nachtbussen te bepleiten; (}_f1 GroenLinks Stadspartij Nijmegen D66",
			"Roept het college op bij Breng en de provincie de terugkeer van de nachtbussen te bepleiten;",
		},
		{
			nijmegenGroups,
			"Roept het college op in het geval van een resterende bezuinigingsopgave, deze te dekken vanuit de saldireserve. SP Jeannette Geers GroenLinks Stadspartij Nijmegen D66 PvdA .1 /. PvdD VVD CDA Lijst De Vries-Kraaijeveld ODV",
			"Roept het college op in het geval van een resterende bezuinigingsopgave, deze te dekken vanuit de saldireserve.",
		},
		{
			nijmegenGroups,
			`Roept het college op in de veiligheidsbrief van 2025 een terugkoppeling te geven over de aanpak van gendergerelateerd geweld in Nijmegen. GroenLinks Marieke Smit A rSP ° ODV " https://www_stichtingopenmind_nl/femicide/ vvvr. 3 l@en. oVQ`,
			"Roept het college op in de veiligheidsbrief van 2025 een terugkoppeling te geven over de aanpak van gendergerelateerd geweld in Nijmegen.",
		},
		{
			nijmegenGroups,
			"Roept het college op om zich in te spannen om een beter beeld te krijgen van de woningen zonder definitief energielabel. @ PvdA Nijmegen",
			"Roept het college op om zich in te spannen om een beter beeld te krijgen van de woningen zonder definitief energielabel.",
		},
		{
			utrechtGroups,
			"Verzoekt het college om de gemeenteraad in 2021 te informeren over de uitkomsten van deze eerste verkenning. Zwinkels (CDA) Isik (PvdA) Te Hoonte (VVD) 1 Zie: https://www.utrecht.nl/wonen-en-leven/duurzame-stad/nieuws/utrecht-onderzoekt-mogelijkheden- aquathermie/",
			"Verzoekt het college om de gemeenteraad in 2021 te informeren over de uitkomsten van deze eerste verkenning.",
		},
	}
	for _, c := range cases {
		got := NewSignatures(c.groups).Strip([]string{"Constaterende dat er iets is;", c.in})
		if len(got) != 2 || got[1] != c.want {
			t.Errorf("Strip(%q)\n got %q\nwant %q", c.in, got, c.want)
		}
	}
}

func TestSignaturesDropAskThatIsOnlySignature(t *testing.T) {
	cases := []struct {
		groups []string
		in     string
	}{
		{nijmegenGroups, "Roept het college op W.H. van Hooft GroenLinks Stadspartij Nijmegen D66 PvdA PvdD VVD CDA Gewoon Nijmegen.nu ODV 'hts://app.powerbicom/view?r=eyJiiciMz8 s'@'¢ ##### 6 3 00pct"},
		{nijmegenGroups, "Roept het college op VVD CDA V00f :15 14eur o"},
		{utrechtGroups, "Draagt het college op saiah P. van Corler R. Streefland J. Koppelaar D66 Utrecht GroenLinks ChristenUnie Student en Starter"},
		{nijmegenGroups, "Roept het college op SP"},
		{utrechtGroups, "Verzoekt het college VVD"},
	}
	for _, c := range cases {
		bullets := []string{"Overwegende dat het belangrijk is;", "Verzoekt het college een plan te maken;", c.in}
		got := NewSignatures(c.groups).Strip(bullets)
		if !slices.Equal(got, bullets[:2]) {
			t.Errorf("Strip(%q) = %q, want the signature bullet dropped", c.in, got)
		}
	}
}

func TestSignaturesKeepOrdinaryText(t *testing.T) {
	bullets := [][]string{
		{"Verzoekt het college hierover in gesprek te gaan met de VVD"},
		{"Verzoekt het college samen met GroenLinks en PvdA een voorstel te maken"},
		{"Roept het college op bij Breng en de provincie de terugkeer van de nachtbussen te bepleiten in Nijmegen;"},
		{"Constaterende dat GroenLinks, PvdA, SP en D66 dit eerder vroegen;", "Verzoekt het college een plan te maken."},
		{"Roept het college op de kosten te dekken. Zie ook de motie van GroenLinks over dit onderwerp."},
	}
	for _, in := range bullets {
		if got := NewSignatures(nijmegenGroups).Strip(in); !slices.Equal(got, in) {
			t.Errorf("Strip(%q) = %q, want unchanged", in, got)
		}
	}
}

func TestNamePatternFoldsSpacing(t *testing.T) {
	signatures := NewSignatures([]string{"Fractie GewoonNijmegen.NU"})
	for _, text := range []string{"Groen Links", "groenLinks", "Christen Unie", "GewoonNijmegen.NU", "50PLUS"} {
		if names := signatures.matchNames(text); len(names) != 1 || names[0] != (span{0, len(text)}) {
			t.Errorf("matchNames(%q) = %v, want the whole text", text, names)
		}
	}
	if names := signatures.matchNames("Stadspartij"); len(names) != 0 {
		t.Errorf(`"SP" matched inside "Stadspartij": %v`, names)
	}
}
