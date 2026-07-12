package ingest

import (
	"time"

	"partijgedrag/internal/source/tweedekamer"
)

type rawRecordProjection struct {
	Collection      string
	SourceID        string
	SourceUpdatedAt *time.Time
	SourceDeleted   bool
	Payload         []byte
	PayloadHash     string
}

type motionProjection struct {
	MotionKey         string
	SourceKey         string
	JurisdictionKey   string
	SourceID          string
	Number            *string
	Title             *string
	Subject           *string
	Status            *string
	Kind              *string
	ParliamentaryYear *string
	ProposedAt        *time.Time
	SourceUpdatedAt   *time.Time
	SourceDeleted     bool
	RawCollection     string
}

type partyProjection struct {
	PartyKey        string
	SourceKey       string
	JurisdictionKey string
	SourceID        string
	Number          *int
	ShortName       *string
	Name            *string
	NameEN          *string
	Seats           *int
	ElectoralVotes  *int
	ActiveFrom      *time.Time
	ActiveTo        *time.Time
	SourceUpdatedAt *time.Time
	SourceDeleted   bool
	RawCollection   string
}

type decisionProjection struct {
	DecisionKey         string
	SourceKey           string
	MotionKey           string
	SourceID            string
	AgendaPointSourceID *string
	VotingType          *string
	DecisionType        *string
	DecisionText        *string
	Comment             *string
	Status              *string
	DecisionOrder       *int
	SourceUpdatedAt     *time.Time
	SourceDeleted       bool
}

type voteProjection struct {
	VoteKey         string
	SourceKey       string
	MotionKey       string
	DecisionKey     string
	SourceID        string
	VoteType        *string
	PartySourceID   *string
	PartyName       *string
	ActorName       *string
	PartySize       *int
	Mistake         bool
	PersonSourceID  *string
	SourceUpdatedAt *time.Time
	SourceDeleted   bool
}

func projectMotionRaw(record tweedekamer.MotionRecord) rawRecordProjection {
	return projectRawRecord(zaakCollection, record.ID, record.ApiGewijzigdOp, record.Verwijderd, record.Raw)
}

func projectMotion(jurisdictionKey string, record tweedekamer.MotionRecord) motionProjection {
	return motionProjection{
		MotionKey:         motionKey(record.ID),
		SourceKey:         tweedeKamerSourceKey,
		JurisdictionKey:   jurisdictionKey,
		SourceID:          record.ID,
		Number:            record.Nummer,
		Title:             title(record),
		Subject:           record.Onderwerp,
		Status:            record.Status,
		Kind:              record.Soort,
		ParliamentaryYear: record.Vergaderjaar,
		ProposedAt:        proposedAt(record),
		SourceUpdatedAt:   timePtr(record.ApiGewijzigdOp),
		SourceDeleted:     boolValue(record.Verwijderd),
		RawCollection:     zaakCollection,
	}
}

func projectPartyRaw(record tweedekamer.PartyRecord) rawRecordProjection {
	return projectRawRecord(fractieCollection, record.ID, record.ApiGewijzigdOp, record.Verwijderd, record.Raw)
}

func projectParty(jurisdictionKey string, record tweedekamer.PartyRecord) partyProjection {
	return partyProjection{
		PartyKey:        partyKey(record.ID),
		SourceKey:       tweedeKamerSourceKey,
		JurisdictionKey: jurisdictionKey,
		SourceID:        record.ID,
		Number:          record.Nummer,
		ShortName:       record.Afkorting,
		Name:            record.NaamNL,
		NameEN:          record.NaamEN,
		Seats:           record.AantalZetels,
		ElectoralVotes:  record.AantalStemmen,
		ActiveFrom:      timePtr(record.DatumActief),
		ActiveTo:        timePtr(record.DatumInactief),
		SourceUpdatedAt: timePtr(record.ApiGewijzigdOp),
		SourceDeleted:   boolValue(record.Verwijderd),
		RawCollection:   fractieCollection,
	}
}

func projectDecisionRaw(record tweedekamer.DecisionRecord) rawRecordProjection {
	return projectRawRecord(besluitCollection, record.ID, record.ApiGewijzigdOp, record.Verwijderd, record.Raw)
}

func projectDecision(motion motionCandidate, record tweedekamer.DecisionRecord) decisionProjection {
	return decisionProjection{
		DecisionKey:         decisionKey(record.ID),
		SourceKey:           tweedeKamerSourceKey,
		MotionKey:           motion.MotionKey,
		SourceID:            record.ID,
		AgendaPointSourceID: record.AgendapuntID,
		VotingType:          record.StemmingsSoort,
		DecisionType:        record.BesluitSoort,
		DecisionText:        record.BesluitTekst,
		Comment:             record.Opmerking,
		Status:              record.Status,
		DecisionOrder:       record.AgendapuntZaakBesluitVolgorde,
		SourceUpdatedAt:     timePtr(record.ApiGewijzigdOp),
		SourceDeleted:       boolValue(record.Verwijderd),
	}
}

func projectVoteRaw(record tweedekamer.VoteRecord) rawRecordProjection {
	return projectRawRecord(stemmingCollection, record.ID, record.ApiGewijzigdOp, record.Verwijderd, record.Raw)
}

func projectVote(motion motionCandidate, decision tweedekamer.DecisionRecord, record tweedekamer.VoteRecord) voteProjection {
	return voteProjection{
		VoteKey:         voteKey(record.ID),
		SourceKey:       tweedeKamerSourceKey,
		MotionKey:       motion.MotionKey,
		DecisionKey:     decisionKey(decision.ID),
		SourceID:        record.ID,
		VoteType:        record.Soort,
		PartySourceID:   record.FractieID,
		PartyName:       record.ActorFractie,
		ActorName:       record.ActorNaam,
		PartySize:       record.FractieGrootte,
		Mistake:         boolValue(record.Vergissing),
		PersonSourceID:  record.PersoonID,
		SourceUpdatedAt: timePtr(record.ApiGewijzigdOp),
		SourceDeleted:   boolValue(record.Verwijderd),
	}
}

func projectRawRecord(collection string, sourceID string, sourceUpdatedAt *tweedekamer.Time, sourceDeleted *bool, payload []byte) rawRecordProjection {
	return rawRecordProjection{
		Collection:      collection,
		SourceID:        sourceID,
		SourceUpdatedAt: timePtr(sourceUpdatedAt),
		SourceDeleted:   boolValue(sourceDeleted),
		Payload:         payload,
		PayloadHash:     hashBytes(payload),
	}
}

func boolValue(value *bool) bool {
	return value != nil && *value
}

func partyKey(sourceID string) string {
	return tweedeKamerSourceKey + ":party:" + sourceID
}
