package odata

/*
 * Source: https://opendata.tweedekamer.nl/documentatie/informatiemodel
 */

import (
	"encoding/json"
	"fmt"
	"strings"
	"time"
)

// date parsing for both RFC3339 and YYYY-MM-DD formats
type CustomDate struct {
	time.Time
	Valid bool
}

// custom JSON unmarshaling for dates
func (cd *CustomDate) UnmarshalJSON(data []byte) error {
	// Remove quotes from the JSON string
	str := strings.Trim(string(data), `"`)

	if str == "null" || str == "" {
		cd.Valid = false
		return nil
	}

	// Try RFC3339 format first (with time)
	if t, err := time.Parse(time.RFC3339, str); err == nil {
		// Check if this is a zero date (0001-01-01)
		if t.Year() == 1 && t.Month() == 1 && t.Day() == 1 {
			cd.Valid = false
			return nil
		}
		cd.Time = t
		cd.Valid = true
		return nil
	}

	// Try YYYY-MM-DD format
	if t, err := time.Parse("2006-01-02", str); err == nil {
		// Check if this is a zero date (0001-01-01)
		if t.Year() == 1 && t.Month() == 1 && t.Day() == 1 {
			cd.Valid = false
			return nil
		}
		cd.Time = t
		cd.Valid = true
		return nil
	}

	// Try YYYY-MM-DDTHH:MM:SS format (without timezone)
	if t, err := time.Parse("2006-01-02T15:04:05", str); err == nil {
		// Check if this is a zero date (0001-01-01)
		if t.Year() == 1 && t.Month() == 1 && t.Day() == 1 {
			cd.Valid = false
			return nil
		}
		cd.Time = t
		cd.Valid = true
		return nil
	}

	return fmt.Errorf("unable to parse date: %s", str)
}

// custom JSON marshaling for dates
func (cd CustomDate) MarshalJSON() ([]byte, error) {
	if !cd.Valid {
		return json.Marshal(nil)
	}
	return json.Marshal(cd.Time.Format(time.RFC3339))
}

// pointer to the time.Time value if valid, otherwise nil
func (cd CustomDate) ToTimePtr() *time.Time {
	if !cd.Valid {
		return nil
	}
	return &cd.Time
}

// a case/motion
// usually it's one of these three proposals:
// - Law addition
// - Law change
// - Individuele stemming
type Zaak struct {
	ID                    string      `json:"Id"`
	Nummer                string      `json:"Nummer"`
	Onderwerp             string      `json:"Onderwerp"`
	Soort                 string      `json:"Soort"`
	Titel                 string      `json:"Titel"`
	Citeertitel           *string     `json:"Citeertitel"`
	Alias                 *string     `json:"Alias"`
	Status                string      `json:"Status"`
	Datum                 *CustomDate `json:"Datum"`
	GestartOp             time.Time   `json:"GestartOp"`
	Organisatie           string      `json:"Organisatie"`
	Grondslagvoorhang     *string     `json:"Grondslagvoorhang"`
	Termijn               *string     `json:"Termijn"`
	Vergaderjaar          string      `json:"Vergaderjaar"`
	Volgnummer            int         `json:"Volgnummer"`
	HuidigeBehandelstatus *string     `json:"HuidigeBehandelstatus"`
	Afgedaan              bool        `json:"Afgedaan"`
	GrootProject          bool        `json:"GrootProject"`
	GewijzigdOp           time.Time   `json:"GewijzigdOp"`
	ApiGewijzigdOp        time.Time   `json:"ApiGewijzigdOp"`
	Verwijderd            bool        `json:"Verwijderd"`
	Kabinetsappreciatie   string      `json:"Kabinetsappreciatie"`
	DatumAfgedaan         *CustomDate `json:"DatumAfgedaan"`
	Kamer                 string      `json:"Kamer"`
	Besluit               []Besluit   `json:"Besluit,omitempty"`
	ZaakActor             []ZaakActor `json:"ZaakActor,omitempty"`
}

type Besluit struct {
	ID                            string     `json:"Id"`
	AgendapuntId                  string     `json:"Agendapunt_Id"`
	StemmingsSoort                *string    `json:"StemmingsSoort"`
	BesluitSoort                  string     `json:"BesluitSoort"`
	BesluitTekst                  string     `json:"BesluitTekst"`
	Opmerking                     *string    `json:"Opmerking"`
	Status                        string     `json:"Status"`
	AgendapuntZaakBesluitVolgorde int        `json:"AgendapuntZaakBesluitVolgorde"`
	GewijzigdOp                   time.Time  `json:"GewijzigdOp"`
	ApiGewijzigdOp                time.Time  `json:"ApiGewijzigdOp"`
	Verwijderd                    bool       `json:"Verwijderd"`
	Stemming                      []Stemming `json:"Stemming,omitempty"`
	Zaak                          []Zaak     `json:"Zaak,omitempty"`
}

type Stemming struct {
	ID              string    `json:"Id"`
	BesluitId       string    `json:"Besluit_Id"`
	Soort           string    `json:"Soort"`
	FractieGrootte  int       `json:"FractieGrootte"`
	ActorNaam       string    `json:"ActorNaam"`
	ActorFractie    string    `json:"ActorFractie"`
	Vergissing      bool      `json:"Vergissing"`
	SidActorLid     *string   `json:"SidActorLid"`
	SidActorFractie string    `json:"SidActorFractie"`
	PersoonId       *string   `json:"Persoon_Id"`
	FractieId       string    `json:"Fractie_Id"`
	GewijzigdOp     time.Time `json:"GewijzigdOp"`
	ApiGewijzigdOp  time.Time `json:"ApiGewijzigdOp"`
	Verwijderd      bool      `json:"Verwijderd"`
	Persoon         *Persoon  `json:"Persoon,omitempty"`
	Fractie         *Fractie  `json:"Fractie,omitempty"`
	Besluit         *Besluit  `json:"Besluit,omitempty"`
}

type Persoon struct {
	ID                string      `json:"Id"`
	Titels            string      `json:"Titels"`
	Initialen         string      `json:"Initialen"`
	Tussenvoegsel     string      `json:"Tussenvoegsel"`
	Achternaam        string      `json:"Achternaam"`
	Voornamen         string      `json:"Voornamen"`
	Roepnaam          string      `json:"Roepnaam"`
	Geslacht          string      `json:"Geslacht"`
	Geboortedatum     *CustomDate `json:"Geboortedatum"`
	Geboorteplaats    string      `json:"Geboorteplaats"`
	Geboorteland      string      `json:"Geboorteland"`
	Overlijdensdatum  *CustomDate `json:"Overlijdensdatum"`
	Overlijdensplaats string      `json:"Overlijdensplaats"`
	Overlijdensland   string      `json:"Overlijdensland"`
	Woonplaats        string      `json:"Woonplaats"`
	Land              string      `json:"Land"`
	Bijgewerkt        time.Time   `json:"Bijgewerkt"`
	Verwijderd        bool        `json:"Verwijderd"`
}

// a political party/faction
type Fractie struct {
	ID             string      `json:"Id"`
	Nummer         int         `json:"Nummer"`
	Afkorting      string      `json:"Afkorting"`
	NaamNL         string      `json:"NaamNL"`
	NaamEN         string      `json:"NaamEN"`
	AantalZetels   int         `json:"AantalZetels"`
	AantalStemmen  int         `json:"AantalStemmen"`
	DatumActief    *CustomDate `json:"DatumActief"`
	DatumInactief  *CustomDate `json:"DatumInactief"`
	ContentType    string      `json:"ContentType"`
	ContentLength  int         `json:"ContentLength"`
	GewijzigdOp    time.Time   `json:"GewijzigdOp"`
	ApiGewijzigdOp time.Time   `json:"ApiGewijzigdOp"`
	Verwijderd     bool        `json:"Verwijderd"`
}

// the relationship between a case and an actor (person)
type ZaakActor struct {
	ID           string    `json:"Id"`
	Relatie      string    `json:"Relatie"`
	ActorNaam    string    `json:"ActorNaam"`
	ActorFractie string    `json:"ActorFractie"`
	Bijgewerkt   time.Time `json:"Bijgewerkt"`
	Verwijderd   bool      `json:"Verwijderd"`
	Persoon      *Persoon  `json:"Persoon,omitempty"`
	Fractie      *Fractie  `json:"Fractie,omitempty"`
	Zaak         *Zaak     `json:"Zaak,omitempty"`
}

type Agendapunt struct {
	ID           string      `json:"Id"`
	Nummer       string      `json:"Nummer"`
	Onderwerp    string      `json:"Onderwerp"`
	Titel        string      `json:"Titel"`
	Omschrijving string      `json:"Omschrijving"`
	Volgorde     int         `json:"Volgorde"`
	Rubriek      string      `json:"Rubriek"`
	Datum        *CustomDate `json:"Datum"`
	Aanvangstijd string      `json:"Aanvangstijd"`
	Eindtijd     string      `json:"Eindtijd"`
	Bijgewerkt   time.Time   `json:"Bijgewerkt"`
	Verwijderd   bool        `json:"Verwijderd"`
	Besluit      []Besluit   `json:"Besluit,omitempty"`
}

type VotingResult struct {
	ZaakID       string            `json:"zaak_id"`
	ZaakNummer   string            `json:"zaak_nummer"`
	ZaakTitel    string            `json:"zaak_titel"`
	ZaakSoort    string            `json:"zaak_soort"`
	BesluitID    string            `json:"besluit_id"`
	BesluitTekst string            `json:"besluit_tekst"`
	BesluitSoort string            `json:"besluit_soort"`
	VotingType   string            `json:"voting_type"`
	PartyVotes   map[string]string `json:"party_votes"` // party -> "voor"/"tegen"/"niet deelgenomen"
	Date         time.Time         `json:"date"`
	Status       string            `json:"status"`
}

// individual vote record
type IndividueleStemming struct {
	ZaakID       string     `json:"zaak_id"`
	ZaakNummer   string     `json:"zaak_nummer"`
	ZaakTitel    string     `json:"zaak_titel"`
	BesluitID    string     `json:"besluit_id"`
	BesluitTekst string     `json:"besluit_tekst"`
	PersonID     string     `json:"person_id"`
	PersonName   string     `json:"person_name"`
	FractieID    string     `json:"fractie_id"`
	FractieName  string     `json:"fractie_name"`
	VoteType     string     `json:"vote_type"` // "voor"/"tegen"/"niet deelgenomen"
	IsCorrection bool       `json:"is_correction"`
	Date         *time.Time `json:"date"`
}

type MotionSubmitter struct {
	ZaakID      string     `json:"zaak_id"`
	ZaakNummer  string     `json:"zaak_nummer"`
	ZaakTitel   string     `json:"zaak_titel"`
	PersonID    string     `json:"person_id"`
	PersonName  string     `json:"person_name"`
	FractieID   string     `json:"fractie_id"`
	FractieName string     `json:"fractie_name"`
	Role        string     `json:"role"`
	Date        *time.Time `json:"date"`
}

type BatchResponse struct {
	TotalCount     int         `json:"total_count"`
	ProcessedCount int         `json:"processed_count"`
	HasMore        bool        `json:"has_more"`
	NextSkip       int         `json:"next_skip"`
	Data           interface{} `json:"data"`
}

// statistics about the import process
type ImportStats struct {
	TotalZaken          int            `json:"total_zaken"`
	TotalBesluiten      int            `json:"total_besluiten"`
	TotalStemmingen     int            `json:"total_stemmingen"`
	TotalPersonen       int            `json:"total_personen"`
	TotalFracties       int            `json:"total_fracties"`
	ZakenByType         map[string]int `json:"zaken_by_type"`
	BesluitenByType     map[string]int `json:"besluiten_by_type"`
	StemmingByType      map[string]int `json:"stemming_by_type"`
	ProcessingErrors    int            `json:"processing_errors"`
	ErrorDetails        []string       `json:"error_details"`
	ProcessingStartTime time.Time      `json:"processing_start_time"`
	ProcessingEndTime   time.Time      `json:"processing_end_time"`
	ProcessingDuration  time.Duration  `json:"processing_duration"`
}

func NewImportStats() *ImportStats {
	return &ImportStats{
		ZakenByType:         make(map[string]int),
		BesluitenByType:     make(map[string]int),
		StemmingByType:      make(map[string]int),
		ErrorDetails:        make([]string, 0),
		ProcessingStartTime: time.Now(),
	}
}

func (s *ImportStats) AddError(error string) {
	s.ProcessingErrors++
	s.ErrorDetails = append(s.ErrorDetails, error)
}

func (s *ImportStats) IncrementZaakType(soort string) {
	s.ZakenByType[soort]++
	s.TotalZaken++
}

// finalize the statistics
func (s *ImportStats) Finalize() {
	s.ProcessingEndTime = time.Now()
	s.ProcessingDuration = s.ProcessingEndTime.Sub(s.ProcessingStartTime)
}
