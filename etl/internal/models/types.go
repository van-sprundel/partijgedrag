package models

import (
	"encoding/xml"
	"time"
)

// Atom feed structure
type Feed struct {
	XMLName xml.Name  `xml:"http://www.w3.org/2005/Atom feed"`
	Title   string    `xml:"title"`
	Updated time.Time `xml:"updated"`
	Links   []Link    `xml:"link"`
	Author  Author    `xml:"author"`
	Rights  string    `xml:"rights"`
	ID      string    `xml:"id"`
	Entries []Entry   `xml:"entry"`
}

// extract the next URL from the feed's links
func (f *Feed) GetNextURL() string {
	for _, link := range f.Links {
		if link.Rel == "next" {
			return link.Href
		}
	}
	return ""
}

type Author struct {
	Name  string `xml:"name"`
	URI   string `xml:"uri"`
	Email string `xml:"email"`
}

type Link struct {
	Rel  string `xml:"rel,attr"`
	Href string `xml:"href,attr"`
}

type Entry struct {
	Title    string    `xml:"title"`
	ID       string    `xml:"id"`
	Updated  time.Time `xml:"updated"`
	Category Category  `xml:"category"`
	Links    []Link    `xml:"link"`
	Content  Content   `xml:"content"`
}

// extract the next URL from the entry's links
func (e *Entry) GetNextURL() string {
	for _, link := range e.Links {
		if link.Rel == "next" {
			return link.Href
		}
	}
	return ""
}

type Category struct {
	Term string `xml:"term,attr"`
}

type Content struct {
	Type             string            `xml:"type,attr"`
	Kamerstukdossier *Kamerstukdossier `xml:"kamerstukdossier,omitempty"`
}

type Kamerstukdossier struct {
	XMLName           xml.Name `xml:"http://www.tweedekamer.nl/xsd/tkData/v1-0 kamerstukdossier"`
	ID                string   `xml:"id,attr"`
	Bijgewerkt        string   `xml:"bijgewerkt,attr"`
	Verwijderd        string   `xml:"verwijderd,attr"`
	Titel             string   `xml:"titel"`
	Citeertitel       *string  `xml:"citeertitel"`
	Alias             *string  `xml:"alias"`
	Nummer            string   `xml:"nummer"`
	Toevoeging        *string  `xml:"toevoeging"`
	HoogsteVolgnummer int      `xml:"hoogsteVolgnummer"`
	Afgesloten        bool     `xml:"afgesloten"`
	Kamer             string   `xml:"kamer"`
}

type PaginationInfo struct {
	CurrentSkiptoken string
	NextSkiptoken    string
	Category         string
	HasMore          bool
}

// for parsing parliamentary XML documents
type KamerDocument struct {
	XMLName   xml.Name         `xml:"kamerwrk"`
	Kamer     string           `xml:"kamer,attr"`
	PublType  string           `xml:"publtype,attr"`
	Metadata  DocumentMetadata `xml:"metadata"`
	Header    DocumentHeader   `xml:"kop"`
	Front     DocumentFront    `xml:"frontm"`
	Body      DocumentBody     `xml:"body"`
	Footnotes []Footnote       `xml:"voetnoot"`
}

type DocumentMetadata struct {
	MetaRecords []MetaRecord `xml:"meta"`
}

type MetaRecord struct {
	Name    string `xml:"name,attr"`
	Scheme  string `xml:"scheme,attr"`
	Content string `xml:"content,attr"`
}

type DocumentHeader struct {
	Title     string   `xml:"titel"`
	Subtitles []string `xml:"subtitel"`
}

type DocumentFront struct {
	Version  Version `xml:"versie"`
	OrderNum string  `xml:"ordernr"`
	Year     string  `xml:"vergjaar"`
	Subject  Subject `xml:"onderw"`
}

type Version struct {
	DTD    string `xml:"dtd,attr"`
	Conv   string `xml:"conv,attr"`
	Markup string `xml:"markup,attr"`
}

type Subject struct {
	Number string `xml:"nummer"`
	Name   string `xml:"naam"`
}

type DocumentBody struct {
	Pieces []DocumentPiece `xml:"stuk"`
}

type DocumentPiece struct {
	LabelText  string           `xml:"ltrlabel"`
	Number     string           `xml:"nummer"`
	Title      string           `xml:"titel"`
	Paragraphs []Paragraph      `xml:"al"`
	Whitespace []string         `xml:"witreg"`
	Signatures []Signature      `xml:"ondtek"`
	Sections   []Section        `xml:"tuskop"`
	Content    []ContentElement `xml:",any"`
}

type Paragraph struct {
	Text string `xml:",chardata"`
	Date string `xml:"datum"`
}

type Signature struct {
	Functions []string `xml:"functie"`
	Names     []string `xml:"naam"`
}

type Section struct {
	Level string `xml:"letat,attr"`
	Title string `xml:",chardata"`
}

type ContentElement struct {
	XMLName xml.Name
	Content string     `xml:",chardata"`
	Attrs   []xml.Attr `xml:",any,attr"`
}

type Footnote struct {
	ID      string      `xml:"id,attr"`
	Number  string      `xml:"nr,attr"`
	Content []Paragraph `xml:"al"`
}

// parsed document structure
type ParsedDocument struct {
	ID           string                 `json:"id"`
	DocumentType string                 `json:"document_type"`
	Chamber      string                 `json:"chamber"`
	OrderNumber  string                 `json:"order_number"`
	Year         string                 `json:"year"`
	Subject      ParsedSubject          `json:"subject"`
	Title        string                 `json:"title"`
	Number       string                 `json:"number"`
	Authors      []ParsedAuthor         `json:"authors"`
	Sections     []ParsedSection        `json:"sections"`
	Footnotes    []ParsedFootnote       `json:"footnotes"`
	Metadata     map[string]interface{} `json:"metadata"`
	FullText     string                 `json:"full_text"`
}

type ParsedSubject struct {
	Number string `json:"number"`
	Name   string `json:"name"`
}

type ParsedAuthor struct {
	Function string `json:"function"`
	Name     string `json:"name"`
}

type ParsedSection struct {
	Level   string `json:"level"`
	Title   string `json:"title"`
	Content string `json:"content"`
	Order   int    `json:"order"`
}

type ParsedFootnote struct {
	ID      string `json:"id"`
	Number  string `json:"number"`
	Content string `json:"content"`
}
