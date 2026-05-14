package rets

import "encoding/json"

// LawSummary is one entry in the /v1/lovgivning/ list. It corresponds to a
// single law *version* (a base enactment or a consolidation).
type LawSummary struct {
	ID              string `json:"id"`
	LawID           string `json:"law_id"`
	VersionNumber   int    `json:"version_number"`
	EffectiveDate   string `json:"effective_date"`
	Year            int    `json:"year"`
	Number          int    `json:"number"`
	ELIURI          string `json:"eli_uri"`
	Title           string `json:"title"`
	ShortName       string `json:"short_name"`
	PopularTitle    string `json:"popular_title"`
	Ressort         string `json:"ressort"`
	DocumentType    string `json:"document_type"`
	DocumentID      int    `json:"document_id"`
	Historical      bool   `json:"historical"`
	SignatureDate   string `json:"signature_date"`
	PublicationDate string `json:"publication_date"`
}

// LawList is the top-level response from /v1/lovgivning/.
type LawList struct {
	Data           []LawSummary  `json:"data"`
	Count          int           `json:"count"`
	CanonicalMatch *ResolveMatch `json:"canonical_match,omitempty"`
}

// ResolveMatch is the response from /v1/lovgivning/resolve.
// When no match is found the API returns an empty object; Number == 0 signals that.
type ResolveMatch struct {
	Year         int     `json:"year"`
	Number       int     `json:"number"`
	DocumentType string  `json:"document_type"`
	ShortName    string  `json:"short_name"`
	PopularTitle string  `json:"popular_title"`
	Confidence   float64 `json:"confidence"`
}

// Law is the full response for /v1/lovgivning/{year}/{number}. The structure
// payload is large; we surface metadata fields directly and keep the structure
// for paragraph navigation in `get --paragraph`.
type Law struct {
	ID              string    `json:"id"`
	LawID           string    `json:"law_id"`
	VersionNumber   int       `json:"version_number"`
	EffectiveDate   string    `json:"effective_date"`
	Year            int       `json:"year"`
	Number          int       `json:"number"`
	ELIURI          string    `json:"eli_uri"`
	Title           string    `json:"title"`
	ShortName       string    `json:"short_name"`
	PopularTitle    string    `json:"popular_title"`
	Ressort         string    `json:"ressort"`
	Ministry        string    `json:"ministry"`
	DocumentType    string    `json:"document_type"`
	DocumentID      int       `json:"document_id"`
	Historical      bool      `json:"historical"`
	IsConsolidated  bool      `json:"is_consolidated"`
	SignatureDate   string    `json:"signature_date"`
	PublicationDate string    `json:"publication_date"`
	Structure       Structure `json:"structure"`
}

// Structure is the body of a law as nested chapters, paragraph groups, and paragraphs.
//
// Note: a few fields (signature, appendices) are kept as json.RawMessage
// because the upstream API uses different shapes across documents (sometimes
// []string, sometimes an object). We don't render them in the table view, and
// --json / --raw still expose them.
type Structure struct {
	Title           string           `json:"title"`
	Preamble        []string         `json:"preamble"`
	Chapters        []Chapter        `json:"chapters"`
	ParagraphGroups []ParagraphGroup `json:"paragraph_groups"` // some laws have no chapters
	Signature       json.RawMessage  `json:"signature,omitempty"`
	Appendices      json.RawMessage  `json:"appendices,omitempty"`
}

type Chapter struct {
	ChapterNumber   string           `json:"chapter_number"`
	ChapterTitle    string           `json:"chapter_title"`
	Afsnit          string           `json:"afsnit"`
	ParagraphGroups []ParagraphGroup `json:"paragraph_groups"`
}

type ParagraphGroup struct {
	ID         string      `json:"id"`
	Number     string      `json:"number"`
	Heading    string      `json:"heading"`
	Paragraphs []Paragraph `json:"paragraphs"`
}

// Paragraph is one § with its subsections (stk) and letters (litra).
type Paragraph struct {
	ID     string `json:"id"`
	Number string `json:"number"`
	Stk    []Stk  `json:"stk"`
}

type Stk struct {
	Number string  `json:"number"`
	Text   string  `json:"text"`
	Litra  []Litra `json:"litra"`
}

type Litra struct {
	Letter string `json:"letter"`
	Text   string `json:"text"`
}

type Appendix struct {
	Number string `json:"number"`
	Title  string `json:"title"`
	Text   string `json:"text"`
}

// History is the response of /v1/lovgivning/{year}/{number}/history — the
// consolidation/amendment chain across all versions of a law.
type History struct {
	LawTitle           string         `json:"law_title"`
	TotalEntries       int            `json:"total_entries"`
	Entries            []HistoryEntry `json:"entries"`
	FirstDate          string         `json:"first_date"`
	LastDate           string         `json:"last_date"`
	ConsolidationCount int            `json:"consolidation_count"`
	AmendmentCount     int            `json:"amendment_count"`
	FilteredByParagraf string         `json:"filtered_by_paragraph"`
}

type HistoryEntry struct {
	Sequence      int    `json:"sequence"`
	ValidFrom     string `json:"valid_from"`
	ValidUntil    string `json:"valid_until"`
	Year          int    `json:"year"`
	Number        int    `json:"number"`
	ELIURI        string `json:"eli_uri"`
	EntryType     string `json:"entry_type"` // original | consolidation | amendment
	DocumentType  string `json:"document_type"`
	ShortName     string `json:"short_name"`
	Title         string `json:"title"`
	VersionNumber int    `json:"version_number"`
	IsGap         bool   `json:"is_gap"`
}

// BillSummary is one entry in /v1/lovgivning/bills/.
type BillSummary struct {
	ID            string `json:"id"`
	FTID          int    `json:"ft_id"`
	Number        string `json:"number"`      // e.g. "L 33"
	Title         string `json:"title"`
	ShortTitle    string `json:"short_title"`
	Status        string `json:"status"`
	DecisionDate  string `json:"decision_date"`
	EnactedLawID  string `json:"enacted_law_id"`
}

type BillList struct {
	Data  []BillSummary `json:"data"`
	Count int           `json:"count"`
}

// Bill is the full response for /v1/lovgivning/bills/{number}.
type Bill struct {
	ID            string `json:"id"`
	FTID          int    `json:"ft_id"`
	Number        string `json:"number"`
	Title         string `json:"title"`
	ShortTitle    string `json:"short_title"`
	Status        string `json:"status"`
	DecisionDate  string `json:"decision_date"`
	EnactedLawID  string `json:"enacted_law_id"`
	Resume        string `json:"resume"`
	PeriodeID     int    `json:"periode_id"`
	PeriodeTitle  string `json:"periode_title"`
	Ressort       string `json:"ressort"`
}
