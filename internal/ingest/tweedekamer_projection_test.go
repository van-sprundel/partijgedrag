package ingest

import (
	"encoding/json"
	"os"
	"testing"

	"partijgedrag/internal/source/tweedekamer"
)

func TestProjectMotionFromFixture(t *testing.T) {
	record := readFixture[tweedekamer.MotionRecord](t, "testdata/tweedekamer_motion.json")

	raw := projectMotionRaw(record)
	if raw.Collection != zaakCollection {
		t.Fatalf("raw.Collection = %q, want %q", raw.Collection, zaakCollection)
	}
	if raw.SourceID != record.ID {
		t.Fatalf("raw.SourceID = %q, want %q", raw.SourceID, record.ID)
	}
	if raw.SourceDeleted {
		t.Fatal("raw.SourceDeleted = true, want false")
	}
	if raw.PayloadHash != hashBytes(record.Raw) {
		t.Fatalf("raw.PayloadHash = %q, want hash of fixture", raw.PayloadHash)
	}
	if got := raw.SourceUpdatedAt.Format("2006-01-02T15:04:05Z07:00"); got != "2026-04-29T14:23:18Z" {
		t.Fatalf("raw.SourceUpdatedAt = %q", got)
	}

	motion := projectMotion("nl-tweede-kamer", record)
	assertString(t, motion.MotionKey, "tweedekamer-odata-v2:2965764e-cc8e-45f1-8f55-00003c0ab123")
	assertString(t, motion.SourceKey, tweedeKamerSourceKey)
	assertString(t, motion.JurisdictionKey, "nl-tweede-kamer")
	assertStringPtr(t, motion.Number, "36410-14")
	assertStringPtr(t, motion.Title, "Motie over uitstel van harmonisatie")
	assertStringPtr(t, motion.Subject, "Motie van het lid Pieter Heerma over uitstel van de invoering van de harmonisatie")
	assertStringPtr(t, motion.Status, "Aangenomen")
	assertStringPtr(t, motion.Kind, "Motie")
	assertStringPtr(t, motion.ParliamentaryYear, "2025-2026")
	if got := motion.ProposedAt.Format("2006-01-02T15:04:05"); got != "2026-02-24T13:10:00" {
		t.Fatalf("motion.ProposedAt = %q", got)
	}
	if motion.SourceDeleted {
		t.Fatal("motion.SourceDeleted = true, want false")
	}
}

func TestProjectPartyFromFixture(t *testing.T) {
	record := readFixture[tweedekamer.PartyRecord](t, "testdata/tweedekamer_party.json")

	raw := projectPartyRaw(record)
	if raw.Collection != fractieCollection {
		t.Fatalf("raw.Collection = %q, want %q", raw.Collection, fractieCollection)
	}
	if raw.SourceID != record.ID {
		t.Fatalf("raw.SourceID = %q, want %q", raw.SourceID, record.ID)
	}
	if raw.PayloadHash != hashBytes(record.Raw) {
		t.Fatalf("raw.PayloadHash = %q, want hash of fixture", raw.PayloadHash)
	}

	party := projectParty("nl-tweede-kamer", record)
	assertString(t, party.PartyKey, "tweedekamer-odata-v2:party:8d46d23c-4f20-49be-b279-5439a2ef8d17")
	assertString(t, party.SourceKey, tweedeKamerSourceKey)
	assertString(t, party.JurisdictionKey, "nl-tweede-kamer")
	assertIntPtr(t, party.Number, 4)
	assertStringPtr(t, party.ShortName, "D66")
	assertStringPtr(t, party.Name, "Democraten 66")
	assertStringPtr(t, party.NameEN, "Democrats 66")
	assertIntPtr(t, party.Seats, 26)
	assertIntPtr(t, party.ElectoralVotes, 1670234)
	if got := party.ActiveFrom.Format("2006-01-02"); got != "2026-02-23" {
		t.Fatalf("party.ActiveFrom = %q", got)
	}
	if party.ActiveTo != nil {
		t.Fatalf("party.ActiveTo = %v, want nil", party.ActiveTo)
	}
	if party.SourceDeleted {
		t.Fatal("party.SourceDeleted = true, want false")
	}
}

func TestProjectDecisionAndVoteFromFixtures(t *testing.T) {
	decisionRecord := readFixture[tweedekamer.DecisionRecord](t, "testdata/tweedekamer_decision.json")
	voteRecord := readFixture[tweedekamer.VoteRecord](t, "testdata/tweedekamer_vote.json")
	motion := motionCandidate{
		MotionKey: "tweedekamer-odata-v2:2965764e-cc8e-45f1-8f55-00003c0ab123",
		SourceID:  "2965764e-cc8e-45f1-8f55-00003c0ab123",
	}

	decisionRaw := projectDecisionRaw(decisionRecord)
	if decisionRaw.Collection != besluitCollection {
		t.Fatalf("decisionRaw.Collection = %q, want %q", decisionRaw.Collection, besluitCollection)
	}
	if decisionRaw.PayloadHash != hashBytes(decisionRecord.Raw) {
		t.Fatal("decisionRaw.PayloadHash does not match fixture payload")
	}

	decision := projectDecision(motion, decisionRecord)
	assertString(t, decision.DecisionKey, "tweedekamer-odata-v2:decision:30fa3782-d2b3-40a7-b1a4-ebeaa9da9d06")
	assertString(t, decision.MotionKey, motion.MotionKey)
	assertStringPtr(t, decision.VotingType, "Hoofdelijk")
	assertStringPtr(t, decision.DecisionType, "Stemming")
	assertStringPtr(t, decision.DecisionText, "Aangenomen.")
	assertIntPtr(t, decision.DecisionOrder, 2)
	if decision.SourceDeleted {
		t.Fatal("decision.SourceDeleted = true, want false")
	}

	voteRaw := projectVoteRaw(voteRecord)
	if voteRaw.Collection != stemmingCollection {
		t.Fatalf("voteRaw.Collection = %q, want %q", voteRaw.Collection, stemmingCollection)
	}
	if voteRaw.PayloadHash != hashBytes(voteRecord.Raw) {
		t.Fatal("voteRaw.PayloadHash does not match fixture payload")
	}

	vote := projectVote(motion, decisionRecord, voteRecord)
	assertString(t, vote.VoteKey, "tweedekamer-odata-v2:vote:a4d4a2fd-c06f-428a-a303-22d28283b8b9")
	assertString(t, vote.MotionKey, motion.MotionKey)
	assertString(t, vote.DecisionKey, decision.DecisionKey)
	assertStringPtr(t, vote.VoteType, "Voor")
	assertStringPtr(t, vote.PartySourceID, "8d46d23c-4f20-49be-b279-5439a2ef8d17")
	assertStringPtr(t, vote.PartyName, "D66")
	assertStringPtr(t, vote.ActorName, "D66")
	assertIntPtr(t, vote.PartySize, 24)
	if !vote.Mistake {
		t.Fatal("vote.Mistake = false, want true")
	}
	if vote.SourceDeleted {
		t.Fatal("vote.SourceDeleted = true, want false")
	}
}

func readFixture[T any](t *testing.T, path string) T {
	t.Helper()

	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	var value T
	if err := json.Unmarshal(raw, &value); err != nil {
		t.Fatal(err)
	}
	return value
}

func assertString(t *testing.T, got string, want string) {
	t.Helper()
	if got != want {
		t.Fatalf("got %q, want %q", got, want)
	}
}

func assertStringPtr(t *testing.T, got *string, want string) {
	t.Helper()
	if got == nil {
		t.Fatalf("got nil, want %q", want)
	}
	if *got != want {
		t.Fatalf("got %q, want %q", *got, want)
	}
}

func assertIntPtr(t *testing.T, got *int, want int) {
	t.Helper()
	if got == nil {
		t.Fatalf("got nil, want %d", want)
	}
	if *got != want {
		t.Fatalf("got %d, want %d", *got, want)
	}
}
