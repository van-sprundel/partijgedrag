package openbesluitvorming

import (
	"os"
	"strings"
	"testing"
)

func readFixture(t *testing.T, name string) string {
	t.Helper()
	data, err := os.ReadFile("testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return string(data)
}

func TestExtractBulletPointsFixtures(t *testing.T) {
	tests := []struct {
		file  string
		count int
		first string
		last  string
	}{
		{
			file:  "utrecht-m287.md",
			count: 7,
			first: "Constaterende dat uit de Programmabegroting 2027 en de stukken bij de MIRT-verkenning niet eenduidig blijkt welke bedragen het college voor de Merwedelijn en de Rijnenburglijn heeft gereserveerd, al heeft uitgegeven en nog verwacht nodig te hebben.",
			last:  "Draagt het college op hierover in elk P&C-document per lijn te rapporteren over het gespaarde, bestede en nog benodigde bedrag.",
		},
		{
			file:  "nijmegen-m16.md",
			count: 18,
			first: "Constaterende dat het college met de Stadsbegroting 2027 structureel circa€ 8 miljoen aan extra parkeerinkomsten wil realiseren, grotendeels door forse verhogingen van bestaande parkeertarieven.",
			last:  "Verzoekt het college de raad uiterlijk voor mei 2027 te informeren over de bevindingen en, indien bijsturing nodig blijkt, concrete voorstellen voor mogelijke tariefaanpassingen voor te leggen, zodat de raad kan voor het zomerreces 2027 kan besluiten of het college alternatieven moet uitwerken in de Stadsbegroting 2028. Dit inclusief de financiële gevolgen, benodigde begrotingswijzigingen en alternatieve dekking, zodat de raad daadwerkelijk kan besluiten tot bijsturing.",
		},
		{
			file:  "nijmegen-behoud-sociaal-goede.md",
			count: 6,
			first: "Constaterende dat in de Voortgangsmonitor veel pro memorie(pm)-posten' zijn opgenomen. Zo bevat het overzicht Financieel effect/financiële gevolgen coronacrisis 25 pm-posten.",
			last:  "Roept het college op ons huidige sociaal beleid niet aan te tasten maar als ondergrens te zien en daarom het sociaal beleid uit te sluiten van mogelijke extra bezuinigingen voor deze raadsperiode.",
		},
		{
			file:  "westland-lasten-2027.md",
			count: 11,
			first: "Constaterende dat in de begroting 2027 wordt uitgegaan van een aanzienlijke verhoging van de afvalstoffenheffing;",
			last:  "Verzoekt het College aanvullende lastenverhogingen steeds afzonderlijk en volledig financieel te onderbouwen en uitdrukkelijk ter afweging aan de Gemeenteraad voor te leggen.",
		},
	}

	for _, test := range tests {
		t.Run(test.file, func(t *testing.T) {
			bullets := ExtractBulletPoints(readFixture(t, test.file))
			if len(bullets) != test.count {
				t.Fatalf("expected %d bullets, got %d:\n%s", test.count, len(bullets), strings.Join(bullets, "\n"))
			}
			if bullets[0] != test.first {
				t.Errorf("first bullet:\n got %q\nwant %q", bullets[0], test.first)
			}
			if bullets[len(bullets)-1] != test.last {
				t.Errorf("last bullet:\n got %q\nwant %q", bullets[len(bullets)-1], test.last)
			}
			for _, bullet := range bullets {
				lower := strings.ToLower(bullet)
				for _, junk := range []string{"pagina", "motie:", "ingediend door", "orde van de dag"} {
					if strings.Contains(lower, junk) {
						t.Errorf("bullet contains %q: %q", junk, bullet)
					}
				}
			}
		})
	}
}

func TestExtractBulletPointsInlineHeading(t *testing.T) {
	markdown := "De gemeenteraad van Houten,\n\nverzoekt het college om de\nspeeltuin te behouden;\n\nen gaat over tot de orde van de dag."
	bullets := ExtractBulletPoints(markdown)
	want := []string{"verzoekt het college om de speeltuin te behouden;"}
	if strings.Join(bullets, "|") != strings.Join(want, "|") {
		t.Fatalf("got %#v, want %#v", bullets, want)
	}
}

func TestExtractBulletPointsJoinsHyphenatedWords(t *testing.T) {
	markdown := "Draagt het college op:\n1. de herschik-\nkingen en de parkeer-\nen vervoersplannen te delen."
	bullets := ExtractBulletPoints(markdown)
	want := "Draagt het college op de herschikkingen en de parkeer- en vervoersplannen te delen."
	if len(bullets) != 1 || bullets[0] != want {
		t.Fatalf("got %#v, want %q", bullets, want)
	}
}

func TestExtractBulletPointsWithoutBody(t *testing.T) {
	markdown := "# Amendement\n\nBesluit:\n\nIn het besluit punt 2 te schrappen."
	if bullets := ExtractBulletPoints(markdown); bullets != nil {
		t.Fatalf("expected no bullets, got %#v", bullets)
	}
}

func TestLowerFirstKeepsAbbreviationsAndNames(t *testing.T) {
	names := map[string]bool{"Nijmegen": true}
	for input, want := range map[string]string{
		"Het college":      "het college",
		"SSH bouwt":        "SSH bouwt",
		"P+R Oost komt":    "P+R Oost komt",
		"€ 8 miljoen":      "€ 8 miljoen",
		"Nijmegen heeft":   "Nijmegen heeft",
		"Gezond en lokaal": "gezond en lokaal",
	} {
		if got := lowerFirst(input, names); got != want {
			t.Errorf("lowerFirst(%q) = %q, want %q", input, got, want)
		}
	}
}

func TestExtractBulletPointsUnmarkedItemsAndGluedHeadings(t *testing.T) {
	markdown := `De gemeenteraad van Nijmegen, in vergadering bijeen op 11 november 2026,

**Constaterende** **dat:**

In het coalitieakkoord één van de ambities was om een onderzoek te
starten naar korte ketens; en dat
Nijmegen is gestopt met het gebruik van gif in eigen groenbeheer;

**Overwegende dat:**
Gezond voedsel ook gifvrij zou moeten zijn. Geeft het college opdracht om:
alles te doen wat binnen zijn mogelijkheden ligt.

En gaat over tot de orde van de dag.`
	want := []string{
		"Constaterende dat in het coalitieakkoord één van de ambities was om een onderzoek te starten naar korte ketens;",
		"Constaterende dat Nijmegen is gestopt met het gebruik van gif in eigen groenbeheer;",
		"Overwegende dat gezond voedsel ook gifvrij zou moeten zijn.",
		"Geeft het college opdracht om alles te doen wat binnen zijn mogelijkheden ligt.",
	}
	got := ExtractBulletPoints(markdown)
	if len(got) != len(want) {
		t.Fatalf("got %d bullets, want %d:\n%s", len(got), len(want), strings.Join(got, "\n"))
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("bullet %d:\n got  %q\n want %q", i, got[i], want[i])
		}
	}
}

func TestExtractBulletPointsVraagtHeading(t *testing.T) {
	data, err := os.ReadFile("testdata/nijmegen-benchmark-dierenwelzijn.md")
	if err != nil {
		t.Fatal(err)
	}
	bullets := ExtractBulletPoints(string(data))
	asks := 0
	for _, bullet := range bullets {
		if strings.HasPrefix(bullet, "Vraagt het college ") {
			asks++
		}
		if strings.Contains(bullet, "Vraagt het college:") {
			t.Errorf("heading left inside an item: %q", bullet)
		}
	}
	if asks == 0 {
		t.Fatalf("no 'Vraagt het college' bullets in:\n%s", strings.Join(bullets, "\n"))
	}
}

func TestExtractBulletPointsStopsAtDamagedClosingAndSkipsFootnotes(t *testing.T) {
	markdown := `**Overwegende dat:**
- moslimdiscriminatie een specifieke vorm van discriminatie is;
1 [Monitor Discriminatiecijfers 2025](https://example.org/monitor)
**Draagt het college op:**
- de [monitor](https://example.org/m) jaarlijks te publiceren.
e} gaat over tot de orde van de dag ! PRO Stadspartij Nijmegen D66 SP FvD CDA Volt`
	want := []string{
		"Overwegende dat moslimdiscriminatie een specifieke vorm van discriminatie is;",
		"Draagt het college op de monitor jaarlijks te publiceren.",
	}
	got := ExtractBulletPoints(markdown)
	if strings.Join(got, "\n") != strings.Join(want, "\n") {
		t.Fatalf("got:\n%s\nwant:\n%s", strings.Join(got, "\n"), strings.Join(want, "\n"))
	}
}
