package models

import (
	"time"
)

// Ref: https://opendata.tweedekamer.nl/documentatie/informatiemodel

type Zaak struct {
	ID                  string     `gorm:"primaryKey" json:"id"`
	Nummer              string     `json:"nummer"`
	Onderwerp           string     `json:"onderwerp"`
	Soort               string     `json:"soort"`
	Titel               string     `json:"titel"`
	Citeertitel         *string    `json:"citeertitel"`
	Alias               *string    `json:"alias"`
	Status              string     `json:"status"`
	GestartOp           *time.Time `json:"gestart_op"`
	Organisatie         string     `json:"organisatie"`
	Termijn             time.Time  `json:"termijn"`
	Vergaderjaar        string     `json:"vergaderjaar"`
	Volgnummer          *int       `json:"volgnummer"`
	Afgedaan            bool       `json:"afgedaan"`
	GrootProject        bool       `json:"groot_project"`
	GewijzigdOp         *time.Time `json:"gewijzigd_op"`
	ApiGewijzigdOp      *time.Time `json:"api_gewijzigd_op"`
	Verwijderd          bool       `json:"verwijderd"`
	Kabinetsappreciatie string     `json:"kabinetsappreciatie"`
}

func (Zaak) TableName() string {
	return "zaken"
}

type Besluit struct {
	ID                            string     `gorm:"primaryKey" json:"id"`
	ZaakID                        *string    `gorm:"index" json:"zaak_id"`
	StemmingsSoort                *string    `json:"stemmings_soort"`
	BesluitSoort                  *string    `json:"besluit_soort"`
	BesluitTekst                  *string    `json:"besluit_tekst"`
	Opmerking                     *string    `json:"opmerking"`
	Status                        *string    `json:"status"`
	AgendapuntZaakBesluitVolgorde *int       `json:"agendapunt_zaak_besluit_volgorde"`
	GewijzigdOp                   *time.Time `json:"gewijzigd_op"`
	ApiGewijzigdOp                *time.Time `json:"api_gewijzigd_op"`
	Verwijderd                    bool       `json:"verwijderd"`
}

func (Besluit) TableName() string {
	return "besluiten"
}

type Stemming struct {
	ID              string     `gorm:"primaryKey" json:"id"`
	BesluitID       *string    `gorm:"index" json:"besluit_id"`
	PersoonID       *string    `gorm:"index" json:"persoon_id"`
	FractieID       *string    `gorm:"index" json:"fractie_id"`
	Soort           *string    `json:"soort"`
	FractieGrootte  *int       `json:"fractie_grootte"`
	ActorNaam       *string    `json:"actor_naam"`
	ActorFractie    *string    `json:"actor_fractie"`
	Vergissing      bool       `json:"vergissing"`
	SidActorLid     *string    `json:"sid_actor_lid"`
	SidActorFractie *string    `json:"sid_actor_fractie"`
	GewijzigdOp     *time.Time `json:"gewijzigd_op"`
	ApiGewijzigdOp  *time.Time `json:"api_gewijzigd_op"`
	Verwijderd      bool       `json:"verwijderd"`
}

func (Stemming) TableName() string {
	return "stemmingen"
}

type Persoon struct {
	ID                string     `gorm:"primaryKey" json:"id"`
	Titels            *string    `json:"titels"`
	Initialen         *string    `json:"initialen"`
	Tussenvoegsel     *string    `json:"tussenvoegsel"`
	Achternaam        *string    `json:"achternaam"`
	Voornamen         *string    `json:"voornamen"`
	Roepnaam          *string    `json:"roepnaam"`
	Geslacht          *string    `json:"geslacht"`
	Geboortedatum     *time.Time `json:"geboortedatum"`
	Geboorteplaats    *string    `json:"geboorteplaats"`
	Geboorteland      *string    `json:"geboorteland"`
	Overlijdensdatum  *time.Time `json:"overlijdensdatum"`
	Overlijdensplaats *string    `json:"overlijdensplaats"`
	Overlijdensland   *string    `json:"overlijdensland"`
	Woonplaats        *string    `json:"woonplaats"`
	Land              *string    `json:"land"`
	Bijgewerkt        *time.Time `json:"bijgewerkt"`
	Verwijderd        bool       `json:"verwijderd"`
}

func (Persoon) TableName() string {
	return "personen"
}

type Fractie struct {
	ID             string     `gorm:"primaryKey" json:"id"`
	Nummer         *int       `json:"nummer"`
	Afkorting      *string    `json:"afkorting"`
	NaamNL         *string    `json:"naam_nl"`
	NaamEN         *string    `json:"naam_en"`
	AantalZetels   *int       `json:"aantal_zetels"`
	AantalStemmen  *int       `json:"aantal_stemmen"`
	DatumActief    *time.Time `json:"datum_actief"`
	DatumInactief  *time.Time `json:"datum_inactief"`
	ContentType    *string    `json:"content_type"`
	ContentLength  *int       `json:"content_length"`
	GewijzigdOp    *time.Time `json:"gewijzigd_op"`
	ApiGewijzigdOp *time.Time `json:"api_gewijzigd_op"`
	Verwijderd     bool       `json:"verwijderd"`
}

func (Fractie) TableName() string {
	return "fracties"
}

type ZaakActor struct {
	ID           string     `gorm:"primaryKey" json:"id"`
	ZaakID       *string    `gorm:"index" json:"zaak_id"`
	PersoonID    *string    `gorm:"index" json:"persoon_id"`
	FractieID    *string    `gorm:"index" json:"fractie_id"`
	Relatie      *string    `json:"relatie"`
	ActorNaam    *string    `json:"actor_naam"`
	ActorFractie *string    `json:"actor_fractie"`
	Bijgewerkt   *time.Time `json:"bijgewerkt"`
	Verwijderd   bool       `json:"verwijderd"`
}

func (ZaakActor) TableName() string {
	return "zaak_actors"
}

type KamerstukdossierDB struct {
	ID                string     `gorm:"primaryKey" json:"id"`
	ZaakID            *string    `gorm:"index" json:"zaak_id"`
	Nummer            string     `json:"nummer"`
	Titel             string     `json:"titel"`
	Citeertitel       *string    `json:"citeertitel"`
	Alias             *string    `json:"alias"`
	Toevoeging        *string    `json:"toevoeging"`
	HoogsteVolgnummer int        `json:"hoogste_volgnummer"`
	Afgesloten        bool       `json:"afgesloten"`
	DatumAangemaakt   *time.Time `json:"datum_aangemaakt"`
	DatumGesloten     *time.Time `json:"datum_gesloten"`
	Kamer             string     `json:"kamer"`
	Bijgewerkt        *time.Time `json:"bijgewerkt"`
	ApiGewijzigdOp    *time.Time `json:"api_gewijzigd_op"`
	Verwijderd        bool       `json:"verwijderd"`
}

func (KamerstukdossierDB) TableName() string {
	return "kamerstukdossiers"
}

type DocumentInfo struct {
	ID            uint      `gorm:"primaryKey;autoIncrement"`
	DossierNummer string    `gorm:"index;uniqueIndex:idx_document_unique" json:"dossier_nummer"`
	Volgnummer    int       `gorm:"uniqueIndex:idx_document_unique" json:"volgnummer"`
	URL           string    `json:"url"`
	Content       string    `gorm:"type:jsonb" json:"content"`
	FetchedAt     time.Time `json:"fetched_at"`
	Success       bool      `json:"success"`
	Error         string    `json:"error"`
}

func (DocumentInfo) TableName() string {
	return "document_info"
}

// Junction table for many-to-many relationship between Zaak and DocumentInfo
type ZaakDocument struct {
	ID         uint      `gorm:"primaryKey;autoIncrement"`
	ZaakID     string    `gorm:"index;uniqueIndex:idx_zaak_document" json:"zaak_id"`
	DocumentID uint      `gorm:"index;uniqueIndex:idx_zaak_document" json:"document_id"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`
}

func (ZaakDocument) TableName() string {
	return "zaak_documents"
}

// Backup table for raw JSON data
type RawOData struct {
	ID        uint      `gorm:"primaryKey;autoIncrement"`
	Type      string    `gorm:"index"`
	EntityID  string    `gorm:"index"`
	Data      string    `gorm:"type:jsonb"`
	CreatedAt time.Time `gorm:"autoCreateTime"`
}

func (RawOData) TableName() string {
	return "raw_odata"
}

// Simple aggregate tables
type VotingResult struct {
	ID           uint   `gorm:"primaryKey;autoIncrement"`
	DocumentID   string `gorm:"index"`
	BesluitID    string `gorm:"index"`
	BesluitTekst string `gorm:"type:text"`
	VotingType   string
	PartyVotes   string `gorm:"type:jsonb"`
	Date         *time.Time
	Status       string
}

func (VotingResult) TableName() string {
	return "voting_results"
}

type IndividueleStemming struct {
	ID           uint    `gorm:"primaryKey;autoIncrement"`
	BesluitID    *string `gorm:"index"`
	PersonID     *string `gorm:"index"`
	PersonName   string
	FractieID    *string `gorm:"index"`
	FractieName  string
	VoteType     string
	IsCorrection bool
	Date         *time.Time
}

func (IndividueleStemming) TableName() string {
	return "individuele_stemming"
}
