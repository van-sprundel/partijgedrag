package analysis

import (
	"reflect"
	"testing"
)

func TestNormalizedPartyNames(t *testing.T) {
	got := normalizedPartyNames([]string{" vvd ", "D66", "", "cda"})
	want := []string{"VVD", "D66", "CDA"}
	if !reflect.DeepEqual(got, want) {
		t.Fatalf("normalizedPartyNames() = %#v, want %#v", got, want)
	}
}
