package parser

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"regexp"
	"strings"
	"time"

	"etl/internal/models"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

type Parser struct{}

func NewParser() *Parser {
	return &Parser{}
}

// parse entire feed
func (p *Parser) ParseFeed(data []byte) (*models.Feed, error) {
	var feed models.Feed

	if err := p.unmarshalWithCharsetSupport(data, &feed); err != nil {
		return nil, fmt.Errorf("parsing feed XML: %w", err)
	}

	return &feed, nil
}

// parse a specific document XML with proper type detection
// parse a specific document XML
func (p *Parser) ParseDocument(data []byte) (*models.ParsedDocument, error) {
	// First, try to determine the document type by looking at the root element
	if strings.Contains(string(data), "<kamerwrk") {
		return p.parseKamerDocument(data)
	} else if strings.Contains(string(data), "<officiele-publicatie") {
		return p.parseOfficielePublicatie(data)
	}

	// If we can't determine the type, try kamerwrk first (legacy behavior)
	parsed, err := p.parseKamerDocument(data)
	if err == nil {
		return parsed, nil
	}

	// If that fails, try officiele-publicatie
	parsed, err = p.parseOfficielePublicatie(data)
	if err == nil {
		return parsed, nil
	}

	return nil, fmt.Errorf("unable to parse document: unknown document type or invalid XML")
}

// parse kamerwrk document type
func (p *Parser) parseKamerDocument(data []byte) (*models.ParsedDocument, error) {
	var doc models.KamerDocument

	if err := p.unmarshalWithCharsetSupport(data, &doc); err != nil {
		return nil, fmt.Errorf("parsing kamerwrk document XML: %w", err)
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

// parse officiele-publicatie document type with proper structure handling
// parse officiele-publicatie document type
func (p *Parser) parseOfficielePublicatie(data []byte) (*models.ParsedDocument, error) {
	var doc models.OfficielePublicatie

	if err := p.unmarshalWithCharsetSupport(data, &doc); err != nil {
		return nil, fmt.Errorf("parsing officiele-publicatie document XML: %w", err)
	}

	if doc.Kamerstuk == nil {
		return nil, fmt.Errorf("officiele-publicatie document missing kamerstuk element")
	}

	// Extract basic information
	vergaderjaar := ""
	kamer := ""
	for _, regel := range doc.Kamerstuk.Kamerstukkop.Tekstregels {
		switch regel.Inhoud {
		case "vergaderjaar":
			vergaderjaar = regel.Value
		case "kameraanduiding":
			kamer = regel.Value
		}
	}

	parsed := &models.ParsedDocument{
		ID:           generateOfficielePublicatieID(doc),
		DocumentType: "officiele-publicatie",
		Chamber:      kamer,
		OrderNumber:  doc.Kamerstuk.Stuk.Stuknr.Ondernummer.Value,
		Year:         vergaderjaar,
		Subject: models.ParsedSubject{
			Number: getPrimaryDossierNumber(doc.Kamerstuk.Dossier),
			Name:   getPrimaryDossierTitle(doc.Kamerstuk.Dossier),
		},
		Title:     doc.Kamerstuk.Stuk.Titel,
		Number:    doc.Kamerstuk.Stuk.Stuknr.Ondernummer.Value,
		Authors:   extractOfficielePublicatieAuthors(doc),
		Sections:  extractOfficielePublicatieSections(doc),
		Footnotes: extractOfficielePublicatieFootnotes(doc),
		Metadata:  extractOfficielePublicatieMetadata(doc),
		FullText:  extractOfficielePublicatieFullText(doc),
		Date:      parseOfficielePublicatieDate(doc),
		Dossiers:  extractDossiers(doc),
	}

	return parsed, nil
}

// generate unique ID for officiele-publicatie documents
func generateOfficielePublicatieID(doc models.OfficielePublicatie) string {
	if doc.Kamerstuk == nil || len(doc.Kamerstuk.Dossier) == 0 {
		return "unknown"
	}
	return fmt.Sprintf("%s-%s", getPrimaryDossierNumber(doc.Kamerstuk.Dossier), doc.Kamerstuk.Stuk.Stuknr.Ondernummer.Value)
}

// generate unique ID for kamer documents
func generateDocumentID(doc models.KamerDocument) string {
	return fmt.Sprintf("%s-%s-%s", doc.Front.Subject.Number, doc.Front.OrderNum, doc.Front.Year)
}

// extract authors from officiele-publicatie documents with proper signature handling
func extractOfficielePublicatieAuthors(doc models.OfficielePublicatie) []models.ParsedAuthor {
	var authors []models.ParsedAuthor

	if doc.Kamerstuk == nil {
		return authors
	}

	// Look for signatures in TekstSluiting sections
	for _, algemeen := range doc.Kamerstuk.Stuk.Algemeen {
		if algemeen.TekstSluiting != nil && algemeen.TekstSluiting.Ondertekening.Naam.Voornaam != "" {
			fullName := strings.TrimSpace(algemeen.TekstSluiting.Ondertekening.Naam.Voornaam + " " + algemeen.TekstSluiting.Ondertekening.Naam.Achternaam)
			authors = append(authors, models.ParsedAuthor{
				Function: strings.TrimSpace(algemeen.TekstSluiting.Ondertekening.Functie),
				Name:     fullName,
			})
		}
	}

	return authors
}

// extract sections from officiele-publicatie documents with proper handling of al-groep and complex text
func extractOfficielePublicatieSections(doc models.OfficielePublicatie) []models.ParsedSection {
	var sections []models.ParsedSection
	order := 0

	if doc.Kamerstuk == nil {
		return sections
	}

	for _, algemeen := range doc.Kamerstuk.Stuk.Algemeen {
		if algemeen.VrijeTekst != nil {
			// Process regular als
			for _, al := range algemeen.VrijeTekst.Tekst.Als {
				order++
				content := extractAlContent(al)
				if strings.TrimSpace(content) != "" {
					sections = append(sections, models.ParsedSection{
						Level:   "paragraph",
						Title:   fmt.Sprintf("Paragraph %d", order),
						Content: content,
						Order:   order,
					})
				}
			}

			// Process al-groep structures
			for _, alGroep := range algemeen.VrijeTekst.Tekst.AlGroep {
				order++
				var groupContent []string
				for _, al := range alGroep.Als {
					content := extractAlContent(al)
					if strings.TrimSpace(content) != "" {
						groupContent = append(groupContent, content)
					}
				}
				if len(groupContent) > 0 {
					sections = append(sections, models.ParsedSection{
						Level:   "group",
						Title:   fmt.Sprintf("Group %d", order),
						Content: strings.Join(groupContent, "\n"),
						Order:   order,
					})
				}
			}
		}
	}

	return sections
}

// extract footnotes from officiele-publicatie documents with proper noot handling
func extractOfficielePublicatieFootnotes(doc models.OfficielePublicatie) []models.ParsedFootnote {
	var footnotes []models.ParsedFootnote

	if doc.Kamerstuk == nil {
		return footnotes
	}

	for _, algemeen := range doc.Kamerstuk.Stuk.Algemeen {
		if algemeen.VrijeTekst != nil {
			footnotes = append(footnotes, extractFootnotesFromTekst(algemeen.VrijeTekst.Tekst)...)
		}
	}

	return footnotes
}

// extract footnotes from tekst content
func extractFootnotesFromTekst(tekst models.TekstContent) []models.ParsedFootnote {
	var footnotes []models.ParsedFootnote

	// Extract from regular als
	for _, al := range tekst.Als {
		footnotes = append(footnotes, extractFootnotesFromAl(al)...)
	}

	// Extract from al-groep
	for _, alGroep := range tekst.AlGroep {
		for _, al := range alGroep.Als {
			footnotes = append(footnotes, extractFootnotesFromAl(al)...)
		}
	}

	return footnotes
}

// extract footnotes from al element
func extractFootnotesFromAl(al models.Al) []models.ParsedFootnote {
	var footnotes []models.ParsedFootnote

	for _, noot := range al.Noot {
		content := strings.TrimSpace(noot.NootAl.Text)

		// Add external references to footnote content
		for _, extRef := range noot.NootAl.ExtRef {
			if extRef.Value != "" {
				content += " " + extRef.Value
			}
		}

		if content != "" {
			footnotes = append(footnotes, models.ParsedFootnote{
				ID:      noot.ID,
				Number:  noot.NootNr,
				Content: content,
			})
		}
	}

	return footnotes
}

// extract content from al element including formatting
func extractAlContent(al models.Al) string {
	var parts []string

	// Add main text
	if strings.TrimSpace(al.Text) != "" {
		parts = append(parts, strings.TrimSpace(al.Text))
	}

	// Add formatted text (nadruk)
	for _, nadruk := range al.Nadruk {
		if strings.TrimSpace(nadruk.Value) != "" {
			parts = append(parts, strings.TrimSpace(nadruk.Value))
		}
	}

	// Add external references
	for _, extRef := range al.ExtRef {
		if strings.TrimSpace(extRef.Value) != "" {
			parts = append(parts, strings.TrimSpace(extRef.Value))
		}
	}

	return strings.Join(parts, " ")
}

// extract all dossiers from the document
func extractDossiers(doc models.OfficielePublicatie) []models.ParsedDossier {
	var dossiers []models.ParsedDossier

	if doc.Kamerstuk == nil {
		return dossiers
	}

	for _, dossier := range doc.Kamerstuk.Dossier {
		dossiers = append(dossiers, models.ParsedDossier{
			Number: dossier.Dossiernummer.Dossiernr,
			Title:  dossier.Titel,
		})
	}

	return dossiers
}

// parse date from officiele-publicatie document
func parseOfficielePublicatieDate(doc models.OfficielePublicatie) *time.Time {
	if doc.Kamerstuk == nil || doc.Kamerstuk.Stuk.Datumtekst.Datum.Isodatum == "" {
		return nil
	}

	if date, err := time.Parse("2006-01-02", doc.Kamerstuk.Stuk.Datumtekst.Datum.Isodatum); err == nil {
		return &date
	}

	return nil
}

// extract metadata from officiele-publicatie documents
func extractOfficielePublicatieMetadata(doc models.OfficielePublicatie) map[string]interface{} {
	metadata := make(map[string]interface{})

	metadata["document_type"] = "officiele-publicatie"

	// Extract metadata records
	metaRecords := make(map[string]string)
	for _, meta := range doc.Metadata.Meta {
		metaRecords[meta.Name] = meta.Content
	}
	if len(metaRecords) > 0 {
		metadata["meta_records"] = metaRecords
	}

	// Extract kamerstukkop information
	if doc.Kamerstuk != nil {
		kamerInfo := make(map[string]string)
		for _, regel := range doc.Kamerstuk.Kamerstukkop.Tekstregels {
			kamerInfo[regel.Inhoud] = regel.Value
		}
		if len(kamerInfo) > 0 {
			metadata["kamerstukkop"] = kamerInfo
		}

		// Add dossier information
		if len(doc.Kamerstuk.Dossier) > 0 {
			metadata["primary_dossier_nummer"] = getPrimaryDossierNumber(doc.Kamerstuk.Dossier)
			metadata["primary_dossier_titel"] = getPrimaryDossierTitle(doc.Kamerstuk.Dossier)
			metadata["all_dossiers"] = extractDossiers(doc)
		}

		// Add stuk information
		metadata["stuk_nummer"] = doc.Kamerstuk.Stuk.Stuknr.Ondernummer.Value
		metadata["stuk_titel"] = doc.Kamerstuk.Stuk.Titel

		// Add date information
		if doc.Kamerstuk.Stuk.Datumtekst.Datum.Isodatum != "" {
			metadata["datum_iso"] = doc.Kamerstuk.Stuk.Datumtekst.Datum.Isodatum
			metadata["datum_text"] = doc.Kamerstuk.Stuk.Datumtekst.Datum.Value
		}
	}

	return metadata
}

// extract full text from officiele-publicatie documents
func extractOfficielePublicatieFullText(doc models.OfficielePublicatie) string {
	var textParts []string

	if doc.Kamerstuk == nil {
		return ""
	}

	// Add title
	if doc.Kamerstuk.Stuk.Titel != "" {
		textParts = append(textParts, doc.Kamerstuk.Stuk.Titel)
	}

	// Add dossier titles
	for _, dossier := range doc.Kamerstuk.Dossier {
		if dossier.Titel != "" {
			textParts = append(textParts, dossier.Titel)
		}
	}

	// Add all text content
	for _, algemeen := range doc.Kamerstuk.Stuk.Algemeen {
		if algemeen.VrijeTekst != nil {
			// Add regular als
			for _, al := range algemeen.VrijeTekst.Tekst.Als {
				content := extractAlContent(al)
				if strings.TrimSpace(content) != "" {
					textParts = append(textParts, content)
				}
			}

			// Add al-groep content
			for _, alGroep := range algemeen.VrijeTekst.Tekst.AlGroep {
				for _, al := range alGroep.Als {
					content := extractAlContent(al)
					if strings.TrimSpace(content) != "" {
						textParts = append(textParts, content)
					}
				}
			}
		}
	}

	return joinParagraphs(textParts)
}

// Legacy kamer document methods (kept for compatibility)
func extractAuthors(doc models.KamerDocument) []models.ParsedAuthor {
	var authors []models.ParsedAuthor

	for _, piece := range doc.Body.Pieces {
		for _, sig := range piece.Signatures {
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

func extractSections(doc models.KamerDocument) []models.ParsedSection {
	var sections []models.ParsedSection
	order := 0

	for _, piece := range doc.Body.Pieces {
		content := extractPieceContent(piece)

		for _, section := range piece.Sections {
			order++
			sections = append(sections, models.ParsedSection{
				Level:   section.Level,
				Title:   section.Title,
				Content: section.Title,
				Order:   order,
			})
		}

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

// getPrimaryDossierNumber returns the first dossier number (primary) from multiple dossiers
func getPrimaryDossierNumber(dossiers []models.Dossier) string {
	if len(dossiers) == 0 {
		return ""
	}
	return dossiers[0].Dossiernummer.Dossiernr
}

// getPrimaryDossierTitle returns the first dossier title (primary) from multiple dossiers
func getPrimaryDossierTitle(dossiers []models.Dossier) string {
	if len(dossiers) == 0 {
		return ""
	}
	return dossiers[0].Titel
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

// unmarshalWithCharsetSupport handles XML with different character encodings
func (p *Parser) unmarshalWithCharsetSupport(data []byte, v interface{}) error {
	decoder := xml.NewDecoder(bytes.NewReader(data))

	// Set up charset reader to handle different encodings
	decoder.CharsetReader = func(charset string, input io.Reader) (io.Reader, error) {
		switch strings.ToLower(charset) {
		case "us-ascii", "ascii":
			// US-ASCII is a subset of UTF-8, so we can read it directly
			return input, nil
		case "iso-8859-1", "latin-1":
			// Convert ISO-8859-1 to UTF-8
			return transform.NewReader(input, charmap.ISO8859_1.NewDecoder()), nil
		case "windows-1252", "cp1252":
			// Convert Windows-1252 to UTF-8
			return transform.NewReader(input, charmap.Windows1252.NewDecoder()), nil
		case "utf-8", "":
			// UTF-8 or no charset specified, read as-is
			return input, nil
		default:
			// For unknown charsets, try to read as-is and hope for the best
			return input, nil
		}
	}

	return decoder.Decode(v)
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
