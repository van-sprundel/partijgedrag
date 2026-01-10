package analysis

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"time"

	"etl/internal/models"
	"etl/pkg/storage"
)

type Service struct {
	store storage.Storage
}

func NewService(store storage.Storage) *Service {
	return &Service{store: store}
}

// CoalitionAnalysis returns party alignment data, optionally filtered by year
func (s *Service) CoalitionAnalysis(ctx context.Context, period string) (*models.AnalysisResult, error) {
	if period == "" {
		period = "all"
	}

	alignments, err := s.store.GetCoalitionAlignments(ctx, period)
	if err != nil {
		return nil, fmt.Errorf("getting coalition alignments: %w", err)
	}

	return &models.AnalysisResult{
		Type:        "coalition_alignment",
		Period:      period,
		GeneratedAt: time.Now(),
		Data:        alignments,
	}, nil
}

// CoalitionMatrix returns alignment data formatted as a matrix for visualization
func (s *Service) CoalitionMatrix(ctx context.Context, period string) (*models.AnalysisResult, error) {
	if period == "" {
		period = "all"
	}

	alignments, err := s.store.GetCoalitionAlignments(ctx, period)
	if err != nil {
		return nil, fmt.Errorf("getting coalition alignments: %w", err)
	}

	// Build matrix structure
	parties := make(map[string]bool)
	for _, a := range alignments {
		parties[a.Fractie1Name] = true
		parties[a.Fractie2Name] = true
	}

	var partyList []string
	for p := range parties {
		partyList = append(partyList, p)
	}
	sort.Strings(partyList)

	// Create matrix
	matrix := make(map[string]map[string]float64)
	for _, p := range partyList {
		matrix[p] = make(map[string]float64)
		matrix[p][p] = 100.0 // Self-alignment is always 100%
	}

	for _, a := range alignments {
		matrix[a.Fractie1Name][a.Fractie2Name] = a.AlignmentPct
		matrix[a.Fractie2Name][a.Fractie1Name] = a.AlignmentPct
	}

	return &models.AnalysisResult{
		Type:        "coalition_matrix",
		Period:      period,
		GeneratedAt: time.Now(),
		Data: map[string]interface{}{
			"parties": partyList,
			"matrix":  matrix,
		},
	}, nil
}

// MPDeviationAnalysis finds MPs who vote against their party most often
func (s *Service) MPDeviationAnalysis(ctx context.Context, period string, limit int) (*models.AnalysisResult, error) {
	if period == "" {
		period = "all"
	}
	if limit <= 0 {
		limit = 50
	}

	deviations, err := s.store.GetMPDeviations(ctx, period, limit)
	if err != nil {
		return nil, fmt.Errorf("getting MP deviations: %w", err)
	}

	return &models.AnalysisResult{
		Type:        "mp_deviation",
		Period:      period,
		GeneratedAt: time.Now(),
		Data:        deviations,
	}, nil
}

// TopicTrendAnalysis returns motion counts by category
func (s *Service) TopicTrendAnalysis(ctx context.Context, period string) (*models.AnalysisResult, error) {
	if period == "" {
		period = "all"
	}

	trends, err := s.store.GetTopicTrends(ctx, period)
	if err != nil {
		return nil, fmt.Errorf("getting topic trends: %w", err)
	}

	return &models.AnalysisResult{
		Type:        "topic_trends",
		Period:      period,
		GeneratedAt: time.Now(),
		Data:        trends,
	}, nil
}

// PartyTopicAnalysis shows how a party votes on different topics
func (s *Service) PartyTopicAnalysis(ctx context.Context, fractieID string, period string) (*models.AnalysisResult, error) {
	if period == "" {
		period = "all"
	}

	voting, err := s.store.GetPartyTopicVoting(ctx, fractieID, period)
	if err != nil {
		return nil, fmt.Errorf("getting party topic voting: %w", err)
	}

	return &models.AnalysisResult{
		Type:        "party_topic_voting",
		Period:      period,
		GeneratedAt: time.Now(),
		Data:        voting,
	}, nil
}

// FullReport generates a complete analysis report
func (s *Service) FullReport(ctx context.Context, period string) (map[string]*models.AnalysisResult, error) {
	results := make(map[string]*models.AnalysisResult)

	coalition, err := s.CoalitionAnalysis(ctx, period)
	if err != nil {
		return nil, err
	}
	results["coalition"] = coalition

	matrix, err := s.CoalitionMatrix(ctx, period)
	if err != nil {
		return nil, err
	}
	results["matrix"] = matrix

	deviations, err := s.MPDeviationAnalysis(ctx, period, 25)
	if err != nil {
		return nil, err
	}
	results["deviations"] = deviations

	topics, err := s.TopicTrendAnalysis(ctx, period)
	if err != nil {
		return nil, err
	}
	results["topics"] = topics

	return results, nil
}

// ToJSON converts an analysis result to JSON
func (s *Service) ToJSON(result *models.AnalysisResult) (string, error) {
	data, err := json.MarshalIndent(result, "", "  ")
	if err != nil {
		return "", err
	}
	return string(data), nil
}
