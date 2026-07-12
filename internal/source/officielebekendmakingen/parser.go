package officielebekendmakingen

import (
	"bytes"
	"encoding/xml"
	"fmt"
	"io"
	"regexp"
	"strings"

	"golang.org/x/text/encoding/charmap"
	"golang.org/x/text/transform"
)

var motionTitleRegex = regexp.MustCompile(`\bmotie\b`)

// bulletPrefixes mark the paragraphs that make up the body of a motion:
// the considerations (constaterende/overwegende) and the request (verzoekt).
var bulletPrefixes = []string{
	"constaterende",
	"overwegende",
	"verzoekt",
}

type al struct {
	Content string `xml:",chardata"`
}

type tekst struct {
	Al []al `xml:"al"`
}

type vrijeTekst struct {
	Tekst tekst `xml:"tekst"`
}

type algemeen struct {
	VrijeTekst vrijeTekst `xml:"vrije-tekst"`
}

type stuk struct {
	Titel    string   `xml:"titel"`
	Algemeen algemeen `xml:"algemeen"`
}

type kamerstuk struct {
	Stuk stuk `xml:"stuk"`
}

type officielePublicatie struct {
	Kamerstuk kamerstuk `xml:"kamerstuk"`
}

type DocumentResult struct {
	Title        string
	BulletPoints []string
}

// ExtractBulletPoints parses a kamerstuk XML publication and returns the
// motion's bullet points. It returns (nil, nil) when the document is not a motion.
func ExtractBulletPoints(xmlData []byte) (*DocumentResult, error) {
	var doc officielePublicatie

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

	title := strings.TrimSpace(doc.Kamerstuk.Stuk.Titel)
	if title == "" || !motionTitleRegex.MatchString(strings.ToLower(title)) {
		return nil, nil
	}

	var bulletPoints []string
	for _, paragraph := range doc.Kamerstuk.Stuk.Algemeen.VrijeTekst.Tekst.Al {
		content := strings.TrimSuffix(strings.TrimSpace(paragraph.Content), ",")
		if content == "" {
			continue
		}
		lowerContent := strings.ToLower(content)
		for _, prefix := range bulletPrefixes {
			if strings.HasPrefix(lowerContent, prefix) {
				bulletPoints = append(bulletPoints, content)
				break
			}
		}
	}

	return &DocumentResult{
		Title:        title,
		BulletPoints: bulletPoints,
	}, nil
}
