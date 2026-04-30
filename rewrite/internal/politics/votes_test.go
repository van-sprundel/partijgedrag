package politics

import "testing"

func TestPartyPosition(t *testing.T) {
	tests := []struct {
		name         string
		votesFor     int
		votesAgainst int
		want         Position
	}{
		{
			name:         "for",
			votesFor:     10,
			votesAgainst: 3,
			want:         PositionFor,
		},
		{
			name:         "against",
			votesFor:     2,
			votesAgainst: 7,
			want:         PositionAgainst,
		},
		{
			name:         "tie",
			votesFor:     4,
			votesAgainst: 4,
			want:         PositionNeutral,
		},
		{
			name:         "no votes",
			votesFor:     0,
			votesAgainst: 0,
			want:         PositionNeutral,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got := PartyPosition(test.votesFor, test.votesAgainst)
			if got != test.want {
				t.Fatalf("PartyPosition(%d, %d) = %q, want %q", test.votesFor, test.votesAgainst, got, test.want)
			}
		})
	}
}
