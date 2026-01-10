package models

import "time"

// CoalitionAlignment represents how often two parties vote the same way
// Note: This is a query result struct, not a persisted table
type CoalitionAlignment struct {
	Fractie1ID   string    `json:"fractie1_id"`
	Fractie2ID   string    `json:"fractie2_id"`
	Period       string    `json:"period"`
	AlignmentPct float64   `json:"alignment_pct"`
	SameVotes    int       `json:"same_votes"`
	TotalVotes   int       `json:"total_votes"`
	UpdatedAt    time.Time `json:"updated_at"`
	Fractie1Name string    `json:"fractie1_name"`
	Fractie2Name string    `json:"fractie2_name"`
}

// MPDeviation tracks how often an MP votes against their party's majority
// Note: This is a query result struct, not a persisted table
type MPDeviation struct {
	PersoonID      string    `json:"persoon_id"`
	FractieID      string    `json:"fractie_id"`
	Period         string    `json:"period"`
	DeviationPct   float64   `json:"deviation_pct"`
	DeviationCount int       `json:"deviation_count"`
	TotalVotes     int       `json:"total_votes"`
	UpdatedAt      time.Time `json:"updated_at"`
	PersoonNaam    string    `json:"persoon_naam"`
	FractieNaam    string    `json:"fractie_naam"`
}

// TopicTrend tracks motion counts by category over time
// Note: This is a query result struct, not a persisted table
type TopicTrend struct {
	CategoryID    string    `json:"category_id"`
	Period        string    `json:"period"`
	MotionCount   int       `json:"motion_count"`
	AcceptedCount int       `json:"accepted_count"`
	RejectedCount int       `json:"rejected_count"`
	UpdatedAt     time.Time `json:"updated_at"`
	CategoryName  string    `json:"category_name"`
}

// PartyTopicVoting tracks how a party votes on specific topics
// Note: This is a query result struct, not a persisted table
type PartyTopicVoting struct {
	FractieID    string    `json:"fractie_id"`
	CategoryID   string    `json:"category_id"`
	Period       string    `json:"period"`
	VotesFor     int       `json:"votes_for"`
	VotesAgainst int       `json:"votes_against"`
	Abstentions  int       `json:"abstentions"`
	TotalVotes   int       `json:"total_votes"`
	ForPct       float64   `json:"for_pct"`
	UpdatedAt    time.Time `json:"updated_at"`
	FractieNaam  string    `json:"fractie_naam"`
	CategoryName string    `json:"category_name"`
}

// AnalysisResult is a generic wrapper for analysis query results
type AnalysisResult struct {
	Type        string      `json:"type"`
	Period      string      `json:"period"`
	GeneratedAt time.Time   `json:"generated_at"`
	Data        interface{} `json:"data"`
}
