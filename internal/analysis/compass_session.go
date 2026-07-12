package analysis

import (
	"context"
	"crypto/rand"
	"encoding/base32"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

const maxCompassAnswers = 100

type CompassAnswer struct {
	MotionKey string `json:"motionKey"`
	Answer    string `json:"answer"`
}

type CompassSession struct {
	SessionKey   string
	Jurisdiction string
	Answers      []CompassAnswer
	MinOverlap   int
	CreatedAt    time.Time
}

type CompassMatch struct {
	PartySourceID string
	PartyName     string
	SameVotes     int
	Overlap       int
	Match         float64
}

type CompassMotionResult struct {
	Motion     VotingCompassMotion
	UserAnswer string
	Positions  []CompassMotionPosition
}

type CompassMotionPosition struct {
	PartySourceID  *string
	PartyName      string
	Position       string
	AgreesWithUser bool
}

type CompassResults struct {
	Session      CompassSession
	Matches      []CompassMatch
	Inconclusive []CompassMatch
	Threshold    int
	Motions      []CompassMotionResult
}

func ValidateCompassAnswers(answers []CompassAnswer) error {
	if len(answers) == 0 {
		return fmt.Errorf("answers must not be empty")
	}
	if len(answers) > maxCompassAnswers {
		return fmt.Errorf("too many answers, maximum is %d", maxCompassAnswers)
	}
	seen := map[string]bool{}
	for _, answer := range answers {
		if strings.TrimSpace(answer.MotionKey) == "" {
			return fmt.Errorf("answer is missing a motion key")
		}
		if answer.Answer != "FOR" && answer.Answer != "AGAINST" {
			return fmt.Errorf("answer for %s must be FOR or AGAINST", answer.MotionKey)
		}
		if seen[answer.MotionKey] {
			return fmt.Errorf("duplicate answer for motion %s", answer.MotionKey)
		}
		seen[answer.MotionKey] = true
	}
	return nil
}

func SaveCompassSession(ctx context.Context, pool *pgxpool.Pool, jurisdiction string, answers []CompassAnswer, minOverlap int) (string, error) {
	if jurisdiction == "" {
		jurisdiction = "nl-tweede-kamer"
	}
	if err := ValidateCompassAnswers(answers); err != nil {
		return "", err
	}
	if minOverlap < 1 {
		minOverlap = 1
	}
	if minOverlap > 50 {
		minOverlap = 50
	}

	sessionKey, err := newSessionKey()
	if err != nil {
		return "", err
	}
	payload, err := json.Marshal(answers)
	if err != nil {
		return "", err
	}

	_, err = pool.Exec(ctx, `
		INSERT INTO compass_sessions (session_key, jurisdiction_key, answers, min_overlap)
		VALUES ($1, $2, $3, $4)
	`, sessionKey, jurisdiction, payload, minOverlap)
	if err != nil {
		return "", err
	}
	return sessionKey, nil
}

func LoadCompassSession(ctx context.Context, pool *pgxpool.Pool, sessionKey string) (CompassSession, error) {
	session := CompassSession{}
	var payload []byte
	err := pool.QueryRow(ctx, `
		SELECT session_key, jurisdiction_key, answers, min_overlap, created_at
		FROM compass_sessions
		WHERE session_key = $1
	`, sessionKey).Scan(&session.SessionKey, &session.Jurisdiction, &payload, &session.MinOverlap, &session.CreatedAt)
	if err != nil {
		return CompassSession{}, err
	}
	if err := json.Unmarshal(payload, &session.Answers); err != nil {
		return CompassSession{}, err
	}
	return session, nil
}

// ScoreCompassSession recomputes party matches from the stored answers using
// the same semantics as the live compass page: a party scores on every
// answered motion where it took a clear position, and the match percentage is
// the share of those positions equal to the user's answer.
func ScoreCompassSession(ctx context.Context, pool *pgxpool.Pool, session CompassSession) (CompassResults, error) {
	motionKeys := make([]string, 0, len(session.Answers))
	answerByMotion := map[string]string{}
	for _, answer := range session.Answers {
		motionKeys = append(motionKeys, answer.MotionKey)
		answerByMotion[answer.MotionKey] = answer.Answer
	}

	motions, err := loadCompassMotionPositions(ctx, pool, session.Jurisdiction, motionKeys)
	if err != nil {
		return CompassResults{}, err
	}

	type score struct {
		partySourceID string
		partyName     string
		same          int
		overlap       int
	}
	scores := map[string]*score{}
	motionResults := make([]CompassMotionResult, 0, len(motions))
	for _, motion := range motions {
		userAnswer := answerByMotion[motion.MotionKey]
		result := CompassMotionResult{
			Motion:     motion,
			UserAnswer: userAnswer,
		}
		for _, position := range motion.Positions {
			agrees := position.Position == userAnswer
			result.Positions = append(result.Positions, CompassMotionPosition{
				PartySourceID:  position.PartySourceID,
				PartyName:      position.PartyName,
				Position:       position.Position,
				AgreesWithUser: agrees,
			})
			if position.PartySourceID == nil {
				continue
			}
			entry := scores[*position.PartySourceID]
			if entry == nil {
				entry = &score{partySourceID: *position.PartySourceID, partyName: position.PartyName}
				scores[*position.PartySourceID] = entry
			}
			entry.overlap++
			if agrees {
				entry.same++
			}
		}
		motionResults = append(motionResults, result)
	}

	threshold := session.MinOverlap
	if threshold > len(session.Answers) {
		threshold = len(session.Answers)
	}
	if threshold < 1 {
		threshold = 1
	}

	all := make([]CompassMatch, 0, len(scores))
	for _, entry := range scores {
		match := CompassMatch{
			PartySourceID: entry.partySourceID,
			PartyName:     entry.partyName,
			SameVotes:     entry.same,
			Overlap:       entry.overlap,
		}
		if entry.overlap > 0 {
			match.Match = (float64(entry.same) / float64(entry.overlap)) * 100
		}
		all = append(all, match)
	}
	sort.Slice(all, func(i, j int) bool {
		if all[i].Match != all[j].Match {
			return all[i].Match > all[j].Match
		}
		if all[i].Overlap != all[j].Overlap {
			return all[i].Overlap > all[j].Overlap
		}
		return strings.ToLower(all[i].PartyName) < strings.ToLower(all[j].PartyName)
	})

	results := CompassResults{
		Session:   session,
		Threshold: threshold,
		Motions:   motionResults,
	}
	for _, match := range all {
		if match.Overlap >= threshold {
			results.Matches = append(results.Matches, match)
		} else {
			results.Inconclusive = append(results.Inconclusive, match)
		}
	}
	sort.Slice(results.Inconclusive, func(i, j int) bool {
		if results.Inconclusive[i].Overlap != results.Inconclusive[j].Overlap {
			return results.Inconclusive[i].Overlap > results.Inconclusive[j].Overlap
		}
		return strings.ToLower(results.Inconclusive[i].PartyName) < strings.ToLower(results.Inconclusive[j].PartyName)
	})

	return results, nil
}

// loadCompassMotionPositions returns the requested motions with each party's
// clear (non-tied) position, preserving the order in which the user answered.
func loadCompassMotionPositions(ctx context.Context, pool *pgxpool.Pool, jurisdiction string, motionKeys []string) ([]VotingCompassMotion, error) {
	if jurisdiction == "" {
		jurisdiction = "nl-tweede-kamer"
	}
	if len(motionKeys) == 0 {
		return []VotingCompassMotion{}, nil
	}

	rows, err := pool.Query(ctx, `
		WITH requested AS (
			SELECT motion_key, ordinality
			FROM unnest($2::text[]) WITH ORDINALITY AS request(motion_key, ordinality)
		),
		party_positions AS (
			SELECT v.motion_key,
			       v.party_source_id,
			       COALESCE(p.short_name, v.party_name, v.actor_name, v.party_source_id, 'unknown') AS party_name,
			       CASE
			         WHEN SUM(CASE WHEN v.vote_type = 'Voor' THEN 1 ELSE 0 END) > SUM(CASE WHEN v.vote_type = 'Tegen' THEN 1 ELSE 0 END) THEN 'FOR'
			         ELSE 'AGAINST'
			       END AS position
			FROM votes v
			JOIN requested r ON r.motion_key = v.motion_key
			LEFT JOIN parties p ON p.source_key = v.source_key
			                   AND p.source_id = v.party_source_id
			WHERE v.source_deleted = false
			  AND v.mistake = false
			  AND v.vote_type IN ('Voor', 'Tegen')
			GROUP BY v.motion_key,
			         v.party_source_id,
			         COALESCE(p.short_name, v.party_name, v.actor_name, v.party_source_id, 'unknown')
			HAVING SUM(CASE WHEN v.vote_type = 'Voor' THEN 1 ELSE 0 END) <> SUM(CASE WHEN v.vote_type = 'Tegen' THEN 1 ELSE 0 END)
		)
		SELECT m.motion_key,
		       m.number,
		       m.title,
		       m.subject,
		       m.proposed_at,
		       pp.party_source_id,
		       pp.party_name,
		       pp.position
		FROM motions m
		JOIN requested r ON r.motion_key = m.motion_key
		LEFT JOIN party_positions pp ON pp.motion_key = m.motion_key
		WHERE m.jurisdiction_key = $1
		  AND m.source_deleted = false
		ORDER BY r.ordinality, pp.party_name
	`, jurisdiction, motionKeys)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	motions := []VotingCompassMotion{}
	index := map[string]int{}
	for rows.Next() {
		var motionKey string
		var number, title, subject *string
		var proposedAt *time.Time
		var partySourceID *string
		var partyName, position *string
		if err := rows.Scan(&motionKey, &number, &title, &subject, &proposedAt, &partySourceID, &partyName, &position); err != nil {
			return nil, err
		}

		at, seen := index[motionKey]
		if !seen {
			motions = append(motions, VotingCompassMotion{
				MotionKey:  motionKey,
				Number:     number,
				Title:      title,
				Subject:    subject,
				ProposedAt: proposedAt,
				Positions:  []VotingCompassPosition{},
			})
			at = len(motions) - 1
			index[motionKey] = at
		}
		if partyName != nil && position != nil {
			motions[at].Positions = append(motions[at].Positions, VotingCompassPosition{
				PartySourceID: partySourceID,
				PartyName:     *partyName,
				Position:      *position,
			})
		}
	}
	return motions, rows.Err()
}

func newSessionKey() (string, error) {
	buffer := make([]byte, 10)
	if _, err := rand.Read(buffer); err != nil {
		return "", err
	}
	return strings.ToLower(base32.StdEncoding.WithPadding(base32.NoPadding).EncodeToString(buffer)), nil
}
