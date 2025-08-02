package parser

import (
	"encoding/xml"
	"fmt"
	"strings"
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

func (p *DocumentParser) ExtractBulletPoints(xmlData []byte) ([]string, error) {
	var doc OfficielePublicatie

	if err := xml.Unmarshal(xmlData, &doc); err != nil {
		return nil, fmt.Errorf("parsing XML: %w", err)
	}

	var bulletPoints []string

	for _, al := range doc.Kamerstuk.Stuk.Algemeen.VrijeTekst.Tekst.Al {
		content := strings.TrimSpace(al.Content)
		if content != "" {
			bulletPoints = append(bulletPoints, content)
		}
	}

	return bulletPoints, nil
}
