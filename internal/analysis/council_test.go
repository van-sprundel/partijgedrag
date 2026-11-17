package analysis

import (
	"testing"
	"time"
)

func TestCouncilSince(t *testing.T) {
	day := func(value string) time.Time {
		parsed, _ := time.Parse("2006-01-02", value)
		return parsed
	}
	nijmegen := []councilParty{
		{day("2019-04-03"), 7}, {day("2020-09-23"), 6}, {day("2019-04-03"), 3}, {day("2019-04-03"), 2},
		{day("2019-04-03"), 2}, {day("2026-06-10"), 1}, {day("2026-06-10"), 1}, {day("2019-04-03"), 1},
		{day("2026-06-10"), 1}, {day("2026-07-08"), 15},
	}
	if got := councilSince(nijmegen); got == nil || !got.Equal(day("2026-07-08")) {
		t.Errorf("Nijmegen since = %v, want 2026-07-08", got)
	}

	withSplit := append(nijmegen, councilParty{day("2027-03-01"), 1})
	if got := councilSince(withSplit); got == nil || !got.Equal(day("2026-07-08")) {
		t.Errorf("since with a late split = %v, want 2026-07-08", got)
	}

	if got := councilSince(nil); got != nil {
		t.Errorf("since without parties = %v, want nil", got)
	}
}
