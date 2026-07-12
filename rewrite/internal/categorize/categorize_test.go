package categorize

import (
	"slices"
	"testing"
)

func newTestMatcher(t *testing.T, categories []Category) *Matcher {
	t.Helper()
	matcher, err := NewMatcher(categories)
	if err != nil {
		t.Fatalf("NewMatcher: %v", err)
	}
	return matcher
}

func stringPtr(value string) *string {
	return &value
}

func TestMatchWholeWordsOnly(t *testing.T) {
	matcher := newTestMatcher(t, []Category{
		{CategoryKey: "internationaal", Keywords: []string{"eu"}},
	})

	if matches := matcher.Match(stringPtr("Motie over de euro"), nil); len(matches) != 0 {
		t.Errorf("expected no match inside 'euro', got %v", matches)
	}
	if matches := matcher.Match(stringPtr("Motie over de EU en verder"), nil); !slices.Contains(matches, "internationaal") {
		t.Errorf("expected match for standalone 'EU', got %v", matches)
	}
}

func TestMatchDiacritics(t *testing.T) {
	matcher := newTestMatcher(t, []Category{
		{CategoryKey: "zorg-en-gezondheid", Keywords: []string{"patiënt"}},
	})

	if matches := matcher.Match(stringPtr("Rechten van de patiënt"), nil); !slices.Contains(matches, "zorg-en-gezondheid") {
		t.Errorf("expected match for 'patiënt', got %v", matches)
	}
	if matches := matcher.Match(stringPtr("Rechten van patiënten"), nil); len(matches) != 0 {
		t.Errorf("expected no match inside 'patiënten', got %v", matches)
	}
}

func TestMatchMultipleCategories(t *testing.T) {
	matcher := newTestMatcher(t, []Category{
		{CategoryKey: "natuur-en-milieu", Keywords: []string{"klimaat"}},
		{CategoryKey: "klimaatverandering", Keywords: []string{"klimaat", "klimaatdoelen"}},
		{CategoryKey: "verkeer", Keywords: []string{"auto"}},
	})

	matches := matcher.Match(stringPtr("Motie over klimaat"), nil)
	want := []string{"natuur-en-milieu", "klimaatverandering"}
	if !slices.Equal(matches, want) {
		t.Errorf("expected %v, got %v", want, matches)
	}
}

func TestMatchEachCategoryAtMostOnce(t *testing.T) {
	matcher := newTestMatcher(t, []Category{
		{CategoryKey: "huisvesting", Keywords: []string{"wonen", "huur"}},
	})

	matches := matcher.Match(stringPtr("Betaalbaar wonen en huur"), nil)
	if len(matches) != 1 || matches[0] != "huisvesting" {
		t.Errorf("expected single match, got %v", matches)
	}
}

func TestMatchPhraseAndHyphenKeywords(t *testing.T) {
	matcher := newTestMatcher(t, []Category{
		{CategoryKey: "verkeer", Keywords: []string{"openbaar vervoer"}},
		{CategoryKey: "natuur-en-milieu", Keywords: []string{"zonne-energie"}},
	})

	if matches := matcher.Match(stringPtr("Investeren in openbaar vervoer"), nil); !slices.Contains(matches, "verkeer") {
		t.Errorf("expected phrase match, got %v", matches)
	}
	if matches := matcher.Match(nil, stringPtr("Subsidie voor zonne-energie")); !slices.Contains(matches, "natuur-en-milieu") {
		t.Errorf("expected hyphenated match, got %v", matches)
	}
}

func TestMatchCaseInsensitive(t *testing.T) {
	matcher := newTestMatcher(t, []Category{
		{CategoryKey: "oorlog", Keywords: []string{"gaza"}},
	})

	if matches := matcher.Match(stringPtr("Humanitaire situatie in Gaza"), nil); !slices.Contains(matches, "oorlog") {
		t.Errorf("expected case-insensitive match, got %v", matches)
	}
}

func TestMatchTitleAndSubjectCombined(t *testing.T) {
	matcher := newTestMatcher(t, []Category{
		{CategoryKey: "onderwijs-en-wetenschap", Keywords: []string{"school"}},
	})

	if matches := matcher.Match(nil, stringPtr("Veiligheid op school")); !slices.Contains(matches, "onderwijs-en-wetenschap") {
		t.Errorf("expected subject match, got %v", matches)
	}
	if matches := matcher.Match(nil, nil); matches != nil {
		t.Errorf("expected nil for empty text, got %v", matches)
	}
}
