// Package adapter defines the contract a site must satisfy to be crawled.
package adapter

import (
	"context"
	"net/http"
	"net/url"
)

// Kind labels a target so the adapter can route it during Parse. Values are adapter-defined.
type Kind string

// Target is one unit of work in the frontier.
type Target struct {
	URL   string
	Kind  Kind
	Depth int
	Meta  map[string]string
}

// Record is a typed payload for the store. Data is adapter-owned.
type Record struct {
	Kind string
	ID   string
	Data any
}

// Blob is an adapter-produced file to persist alongside the records.
type Blob struct {
	OwnerID string
	Name    string
	Data    []byte
	Gzip    bool
	// Archive puts the file in the run-wide raw archive rather than the owner directory.
	Archive bool
}

// Result is what parsing one page produced.
type Result struct {
	Records  []Record
	Discover []Target
	Blobs    []Blob
}

// Fetcher is the subset of the engine an adapter may use during Seeds.
type Fetcher interface {
	Get(ctx context.Context, rawURL string) ([]byte, *http.Response, error)
}

// Adapter is the per-site contract. The engine owns scheduling, politeness and storage;
// the adapter owns what to fetch and what a page means.
type Adapter interface {
	Name() string
	Seeds(ctx context.Context, f Fetcher) ([]Target, error)
	// Classify reports the kind of a URL and whether it belongs to this adapter at all.
	Classify(u *url.URL) (Kind, bool)
	// Request builds the HTTP request, letting the adapter set headers such as RSC: 1.
	Request(ctx context.Context, t Target) (*http.Request, error)
	Parse(ctx context.Context, t Target, body []byte, hdr http.Header) (Result, error)
}

// Renderer turns a stored record into a derived artifact. Renderers never touch the
// network, so artifacts can be regenerated offline after a generator change.
type Renderer interface {
	Name() string
	Version() string
	Render(rec Record) ([]byte, error)
}
