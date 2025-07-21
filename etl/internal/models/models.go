package models

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// CustomDate handles various date formats from the API
type CustomDate struct {
	time.Time
	Valid bool
}

func (cd *CustomDate) UnmarshalJSON(data []byte) error {
	str := strings.Trim(string(data), `"`)

	if str == "null" || str == "" {
		cd.Valid = false
		return nil
	}

	formats := []string{
		time.RFC3339,
		"2006-01-02T15:04:05",
		"2006-01-02",
	}

	for _, format := range formats {
		if t, err := time.Parse(format, str); err == nil {
			if t.Year() == 1 && t.Month() == 1 && t.Day() == 1 {
				cd.Valid = false
				return nil
			}
			cd.Time = t
			cd.Valid = true
			return nil
		}
	}

	return fmt.Errorf("unable to parse date: %s", str)
}

func (cd CustomDate) MarshalJSON() ([]byte, error) {
	if !cd.Valid {
		return json.Marshal(nil)
	}
	return json.Marshal(cd.Time.Format(time.RFC3339))
}

func (cd CustomDate) ToTimePtr() *time.Time {
	if !cd.Valid {
		return nil
	}
	return &cd.Time
}

// CustomStringNumber handles fields that can be either string or number
type CustomStringNumber struct {
	Value string
	Valid bool
}

func (csn *CustomStringNumber) UnmarshalJSON(data []byte) error {
	str := strings.Trim(string(data), `"`)

	if str == "null" || str == "" {
		csn.Valid = false
		return nil
	}

	var strValue string
	if err := json.Unmarshal(data, &strValue); err == nil {
		csn.Value = strValue
		csn.Valid = true
		return nil
	}

	var numValue float64
	if err := json.Unmarshal(data, &numValue); err == nil {
		if numValue == float64(int64(numValue)) {
			csn.Value = fmt.Sprintf("%.0f", numValue)
		} else {
			csn.Value = fmt.Sprintf("%g", numValue)
		}
		csn.Valid = true
		return nil
	}

	return fmt.Errorf("unable to parse string/number: %s", str)
}

func (csn CustomStringNumber) MarshalJSON() ([]byte, error) {
	if !csn.Valid {
		return json.Marshal(nil)
	}
	return json.Marshal(csn.Value)
}

func (csn CustomStringNumber) String() string {
	if !csn.Valid {
		return ""
	}
	return csn.Value
}

// Unified models that work for both OData JSON and database storage

type Zaak struct {
	ID                    string      `json:"Id" gorm:"primaryKey;column:id"`
	Nummer                string      `json:"Nummer" gorm:"column:nummer"`
	Onderwerp             string      `json:"Onderwerp" gorm:"column:onderwerp"`
	Soort                 string      `json:"Soort" gorm:"column:soort"`
	Titel                 string      `json:"Titel" gorm:"column:titel"`
	Citeertitel           *string     `json:"Citeertitel" gorm:"column:citeertitel"`
	Alias                 *string     `json:"Alias" gorm:"column:alias"`
	Status                string      `json:"Status" gorm:"column:status"`
	Datum                 *CustomDate `json:"Datum" gorm:"column:datum"`
	GestartOp             time.Time   `json:"GestartOp" gorm:"column:gestart_op"`
	Organisatie           string      `json:"Organisatie" gorm:"column:organisatie"`
	Grondslagvoorhang     *string     `json:"Grondslagvoorhang" gorm:"column:grondslagvoorhang"`
	Termijn               *string     `json:"Termijn" gorm:"column:termijn"`
	Vergaderjaar          string      `json:"Vergaderjaar" gorm:"column:vergaderjaar"`
	Volgnummer            int         `json:"Volgnummer" gorm:"column:volgnummer"`
	HuidigeBehandelstatus *string     `json:"HuidigeBehandelstatus" gorm:"column:huidige_behandelstatus"`
	Afgedaan              bool        `json:"Afgedaan" gorm:"column:afgedaan"`
	GrootProject          bool        `json:"GrootProject" gorm:"column:groot_project"`
	GewijzigdOp           time.Time   `json:"GewijzigdOp" gorm:"column:gewijzigd_op"`
	ApiGewijzigdOp        time.Time   `json:"ApiGewijzigdOp" gorm:"column:api_gewijzigd_op"`
	Verwijderd            bool        `json:"Verwijderd" gorm:"column:verwijderd"`
	Kabinetsappreciatie   string      `json:"Kabinetsappreciatie" gorm:"column:kabinetsappreciatie"`
	DatumAfgedaan         *CustomDate `json:"DatumAfgedaan" gorm:"column:datum_afgedaan"`
	Kamer                 string      `json:"Kamer" gorm:"column:kamer"`

	// Relations - only populated from API, not stored directly
	Besluit          []Besluit          `json:"Besluit,omitempty" gorm:"-"`
	ZaakActor        []ZaakActor        `json:"ZaakActor,omitempty" gorm:"-"`
	Kamerstukdossier []Kamerstukdossier `json:"Kamerstukdossier,omitempty" gorm:"-"`
}

func (Zaak) TableName() string {
	return "zaken"
}

type Besluit struct {
	ID                            string    `json:"Id" gorm:"primaryKey;column:id"`
	AgendapuntId                  string    `json:"Agendapunt_Id" gorm:"column:agendapunt_id"`
	ZaakID                        *string   `json:"zaak_id,omitempty" gorm:"column:zaak_id;index"`
	StemmingsSoort                *string   `json:"StemmingsSoort" gorm:"column:stemmings_soort"`
	BesluitSoort                  string    `json:"BesluitSoort" gorm:"column:besluit_soort"`
	BesluitTekst                  string    `json:"BesluitTekst" gorm:"column:besluit_tekst"`
	Opmerking                     *string   `json:"Opmerking" gorm:"column:opmerking"`
	Status                        string    `json:"Status" gorm:"column:status"`
	AgendapuntZaakBesluitVolgorde int       `json:"AgendapuntZaakBesluitVolgorde" gorm:"column:agendapunt_zaak_besluit_volgorde"`
	GewijzigdOp                   time.Time `json:"GewijzigdOp" gorm:"column:gewijzigd_op"`
	ApiGewijzigdOp                time.Time `json:"ApiGewijzigdOp" gorm:"column:api_gewijzigd_op"`
	Verwijderd                    bool      `json:"Verwijderd" gorm:"column:verwijderd"`

	// Relations
	Stemming []Stemming `json:"Stemming,omitempty" gorm:"-"`
	Zaak     []Zaak     `json:"Zaak,omitempty" gorm:"-"`
}

func (Besluit) TableName() string {
	return "besluiten"
}

type Stemming struct {
	ID              string    `json:"Id" gorm:"primaryKey;column:id"`
	BesluitId       string    `json:"Besluit_Id" gorm:"column:besluit_id_raw"`
	BesluitID       *string   `json:"besluit_id,omitempty" gorm:"column:besluit_id;index"`
	Soort           string    `json:"Soort" gorm:"column:soort"`
	FractieGrootte  int       `json:"FractieGrootte" gorm:"column:fractie_grootte"`
	ActorNaam       string    `json:"ActorNaam" gorm:"column:actor_naam"`
	ActorFractie    string    `json:"ActorFractie" gorm:"column:actor_fractie"`
	Vergissing      bool      `json:"Vergissing" gorm:"column:vergissing"`
	SidActorLid     *string   `json:"SidActorLid" gorm:"column:sid_actor_lid"`
	SidActorFractie string    `json:"SidActorFractie" gorm:"column:sid_actor_fractie"`
	PersoonId       *string   `json:"Persoon_Id" gorm:"column:persoon_id_raw"`
	PersoonID       *string   `json:"persoon_id,omitempty" gorm:"column:persoon_id;index"`
	FractieId       string    `json:"Fractie_Id" gorm:"column:fractie_id_raw"`
	FractieID       *string   `json:"fractie_id,omitempty" gorm:"column:fractie_id;index"`
	GewijzigdOp     time.Time `json:"GewijzigdOp" gorm:"column:gewijzigd_op"`
	ApiGewijzigdOp  time.Time `json:"ApiGewijzigdOp" gorm:"column:api_gewijzigd_op"`
	Verwijderd      bool      `json:"Verwijderd" gorm:"column:verwijderd"`

	// Relations
	Persoon *Persoon `json:"Persoon,omitempty" gorm:"-"`
	Fractie *Fractie `json:"Fractie,omitempty" gorm:"-"`
	Besluit *Besluit `json:"Besluit,omitempty" gorm:"-"`
}

func (Stemming) TableName() string {
	return "stemmingen"
}

type Persoon struct {
	ID                string      `json:"Id" gorm:"primaryKey;column:id"`
	Titels            string      `json:"Titels" gorm:"column:titels"`
	Initialen         string      `json:"Initialen" gorm:"column:initialen"`
	Tussenvoegsel     string      `json:"Tussenvoegsel" gorm:"column:tussenvoegsel"`
	Achternaam        string      `json:"Achternaam" gorm:"column:achternaam"`
	Voornamen         string      `json:"Voornamen" gorm:"column:voornamen"`
	Roepnaam          string      `json:"Roepnaam" gorm:"column:roepnaam"`
	Geslacht          string      `json:"Geslacht" gorm:"column:geslacht"`
	Geboortedatum     *CustomDate `json:"Geboortedatum" gorm:"column:geboortedatum"`
	Geboorteplaats    string      `json:"Geboorteplaats" gorm:"column:geboorteplaats"`
	Geboorteland      string      `json:"Geboorteland" gorm:"column:geboorteland"`
	Overlijdensdatum  *CustomDate `json:"Overlijdensdatum" gorm:"column:overlijdensdatum"`
	Overlijdensplaats string      `json:"Overlijdensplaats" gorm:"column:overlijdensplaats"`
	Overlijdensland   string      `json:"Overlijdensland" gorm:"column:overlijdensland"`
	Woonplaats        string      `json:"Woonplaats" gorm:"column:woonplaats"`
	Land              string      `json:"Land" gorm:"column:land"`
	Bijgewerkt        time.Time   `json:"Bijgewerkt" gorm:"column:bijgewerkt"`
	Verwijderd        bool        `json:"Verwijderd" gorm:"column:verwijderd"`
}

func (Persoon) TableName() string {
	return "personen"
}

type Fractie struct {
	ID             string      `json:"Id" gorm:"primaryKey;column:id"`
	Nummer         int         `json:"Nummer" gorm:"column:nummer"`
	Afkorting      string      `json:"Afkorting" gorm:"column:afkorting"`
	NaamNL         string      `json:"NaamNL" gorm:"column:naam_nl"`
	NaamEN         string      `json:"NaamEN" gorm:"column:naam_en"`
	AantalZetels   int         `json:"AantalZetels" gorm:"column:aantal_zetels"`
	AantalStemmen  int         `json:"AantalStemmen" gorm:"column:aantal_stemmen"`
	DatumActief    *CustomDate `json:"DatumActief" gorm:"column:datum_actief"`
	DatumInactief  *CustomDate `json:"DatumInactief" gorm:"column:datum_inactief"`
	ContentType    string      `json:"ContentType" gorm:"column:content_type"`
	ContentLength  int         `json:"ContentLength" gorm:"column:content_length"`
	GewijzigdOp    time.Time   `json:"GewijzigdOp" gorm:"column:gewijzigd_op"`
	ApiGewijzigdOp time.Time   `json:"ApiGewijzigdOp" gorm:"column:api_gewijzigd_op"`
	Verwijderd     bool        `json:"Verwijderd" gorm:"column:verwijderd"`
}

func (Fractie) TableName() string {
	return "fracties"
}

type ZaakActor struct {
	ID           string    `json:"Id" gorm:"primaryKey;column:id"`
	ZaakID       *string   `json:"zaak_id,omitempty" gorm:"column:zaak_id;index"`
	PersoonID    *string   `json:"persoon_id,omitempty" gorm:"column:persoon_id;index"`
	FractieID    *string   `json:"fractie_id,omitempty" gorm:"column:fractie_id;index"`
	Relatie      string    `json:"Relatie" gorm:"column:relatie"`
	ActorNaam    string    `json:"ActorNaam" gorm:"column:actor_naam"`
	ActorFractie string    `json:"ActorFractie" gorm:"column:actor_fractie"`
	Bijgewerkt   time.Time `json:"Bijgewerkt" gorm:"column:bijgewerkt"`
	Verwijderd   bool      `json:"Verwijderd" gorm:"column:verwijderd"`

	// Relations
	Persoon *Persoon `json:"Persoon,omitempty" gorm:"-"`
	Fractie *Fractie `json:"Fractie,omitempty" gorm:"-"`
	Zaak    *Zaak    `json:"Zaak,omitempty" gorm:"-"`
}

func (ZaakActor) TableName() string {
	return "zaak_actors"
}

type Kamerstukdossier struct {
	ID                string             `json:"Id" gorm:"primaryKey;column:id"`
	ZaakID            *string            `json:"zaak_id,omitempty" gorm:"column:zaak_id;index"`
	Nummer            CustomStringNumber `json:"Nummer" gorm:"column:nummer"`
	Titel             string             `json:"Titel" gorm:"column:titel"`
	Citeertitel       *string            `json:"Citeertitel" gorm:"column:citeertitel"`
	Alias             *string            `json:"Alias" gorm:"column:alias"`
	Toevoeging        *string            `json:"Toevoeging" gorm:"column:toevoeging"`
	HoogsteVolgnummer int                `json:"HoogsteVolgnummer" gorm:"column:hoogste_volgnummer"`
	Afgesloten        bool               `json:"Afgesloten" gorm:"column:afgesloten"`
	DatumAangemaakt   *CustomDate        `json:"DatumAangemaakt" gorm:"column:datum_aangemaakt"`
	DatumGesloten     *CustomDate        `json:"DatumGesloten" gorm:"column:datum_gesloten"`
	Kamer             string             `json:"Kamer" gorm:"column:kamer"`
	Bijgewerkt        time.Time          `json:"Bijgewerkt" gorm:"column:bijgewerkt"`
	ApiGewijzigdOp    time.Time          `json:"ApiGewijzigdOp" gorm:"column:api_gewijzigd_op"`
	Verwijderd        bool               `json:"Verwijderd" gorm:"column:verwijderd"`
}

func (Kamerstukdossier) TableName() string {
	return "kamerstukdossiers"
}

// Support types
type DocumentInfo struct {
	ID            uint                   `gorm:"primaryKey;autoIncrement"`
	DossierNummer string                 `gorm:"index;uniqueIndex:idx_document_unique" json:"dossier_nummer"`
	Volgnummer    int                    `gorm:"uniqueIndex:idx_document_unique" json:"volgnummer"`
	URL           string                 `json:"url"`
	Content       map[string]interface{} `gorm:"type:jsonb" json:"content"`
	FetchedAt     time.Time              `json:"fetched_at"`
	Success       bool                   `json:"success"`
	Error         string                 `json:"error"`
}

func (DocumentInfo) TableName() string {
	return "document_info"
}

type ZaakDocument struct {
	ID         uint      `gorm:"primaryKey;autoIncrement"`
	ZaakID     string    `gorm:"index;uniqueIndex:idx_zaak_document" json:"zaak_id"`
	DocumentID uint      `gorm:"index;uniqueIndex:idx_zaak_document" json:"document_id"`
	CreatedAt  time.Time `gorm:"autoCreateTime"`
}

func (ZaakDocument) TableName() string {
	return "zaak_documents"
}

// ImportStats tracks import progress
type ImportStats struct {
	TotalZaken             int            `json:"total_zaken"`
	TotalBesluiten         int            `json:"total_besluiten"`
	TotalStemmingen        int            `json:"total_stemmingen"`
	TotalPersonen          int            `json:"total_personen"`
	TotalFracties          int            `json:"total_fracties"`
	TotalKamerstukdossiers int            `json:"total_kamerstukdossiers"`
	ZakenByType            map[string]int `json:"zaken_by_type"`
	ProcessingErrors       int            `json:"processing_errors"`
	ErrorDetails           []string       `json:"error_details"`
	ProcessingStartTime    time.Time      `json:"processing_start_time"`
	ProcessingEndTime      time.Time      `json:"processing_end_time"`
	ProcessingDuration     time.Duration  `json:"processing_duration"`
}

func NewImportStats() *ImportStats {
	return &ImportStats{
		ZakenByType:         make(map[string]int),
		ErrorDetails:        make([]string, 0),
		ProcessingStartTime: time.Now(),
	}
}

func (s *ImportStats) IncrementZaakType(soort string) {
	s.ZakenByType[soort]++
	s.TotalZaken++
}

func (s *ImportStats) AddError(err string) {
	s.ProcessingErrors++
	s.ErrorDetails = append(s.ErrorDetails, err)
}

func (s *ImportStats) Finalize() {
	s.ProcessingEndTime = time.Now()
	s.ProcessingDuration = s.ProcessingEndTime.Sub(s.ProcessingStartTime)
}
