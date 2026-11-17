package openbesluitvorming

import "testing"

func TestOutcomePrefersVotesOverResult(t *testing.T) {
	payload := MotionPayload{
		Status: "Niet aangenomen",
		Result: "aangenomen",
		Votes: []MotionVote{
			{Option: "voor", Voter: "person:1"},
			{Option: "tegen", Voter: "person:2"},
			{Option: "tegen", Voter: "person:3"},
		},
	}
	if got := payload.Outcome(); got != OutcomeRejected {
		t.Fatalf("Outcome() = %q, want %q", got, OutcomeRejected)
	}
}

func TestOutcomeFallbacks(t *testing.T) {
	tests := []struct {
		name    string
		payload MotionPayload
		want    string
	}{
		{"tally without votes", MotionPayload{Result: "overig", Tally: &Tally{InFavour: 20, Against: 11}}, OutcomeAdopted},
		{"tie", MotionPayload{Votes: []MotionVote{{Option: "voor", Voter: "a"}, {Option: "tegen", Voter: "b"}}}, OutcomeTied},
		{"tie rejected by a full council", MotionPayload{Result: "verworpen", Votes: []MotionVote{{Option: "voor", Voter: "a"}, {Option: "tegen", Voter: "b"}}}, OutcomeRejected},
		{"withdrawn, no vote", MotionPayload{Result: "ingetrokken"}, OutcomeWithdrawn},
		{"nothing known", MotionPayload{}, ""},
	}
	for _, test := range tests {
		if got := test.payload.Outcome(); got != test.want {
			t.Errorf("%s: Outcome() = %q, want %q", test.name, got, test.want)
		}
	}
}

func TestUniqueVotesKeepsLastVotePerMember(t *testing.T) {
	payload := MotionPayload{Votes: []MotionVote{
		{Option: "voor", Voter: "person:1"},
		{Option: "tegen", VoterName: "Jansen, P."},
		{Option: "tegen", Voter: "person:1"},
		{Option: "", Voter: "person:3"},
	}}
	votes := payload.UniqueVotes()
	if len(votes) != 2 {
		t.Fatalf("expected 2 votes, got %#v", votes)
	}
	if votes[0].Option != "tegen" || votes[1].VoterKey() != "name:Jansen, P." {
		t.Fatalf("unexpected votes %#v", votes)
	}
}

func TestNumberAndTitle(t *testing.T) {
	tests := map[string][2]string{
		"M287 Transparant sparen voor Merwedelijn en Rijnenburglijn": {"M287", "Transparant sparen voor Merwedelijn en Rijnenburglijn"},
		"A53 Tegenbegroting": {"A53", "Tegenbegroting"},
		"M24-120 - Kansen verzilveren in Den Haag en Haarlem":             {"M24-120", "Kansen verzilveren in Den Haag en Haarlem"},
		"01. Motie WV-FvD inzake Beperking stijging gemeentelijke lasten": {"01", "Motie WV-FvD inzake Beperking stijging gemeentelijke lasten"},
		"Motie Behoud het sociaal goede":                                  {"", "Motie Behoud het sociaal goede"},
		"Mobiliteitsplan":                                                 {"", "Mobiliteitsplan"},
	}
	for name, want := range tests {
		number, title := MotionPayload{Name: name}.NumberAndTitle()
		if number != want[0] || title != want[1] {
			t.Errorf("NumberAndTitle(%q) = (%q, %q), want (%q, %q)", name, number, title, want[0], want[1])
		}
	}
}

func TestKind(t *testing.T) {
	for motionType, want := range map[string]string{
		"Moties":       "Motie",
		"Motie Vreemd": "Motie",
		"Amendementen": "Amendement",
		"Initiatief":   "Initiatief",
	} {
		if got := (MotionPayload{MotionType: motionType}).Kind(); got != want {
			t.Errorf("Kind(%q) = %q, want %q", motionType, got, want)
		}
	}
}
