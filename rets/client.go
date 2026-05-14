package rets

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"
)

// BaseURL is the root of the Retsinformation REST API.
//
// Endpoint chosen: https://retsinformation-api.dk/v1 — the third-party but
// well-documented mirror with OpenAPI spec, version history, diffs, paragraph
// addressing, markdown/PDF export, and bills (Folketinget) joined in. The
// official harvest endpoint at api.retsinformation.dk only exposes raw XML
// dumps; retsinformation-api.dk gives us structured JSON across the full
// corpus and is the only source for the consolidation/diff features.
//
// Rate limits: 20 req/hour, 50 req/day per IP (see /docs).
const BaseURL = "https://retsinformation-api.dk/v1"

// APIError is returned for non-2xx responses from the Retsinformation API.
type APIError struct {
	StatusCode int
	Body       string
}

func (e *APIError) Error() string {
	return fmt.Sprintf("Retsinformation API error %d: %s", e.StatusCode, e.Body)
}

// Client is a thin HTTP wrapper around retsinformation-api.dk.
type Client struct {
	http    *http.Client
	baseURL string
}

// NewClient constructs a Client with a 30s HTTP timeout, matching dstcli/virkcli.
func NewClient() *Client {
	return &Client{
		http:    &http.Client{Timeout: 30 * time.Second},
		baseURL: BaseURL,
	}
}

// getRaw performs a GET and returns the raw response body, surfacing non-2xx
// status codes as *APIError so callers can format them nicely.
func (c *Client) getRaw(path string, query url.Values) ([]byte, error) {
	u := c.baseURL + path
	if len(query) > 0 {
		u += "?" + query.Encode()
	}
	resp, err := c.http.Get(u)
	if err != nil {
		return nil, fmt.Errorf("GET %s: %w", u, err)
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		return nil, &APIError{StatusCode: resp.StatusCode, Body: string(body)}
	}
	return body, nil
}

// getJSON GETs a JSON endpoint and unmarshals into dst.
func (c *Client) getJSON(path string, query url.Values, dst any) error {
	body, err := c.getRaw(path, query)
	if err != nil {
		return err
	}
	if err := json.Unmarshal(body, dst); err != nil {
		return fmt.Errorf("decoding %s: %w", path, err)
	}
	return nil
}

// SearchOpts holds filters for the /v1/lovgivning/ list endpoint.
type SearchOpts struct {
	Query             string // free-text search in title
	Year              int    // exact year
	Ressort           string // ministry filter (free text matched server-side)
	DocumentType      string // LOV, LOVH, LBK, BEK, CIR, VEJ, KEN, SKR, etc.
	Historical        *bool  // nil = no filter
	IncludeAmendments *bool  // nil = no filter
	Sort              string // signature_date | publication_date | effective_date | title | year | number
	Limit             int
	Skip              int
}

func (o SearchOpts) values() url.Values {
	v := url.Values{}
	if o.Query != "" {
		v.Set("search", o.Query)
	}
	if o.Year != 0 {
		v.Set("year", strconv.Itoa(o.Year))
	}
	if o.Ressort != "" {
		v.Set("ressort", o.Ressort)
	}
	if o.DocumentType != "" {
		v.Set("document_type", strings.ToUpper(o.DocumentType))
	}
	if o.Historical != nil {
		v.Set("historical", strconv.FormatBool(*o.Historical))
	}
	if o.IncludeAmendments != nil {
		v.Set("include_amendments", strconv.FormatBool(*o.IncludeAmendments))
	}
	if o.Sort != "" {
		v.Set("sort", o.Sort)
	}
	if o.Limit > 0 {
		v.Set("limit", strconv.Itoa(o.Limit))
	}
	if o.Skip > 0 {
		v.Set("skip", strconv.Itoa(o.Skip))
	}
	return v
}

// SearchRaw returns the raw JSON body of /v1/lovgivning/.
func (c *Client) SearchRaw(o SearchOpts) ([]byte, error) {
	return c.getRaw("/lovgivning/", o.values())
}

// Search returns the parsed law-version summaries from /v1/lovgivning/.
func (c *Client) Search(o SearchOpts) (*LawList, error) {
	var resp LawList
	if err := c.getJSON("/lovgivning/", o.values(), &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// ResolveRaw returns the raw JSON body of /v1/lovgivning/resolve.
func (c *Client) ResolveRaw(q string) ([]byte, error) {
	v := url.Values{}
	v.Set("q", q)
	return c.getRaw("/lovgivning/resolve", v)
}

// Resolve looks up a law name and returns the canonical (year, number) match.
// The API may return an empty object (no match) — caller should check Number != 0.
func (c *Client) Resolve(q string) (*ResolveMatch, error) {
	v := url.Values{}
	v.Set("q", q)
	var resp ResolveMatch
	if err := c.getJSON("/lovgivning/resolve", v, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetRaw returns the raw JSON body for a specific law version.
func (c *Client) GetRaw(year, number int, include string) ([]byte, error) {
	v := url.Values{}
	if include != "" {
		v.Set("include", include)
	}
	return c.getRaw(fmt.Sprintf("/lovgivning/%d/%d", year, number), v)
}

// Get returns the parsed law with full structure.
func (c *Client) Get(year, number int, include string) (*Law, error) {
	v := url.Values{}
	if include != "" {
		v.Set("include", include)
	}
	var resp Law
	if err := c.getJSON(fmt.Sprintf("/lovgivning/%d/%d", year, number), v, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetMarkdown fetches a law (or a paragraph range) rendered as Markdown.
// paragraphs may be "" (whole law), "1-5", or "10,15,20".
func (c *Client) GetMarkdown(year, number int, paragraphs, exclude string) ([]byte, error) {
	v := url.Values{}
	if paragraphs != "" {
		v.Set("paragraphs", paragraphs)
	}
	if exclude != "" {
		v.Set("exclude", exclude)
	}
	return c.getRaw(fmt.Sprintf("/lovgivning/%d/%d/markdown", year, number), v)
}

// GetParagraph fetches a single paragraph from the base (signed) version of a law.
func (c *Client) GetParagraph(year, number int, paragraph string) (*Paragraph, error) {
	var resp Paragraph
	if err := c.getJSON(fmt.Sprintf("/lovgivning/%d/%d/paragraphs/%s", year, number, url.PathEscape(paragraph)), nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// GetParagraphRaw is the raw-body variant of GetParagraph.
func (c *Client) GetParagraphRaw(year, number int, paragraph string) ([]byte, error) {
	return c.getRaw(fmt.Sprintf("/lovgivning/%d/%d/paragraphs/%s", year, number, url.PathEscape(paragraph)), nil)
}

// HistoryRaw returns the raw unified history across all consolidations.
func (c *Client) HistoryRaw(year, number int, paragraph string) ([]byte, error) {
	v := url.Values{}
	if paragraph != "" {
		v.Set("paragraph", paragraph)
	}
	return c.getRaw(fmt.Sprintf("/lovgivning/%d/%d/history", year, number), v)
}

// History returns the parsed unified history.
func (c *Client) History(year, number int, paragraph string) (*History, error) {
	v := url.Values{}
	if paragraph != "" {
		v.Set("paragraph", paragraph)
	}
	var resp History
	if err := c.getJSON(fmt.Sprintf("/lovgivning/%d/%d/history", year, number), v, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// BillSearchOpts mirrors SearchOpts for /v1/lovgivning/bills/.
type BillSearchOpts struct {
	Query     string
	Status    string
	PeriodeID int
	Enacted   *bool
	Ressort   string
	Limit     int
	Skip      int
}

func (o BillSearchOpts) values() url.Values {
	v := url.Values{}
	if o.Query != "" {
		v.Set("search", o.Query)
	}
	if o.Status != "" {
		v.Set("status", o.Status)
	}
	if o.PeriodeID != 0 {
		v.Set("periode_id", strconv.Itoa(o.PeriodeID))
	}
	if o.Enacted != nil {
		v.Set("enacted", strconv.FormatBool(*o.Enacted))
	}
	if o.Ressort != "" {
		v.Set("ressort", o.Ressort)
	}
	if o.Limit > 0 {
		v.Set("limit", strconv.Itoa(o.Limit))
	}
	if o.Skip > 0 {
		v.Set("skip", strconv.Itoa(o.Skip))
	}
	return v
}

// BillsRaw returns the raw JSON body of /v1/lovgivning/bills/.
func (c *Client) BillsRaw(o BillSearchOpts) ([]byte, error) {
	return c.getRaw("/lovgivning/bills/", o.values())
}

// Bills returns the parsed bills list.
func (c *Client) Bills(o BillSearchOpts) (*BillList, error) {
	var resp BillList
	if err := c.getJSON("/lovgivning/bills/", o.values(), &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// BillRaw fetches a specific bill by its short number (e.g. "L 33").
func (c *Client) BillRaw(number string) ([]byte, error) {
	return c.getRaw("/lovgivning/bills/"+url.PathEscape(number), nil)
}

// Bill fetches a parsed bill.
func (c *Client) Bill(number string) (*Bill, error) {
	var resp Bill
	if err := c.getJSON("/lovgivning/bills/"+url.PathEscape(number), nil, &resp); err != nil {
		return nil, err
	}
	return &resp, nil
}

// QueryRaw exposes any endpoint under /v1/ for power users via `retscli query`.
// path must start with "/" (e.g. "/lovgivning/2018/502/timeline"). query may be nil.
func (c *Client) QueryRaw(path string, query url.Values) ([]byte, error) {
	if !strings.HasPrefix(path, "/") {
		path = "/" + path
	}
	return c.getRaw(path, query)
}
