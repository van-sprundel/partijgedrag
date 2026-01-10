package analysis

import (
	"fmt"
	"sort"
	"strings"

	"etl/internal/models"
)

// FormatCoalitionAlignments formats alignment data for terminal output
func FormatCoalitionAlignments(alignments []models.CoalitionAlignment) string {
	if len(alignments) == 0 {
		return "No alignment data available."
	}

	var sb strings.Builder
	sb.WriteString("\n=== Coalition Alignment ===\n")
	sb.WriteString(fmt.Sprintf("%-10s %-10s %8s %8s %8s\n", "Party 1", "Party 2", "Align%", "Same", "Total"))
	sb.WriteString(strings.Repeat("-", 50) + "\n")

	for _, a := range alignments {
		sb.WriteString(fmt.Sprintf("%-10s %-10s %7.1f%% %8d %8d\n",
			truncate(a.Fractie1Name, 10),
			truncate(a.Fractie2Name, 10),
			a.AlignmentPct,
			a.SameVotes,
			a.TotalVotes,
		))
	}

	return sb.String()
}

// FormatCoalitionMatrix formats alignment as a heatmap-style matrix
func FormatCoalitionMatrix(parties []string, matrix map[string]map[string]float64) string {
	if len(parties) == 0 {
		return "No matrix data available."
	}

	var sb strings.Builder
	sb.WriteString("\n=== Coalition Matrix (% agreement) ===\n\n")

	// Header row
	sb.WriteString(fmt.Sprintf("%8s", ""))
	for _, p := range parties {
		sb.WriteString(fmt.Sprintf(" %6s", truncate(p, 6)))
	}
	sb.WriteString("\n")

	// Data rows
	for _, p1 := range parties {
		sb.WriteString(fmt.Sprintf("%8s", truncate(p1, 8)))
		for _, p2 := range parties {
			val := matrix[p1][p2]
			if p1 == p2 {
				sb.WriteString(fmt.Sprintf(" %6s", "---"))
			} else {
				sb.WriteString(fmt.Sprintf(" %5.0f%%", val))
			}
		}
		sb.WriteString("\n")
	}

	return sb.String()
}

// FormatMPDeviations formats MP deviation data for terminal output
func FormatMPDeviations(deviations []models.MPDeviation) string {
	if len(deviations) == 0 {
		return "No deviation data available."
	}

	var sb strings.Builder
	sb.WriteString("\n=== MP Party Line Deviations ===\n")
	sb.WriteString("(MPs who vote against their party most often)\n\n")
	sb.WriteString(fmt.Sprintf("%-25s %-10s %8s %8s %8s\n", "Name", "Party", "Dev%", "Devs", "Total"))
	sb.WriteString(strings.Repeat("-", 65) + "\n")

	for _, d := range deviations {
		sb.WriteString(fmt.Sprintf("%-25s %-10s %7.1f%% %8d %8d\n",
			truncate(d.PersoonNaam, 25),
			truncate(d.FractieNaam, 10),
			d.DeviationPct,
			d.DeviationCount,
			d.TotalVotes,
		))
	}

	return sb.String()
}

// FormatTopicTrends formats topic trend data for terminal output
func FormatTopicTrends(trends []models.TopicTrend) string {
	if len(trends) == 0 {
		return "No topic trend data available."
	}

	var sb strings.Builder
	sb.WriteString("\n=== Topic Trends ===\n")
	sb.WriteString(fmt.Sprintf("%-25s %8s %10s %10s\n", "Category", "Motions", "Accepted", "Rejected"))
	sb.WriteString(strings.Repeat("-", 55) + "\n")

	for _, t := range trends {
		sb.WriteString(fmt.Sprintf("%-25s %8d %10d %10d\n",
			truncate(t.CategoryName, 25),
			t.MotionCount,
			t.AcceptedCount,
			t.RejectedCount,
		))
	}

	return sb.String()
}

// FormatPartyTopicVoting formats party voting patterns by topic
func FormatPartyTopicVoting(voting []models.PartyTopicVoting) string {
	if len(voting) == 0 {
		return "No party topic voting data available."
	}

	// Group by party
	byParty := make(map[string][]models.PartyTopicVoting)
	for _, v := range voting {
		byParty[v.FractieNaam] = append(byParty[v.FractieNaam], v)
	}

	var sb strings.Builder
	sb.WriteString("\n=== Party Voting by Topic ===\n")

	// Sort parties
	var parties []string
	for p := range byParty {
		parties = append(parties, p)
	}
	sort.Strings(parties)

	for _, party := range parties {
		sb.WriteString(fmt.Sprintf("\n--- %s ---\n", party))
		sb.WriteString(fmt.Sprintf("%-25s %8s %8s %8s\n", "Topic", "For%", "For", "Against"))
		sb.WriteString(strings.Repeat("-", 55) + "\n")

		for _, v := range byParty[party] {
			sb.WriteString(fmt.Sprintf("%-25s %7.0f%% %8d %8d\n",
				truncate(v.CategoryName, 25),
				v.ForPct,
				v.VotesFor,
				v.VotesAgainst,
			))
		}
	}

	return sb.String()
}

func truncate(s string, maxLen int) string {
	if len(s) <= maxLen {
		return s
	}
	return s[:maxLen-1] + "…"
}
