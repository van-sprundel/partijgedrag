package openbesluitvorming

import (
	"regexp"
	"strings"
	"time"
)

const (
	OutcomeAdopted   = "aangenomen"
	OutcomeRejected  = "verworpen"
	OutcomeTied      = "staken"
	OutcomeWithdrawn = "ingetrokken"
	OutcomeHeld      = "aangehouden"
	OutcomeOther     = "overig"
)

func (payload MotionPayload) Outcome() string {
	votesFor, votesAgainst := 0, 0
	for _, vote := range payload.UniqueVotes() {
		switch vote.Option {
		case "voor":
			votesFor++
		case "tegen":
			votesAgainst++
		}
	}
	if votesFor == 0 && votesAgainst == 0 && payload.Tally != nil {
		votesFor, votesAgainst = payload.Tally.InFavour, payload.Tally.Against
	}

	switch {
	case votesFor > votesAgainst:
		return OutcomeAdopted
	case votesAgainst > votesFor:
		return OutcomeRejected
	case votesFor > 0:
		if payload.Result == OutcomeRejected {
			return OutcomeRejected
		}
		return OutcomeTied
	}
	return payload.Result
}

func (payload MotionPayload) UniqueVotes() []MotionVote {
	index := map[string]int{}
	var votes []MotionVote
	for _, vote := range payload.Votes {
		if vote.Option != "voor" && vote.Option != "tegen" {
			continue
		}
		key := vote.VoterKey()
		if key == "" {
			continue
		}
		if existing, ok := index[key]; ok {
			votes[existing] = vote
			continue
		}
		index[key] = len(votes)
		votes = append(votes, vote)
	}
	return votes
}

func (vote MotionVote) VoterKey() string {
	if vote.Voter != "" {
		return vote.Voter
	}
	if name := strings.TrimSpace(vote.VoterName); name != "" {
		return "name:" + name
	}
	return ""
}

func (payload MotionPayload) ProposedAt() *time.Time {
	for _, value := range []string{payload.LastDiscussedAt, payload.Date} {
		if value == "" {
			continue
		}
		for _, layout := range []string{time.RFC3339Nano, time.RFC3339, "2006-01-02"} {
			if parsed, err := time.Parse(layout, value); err == nil {
				return &parsed
			}
		}
	}
	return nil
}

var motionNumber = regexp.MustCompile(`^((?:[MA]\d+(?:[-.]\d+)?[a-z]?)|\d{1,3}\.)\s*-?\s+`)

func (payload MotionPayload) NumberAndTitle() (string, string) {
	name := strings.TrimSpace(payload.Name)
	match := motionNumber.FindStringSubmatch(name)
	if match == nil {
		return "", name
	}
	title := strings.TrimSpace(name[len(match[0]):])
	if title == "" {
		return "", name
	}
	return strings.TrimSuffix(match[1], "."), title
}

func (payload MotionPayload) Kind() string {
	lower := strings.ToLower(payload.MotionType)
	switch {
	case strings.Contains(lower, "amendement"):
		return "Amendement"
	case strings.Contains(lower, "motie"), strings.Contains(lower, "moties"):
		return "Motie"
	}
	return strings.TrimSpace(payload.MotionType)
}
