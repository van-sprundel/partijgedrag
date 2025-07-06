package parser

import (
	"encoding/xml"
	"fmt"
	"regexp"
	"strings"

	"etl/internal/models"
)

type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

// parse entire feed
func (p *Parser) ParseFeed(data []byte) (*models.Feed, error) {
	var feed models.Feed

	if err := xml.Unmarshal(data, &feed); err != nil {
		return nil, fmt.Errorf("parsing feed XML: %w", err)
	}

	return &feed, nil
}

// parse a specific document XML
func (p *Parser) ParseDocument(data []byte) (*models.ParsedDocument, error) {
	var doc models.KamerDocument

	if err := xml.Unmarshal(data, &doc); err != nil {
		return nil, fmt.Errorf("parsing document XML: %w", err)
	}

	parsed := &models.ParsedDocument{
		ID:           generateDocumentID(doc),
		DocumentType: doc.PublType,
		Chamber:      doc.Kamer,
		OrderNumber:  doc.Front.OrderNum,
		Year:         doc.Front.Year,
		Subject: models.ParsedSubject{
			Number: doc.Front.Subject.Number,
			Name:   doc.Front.Subject.Name,
		},
		Authors:   extractAuthors(doc),
		Sections:  extractSections(doc),
		Footnotes: extractFootnotes(doc),
		Metadata:  extractMetadata(doc),
		FullText:  extractFullText(doc),
	}

	if len(doc.Body.Pieces) > 0 {
		parsed.Title = doc.Body.Pieces[0].Title
		parsed.Number = doc.Body.Pieces[0].Number
	}

	return parsed, nil
}

// unique ID
func generateDocumentID(doc models.KamerDocument) string {
	return fmt.Sprintf("%s-%s-%s", doc.Front.Subject.Number, doc.Front.OrderNum, doc.Front.Year)
}

// extract author information from signatures
func extractAuthors(doc models.KamerDocument) []models.ParsedAuthor {
	var authors []models.ParsedAuthor

	for _, piece := range doc.Body.Pieces {
		for _, sig := range piece.Signatures {
			// handle multiple function/name pairs in a single sign block
			minLen := len(sig.Functions)
			if len(sig.Names) < minLen {
				minLen = len(sig.Names)
			}

			for i := 0; i < minLen; i++ {
				authors = append(authors, models.ParsedAuthor{
					Function: sig.Functions[i],
					Name:     sig.Names[i],
				})
			}
		}
	}

	return authors
}

// extract hierarchical sections
func extractSections(doc models.KamerDocument) []models.ParsedSection {
	var sections []models.ParsedSection
	order := 0

	for _, piece := range doc.Body.Pieces {
		content := extractPieceContent(piece)

		// (tuskop elements)
		for _, section := range piece.Sections {
			order++
			sections = append(sections, models.ParsedSection{
				Level:   section.Level,
				Title:   section.Title,
				Content: section.Title,
				Order:   order,
			})
		}

		// TODO should we provide a default or skip? idk
		if len(piece.Sections) == 0 && len(piece.Paragraphs) > 0 {
			order++
			sections = append(sections, models.ParsedSection{
				Level:   "default",
				Title:   piece.Title,
				Content: content,
				Order:   order,
			})
		}
	}

	return sections
}

func extractPieceContent(piece models.DocumentPiece) string {
	var content []string

	for _, para := range piece.Paragraphs {
		if para.Text != "" {
			content = append(content, para.Text)
		}
	}

	return joinParagraphs(content)
}

func extractFootnotes(doc models.KamerDocument) []models.ParsedFootnote {
	var footnotes []models.ParsedFootnote

	for _, footnote := range doc.Footnotes {
		var content []string
		for _, para := range footnote.Content {
			if para.Text != "" {
				content = append(content, para.Text)
			}
		}

		footnotes = append(footnotes, models.ParsedFootnote{
			ID:      footnote.ID,
			Number:  footnote.Number,
			Content: joinParagraphs(content),
		})
	}

	return footnotes
}

func extractMetadata(doc models.KamerDocument) map[string]interface{} {
	metadata := make(map[string]interface{})

	metadata["kamer"] = doc.Kamer
	metadata["publtype"] = doc.PublType
	metadata["order_number"] = doc.Front.OrderNum
	metadata["year"] = doc.Front.Year
	metadata["version"] = map[string]string{
		"dtd":    doc.Front.Version.DTD,
		"conv":   doc.Front.Version.Conv,
		"markup": doc.Front.Version.Markup,
	}

	if doc.Header.Title != "" {
		metadata["header_title"] = doc.Header.Title
	}
	if len(doc.Header.Subtitles) > 0 {
		metadata["header_subtitles"] = doc.Header.Subtitles
	}

	metaRecords := make(map[string]string)
	for _, meta := range doc.Metadata.MetaRecords {
		metaRecords[meta.Name] = meta.Content
	}
	if len(metaRecords) > 0 {
		metadata["meta_records"] = metaRecords
	}

	return metadata
}

func extractFullText(doc models.KamerDocument) string {
	var textParts []string

	if doc.Header.Title != "" {
		textParts = append(textParts, doc.Header.Title)
	}
	for _, subtitle := range doc.Header.Subtitles {
		textParts = append(textParts, subtitle)
	}

	if doc.Front.Subject.Name != "" {
		textParts = append(textParts, fmt.Sprintf("%s %s", doc.Front.Subject.Number, doc.Front.Subject.Name))
	}

	for _, piece := range doc.Body.Pieces {
		if piece.Title != "" {
			textParts = append(textParts, piece.Title)
		}

		for _, section := range piece.Sections {
			textParts = append(textParts, section.Title)
		}

		for _, para := range piece.Paragraphs {
			if para.Text != "" {
				textParts = append(textParts, para.Text)
			}
		}
	}

	for _, footnote := range doc.Footnotes {
		for _, para := range footnote.Content {
			if para.Text != "" {
				textParts = append(textParts, para.Text)
			}
		}
	}

	return joinParagraphs(textParts)
}

// join text parts with appropriate spacing
func joinParagraphs(parts []string) string {
	var cleanParts []string
	for _, part := range parts {
		if trimmed := strings.TrimSpace(part); trimmed != "" {
			cleanParts = append(cleanParts, trimmed)
		}
	}
	return strings.Join(cleanParts, "\n\n")
}

// extract all kamerstukdossier entries
func (p *Parser) ExtractKamerstukdossiers(feed *models.Feed) []*models.Kamerstukdossier {
	var kamerstukdossiers []*models.Kamerstukdossier

	for _, entry := range feed.Entries {
		if entry.Content.Kamerstukdossier != nil {
			kamerstukdossiers = append(kamerstukdossiers, entry.Content.Kamerstukdossier)
		}
	}

	return kamerstukdossiers
}

// extract pagination information from a feed
func (p *Parser) GetPaginationInfo(feed *models.Feed) *models.PaginationInfo {
	nextSkiptoken := ""

	// check feed nextlink first
	feedNextURL := feed.GetNextURL()
	if feedNextURL != "" {
		if token, err := extractSkiptoken(feedNextURL); err == nil {
			nextSkiptoken = token
		}
	}

	// if no feed next link, get the last entry's next link
	if nextSkiptoken == "" && len(feed.Entries) > 0 {
		// iterate backwards to find the last entry with a next link
		for i := len(feed.Entries) - 1; i >= 0; i-- {
			entry := feed.Entries[i]
			entryNextURL := entry.GetNextURL()
			if entryNextURL != "" {
				if token, err := extractSkiptoken(entryNextURL); err == nil {
					nextSkiptoken = token
					break
				}
			}
		}
	}

	hasMore := nextSkiptoken != ""

	return &models.PaginationInfo{
		NextSkiptoken: nextSkiptoken,
		HasMore:       hasMore,
	}
}

// extract the skiptoken from a URL
func extractSkiptoken(url string) (string, error) {
	if url == "" {
		return "", nil
	}

	re := regexp.MustCompile(`skiptoken=([^&]+)`)
	matches := re.FindStringSubmatch(url)

	if len(matches) < 2 {
		return "", fmt.Errorf("could not extract skiptoken from URL: %s", url)
	}

	return matches[1], nil
}
