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

type Category struct {
	Term string `xml:"term,attr"`
}

type Content struct {
	Type             string            `xml:"type,attr"`
	Kamerstukdossier *Kamerstukdossier `xml:"kamerstukdossier,omitempty"`
	Besluit          *Besluit          `xml:"besluit,omitempty"`
	Stemming         *Stemming         `xml:"stemming,omitempty"`
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

type OfficielePublicatie struct {
	XMLName                   xml.Name                    `xml:"officiele-publicatie"`
	NoNamespaceSchemaLocation string                      `xml:"noNamespaceSchemaLocation,attr"`
	Metadata                  OfficielePublicatieMetadata `xml:"metadata"`
	Kamerstuk                 *Kamerstuk                  `xml:"kamerstuk"`
}

type OfficielePublicatieMetadata struct {
	Meta []OfficielePublicatieMetaRecord `xml:"meta"`
}

type OfficielePublicatieMetaRecord struct {
	Name    string `xml:"name,attr"`
	Scheme  string `xml:"scheme,attr"`
	Content string `xml:"content,attr"`
}

type Kamerstuk struct {
	Kamerstukkop Kamerstukkop `xml:"kamerstukkop"`
	Dossier      []Dossier    `xml:"dossier"`
	Stuk         Stuk         `xml:"stuk"`
}

type Kamerstukkop struct {
	Tekstregels []Tekstregel `xml:"tekstregel"`
}

type Tekstregel struct {
	Inhoud string `xml:"inhoud,attr"`
	Value  string `xml:",chardata"`
}

type Dossier struct {
	Dossiernummer Dossiernummer `xml:"dossiernummer"`
	Titel         string        `xml:"titel"`
}

type Dossiernummer struct {
	Dossiernr string `xml:"dossiernr"`
}

type Stuk struct {
	Stuknr     Stuknr     `xml:"stuknr"`
	Titel      string     `xml:"titel"`
	Datumtekst Datumtekst `xml:"datumtekst"`
	Algemeen   []Algemeen `xml:"algemeen"`
}

type Stuknr struct {
	Nr          string      `xml:",chardata"`
	Ondernummer Ondernummer `xml:"ondernummer"`
}

type Ondernummer struct {
	Kamer string `xml:"kamer,attr"`
	Value string `xml:",chardata"`
}

type Datumtekst struct {
	Value string `xml:",chardata"`
	Datum Datum  `xml:"datum"`
}

type Datum struct {
	Isodatum string `xml:"isodatum,attr"`
	Value    string `xml:",chardata"`
}

type Algemeen struct {
	VrijeTekst    *VrijeTekst    `xml:"vrije-tekst"`
	TekstSluiting *TekstSluiting `xml:"tekst-sluiting"`
}

type VrijeTekst struct {
	Tekst TekstContent `xml:"tekst"`
}

type TekstContent struct {
	Status  string    `xml:"status,attr"`
	Als     []Al      `xml:"al"`
	AlGroep []AlGroep `xml:"al-groep"`
}

type Al struct {
	Text   string   `xml:",chardata"`
	Nadruk []Nadruk `xml:"nadruk"`
	Noot   []Noot   `xml:"noot"`
	ExtRef []ExtRef `xml:"extref"`
}

type AlGroep struct {
	Als []Al `xml:"al"`
}

type Nadruk struct {
	Type  string `xml:"type,attr"`
	Value string `xml:",chardata"`
}

type Noot struct {
	ID     string `xml:"id,attr"`
	Type   string `xml:"type,attr"`
	NootNr string `xml:"noot.nr"`
	NootAl NootAl `xml:"noot.al"`
}

type NootAl struct {
	Text   string   `xml:",chardata"`
	ExtRef []ExtRef `xml:"extref"`
}

type ExtRef struct {
	Doc    string `xml:"doc,attr"`
	Soort  string `xml:"soort,attr"`
	Status string `xml:"status,attr"`
	Value  string `xml:",chardata"`
}

type TekstSluiting struct {
	Ondertekening Ondertekening `xml:"ondertekening"`
}

type Ondertekening struct {
	Functie string `xml:"functie"`
	Naam    Naam   `xml:"naam"`
}

type Naam struct {
	Voornaam   string `xml:"voornaam"`
	Achternaam string `xml:"achternaam"`
}

// Legacy document structure for kamerwrk documents
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
	Date         *time.Time             `json:"date,omitempty"`
	Dossiers     []ParsedDossier        `json:"dossiers,omitempty"`
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

type ParsedDossier struct {
	Number string `json:"number"`
	Title  string `json:"title"`
}

type PaginationInfo struct {
	CurrentSkiptoken string
	NextSkiptoken    string
	Category         string
	HasMore          bool
}

// used in XML documents
type Reference struct {
	Ref string `xml:"ref,attr"`
}
