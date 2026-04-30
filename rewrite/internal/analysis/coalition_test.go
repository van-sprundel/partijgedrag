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
