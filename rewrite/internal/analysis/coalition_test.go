package analysis

import (
	"reflect"
	"testing"
)

func TestNormalizedPartyNames(t *testing.T) {
	got := normalizedPartyNames([]string{" vvd ", "D66", "", "cda", "CU", "NSC", "CU"})
	want := []string{"VVD", "D66", "CDA", "CU", "CHRISTENUNIE", "NSC", "NIEUW SOCIAAL CONTRACT"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("normalizedPartyNames() = %#v, want %#v", got, want)
	}
}

func TestNormalizeCoalitionRelation(t *testing.T) {
	for input, want := range map[string]string{
		"":        "all",
		"all":     "all",
		" WITH ":  "with",
		"against": "against",
		"AGAINST": "against",
	} {
		got, ok := NormalizeCoalitionRelation(input)
		if !ok {
			t.Fatalf("NormalizeCoalitionRelation(%q) returned !ok", input)
		}
		if got != want {
			t.Fatalf("NormalizeCoalitionRelation(%q) = %q, want %q", input, got, want)
		}
	}

	if got, ok := NormalizeCoalitionRelation("split"); ok || got != "" {
		t.Fatalf("NormalizeCoalitionRelation invalid = %q, %v; want empty, false", got, ok)
	}
}
