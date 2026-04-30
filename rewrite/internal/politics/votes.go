package politics

type Position string

const (
	PositionFor     Position = "FOR"
	PositionAgainst Position = "AGAINST"
	PositionNeutral Position = "NEUTRAL"
)

func PartyPosition(votesFor int, votesAgainst int) Position {
	if votesFor > votesAgainst {
		return PositionFor
	}
	if votesAgainst > votesFor {
		return PositionAgainst
	}
	return PositionNeutral
}
