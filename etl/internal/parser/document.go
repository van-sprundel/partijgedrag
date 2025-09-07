package parser

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"strings"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

type DocumentParser struct{}

func NewDocumentParser() *DocumentParser {
	return &DocumentParser{}
}

type Al struct {
	Content string `xml:",chardata"`
}

type Tekst struct {
	Al []Al `xml:"al"`
}

type VrijeTekst struct {
	Tekst Tekst `xml:"tekst"`
}

type Algemeen struct {
	VrijeTekst VrijeTekst `xml:"vrije-tekst"`
}

type Stuk struct {
	Algemeen Algemeen `xml:"algemeen"`
}

type Kamerstuk struct {
	Stuk Stuk `xml:"stuk"`
}

type OfficielePublicatie struct {
	Kamerstuk Kamerstuk `xml:"kamerstuk"`
}

// DocumentResult contains the parsed bullet points and document URL
type DocumentResult struct {
	BulletPoints []string
	URL          string
}

func (p *DocumentParser) ExtractBulletPoints(xmlData []byte, documentURL string) (*DocumentResult, error) {
	var doc OfficielePublicatie

	decoder := xml.NewDecoder(bytes.NewReader(xmlData))
	decoder.CharsetReader = func(charset string, input io.Reader) (io.Reader, error) {
		switch strings.ToLower(charset) {
		case "us-ascii":
			return input, nil
		case "iso-8859-1", "latin1":
			return transform.NewReader(input, charmap.ISO8859_1.NewDecoder()), nil
		case "windows-1252":
			return transform.NewReader(input, charmap.Windows1252.NewDecoder()), nil
		default:
			return nil, fmt.Errorf("unsupported charset: %s", charset)
		}
	}

	if err := decoder.Decode(&doc); err != nil {
		return nil, fmt.Errorf("parsing XML: %w", err)
	}

	var bulletPoints []string

	// skip common meaningless starting points
	skipPhrases := map[string]struct{}{
		"De Kamer,":                 {},
		"gehoord de beraadslaging,": {},
	}

	for _, al := range doc.Kamerstuk.Stuk.Algemeen.VrijeTekst.Tekst.Al {
		content := strings.TrimSpace(al.Content)
		if content != "" {
			if _, shouldSkip := skipPhrases[content]; !shouldSkip {
				bulletPoints = append(bulletPoints, content)
			}
		}
	}

	return &DocumentResult{
		BulletPoints: bulletPoints,
		URL:          documentURL,
	}, nil
}
