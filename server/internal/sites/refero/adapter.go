// Package refero adapts styles.refero.design to the crawl engine.
package refero

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
	"regexp"
	"strings"

	"github.com/kanakkholwal/design-supply/internal/adapter"
	"github.com/kanakkholwal/design-supply/internal/flight"
)

const (
	Host    = "styles.refero.design"
	baseURL = "https://" + Host

	KindStyle      adapter.Kind = "style"
	KindCollection adapter.Kind = "collection"
	KindIndex      adapter.Kind = "index"
)

// Options tunes discovery. FollowRelated closes the related-style graph, which is how the
// true inventory gets measured rather than trusted from the sitemap.
type Options struct {
	FollowRelated bool
	Collections   bool
	MaxDepth      int
	OnlyIDs       []string
}

type Adapter struct{ opts Options }

func New(o Options) *Adapter {
	if o.MaxDepth == 0 {
		o.MaxDepth = 6
	}
	return &Adapter{opts: o}
}

func (a *Adapter) Name() string { return "refero" }

var styleIDRe = regexp.MustCompile(`^/style/([0-9a-fA-F-]{36})$`)

func (a *Adapter) Classify(u *url.URL) (adapter.Kind, bool) {
	if u.Host != Host {
		return "", false
	}
	switch {
	case styleIDRe.MatchString(u.Path):
		return KindStyle, true
	case strings.HasPrefix(u.Path, "/design-styles/"),
		strings.HasPrefix(u.Path, "/examples/"):
		return KindCollection, true
	case u.Path == "" || u.Path == "/":
		return KindIndex, true
	}
	return "", false
}

// Request fetches the page as an RSC payload. That is the same URL a browser uses, just
// without the HTML shell: 113 KB instead of 278 KB, and no /api/ path is involved.
func (a *Adapter) Request(ctx context.Context, t adapter.Target) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, t.URL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("RSC", "1")
	req.Header.Set("Accept", "text/x-component, text/html;q=0.9")
	return req, nil
}

// ---- seeds -------------------------------------------------------------------

type sitemapIndex struct {
	Sitemaps []struct {
		Loc string `xml:"loc"`
	} `xml:"sitemap"`
}

type urlset struct {
	URLs []struct {
		Loc     string `xml:"loc"`
		LastMod string `xml:"lastmod"`
	} `xml:"url"`
}

func (a *Adapter) Seeds(ctx context.Context, f adapter.Fetcher) ([]adapter.Target, error) {
	if len(a.opts.OnlyIDs) > 0 {
		out := make([]adapter.Target, 0, len(a.opts.OnlyIDs))
		for _, id := range a.opts.OnlyIDs {
			out = append(out, adapter.Target{URL: StyleURL(id), Kind: KindStyle})
		}
		return out, nil
	}

	body, _, err := f.Get(ctx, baseURL+"/sitemap.xml")
	if err != nil {
		return nil, fmt.Errorf("refero: sitemap index: %w", err)
	}
	var idx sitemapIndex
	if err := xml.Unmarshal(body, &idx); err != nil {
		return nil, fmt.Errorf("refero: parse sitemap index: %w", err)
	}

	seen := map[string]bool{}
	var out []adapter.Target
	for _, sm := range idx.Sitemaps {
		child, _, err := f.Get(ctx, sm.Loc)
		if err != nil {
			return nil, fmt.Errorf("refero: sitemap %s: %w", sm.Loc, err)
		}
		var set urlset
		if err := xml.Unmarshal(child, &set); err != nil {
			return nil, fmt.Errorf("refero: parse %s: %w", sm.Loc, err)
		}
		for _, u := range set.URLs {
			loc := strings.TrimSpace(u.Loc)
			if loc == "" || seen[loc] {
				continue
			}
			parsed, err := url.Parse(loc)
			if err != nil {
				continue
			}
			kind, ok := a.Classify(parsed)
			if !ok {
				continue
			}
			if kind != KindStyle && !a.opts.Collections {
				continue
			}
			seen[loc] = true
			t := adapter.Target{URL: loc, Kind: kind}
			if u.LastMod != "" {
				t.Meta = map[string]string{"lastmod": u.LastMod}
			}
			out = append(out, t)
		}
	}
	return out, nil
}

// StyleURL builds the canonical page URL for a style id.
func StyleURL(id string) string { return baseURL + "/style/" + id }

// StyleID extracts the uuid from a style page URL, or "" if the URL is not one.
func StyleID(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	m := styleIDRe.FindStringSubmatch(u.Path)
	if m == nil {
		return ""
	}
	return strings.ToLower(m[1])
}

// ---- parse -------------------------------------------------------------------

func (a *Adapter) Parse(ctx context.Context, t adapter.Target, body []byte, _ http.Header) (adapter.Result, error) {
	var out adapter.Result
	doc, err := flight.Parse(body)
	if err != nil {
		return out, fmt.Errorf("refero: %s: %w", t.URL, err)
	}

	styleID := StyleID(t.URL)
	page, err := ExtractPage(doc, styleID)
	if err != nil && t.Kind == KindStyle {
		return out, fmt.Errorf("refero: %s: %w", t.URL, err)
	}

	// Cards first: a collection edge or a style row may reference them.
	for _, c := range page.Cards {
		out.Records = append(out.Records, adapter.Record{Kind: "card", ID: c.ID, Data: c})
	}

	if page.Self != nil {
		page.Self.SourceURL = t.URL
		out.Records = append(out.Records, adapter.Record{Kind: "style", ID: page.Self.ID, Data: page.Self})
		// result.json is re-serialised from the untyped tree, not the Go struct, so a
		// field added upstream survives even before the model knows about it.
		if raw, err := RawResultJSON(doc); err == nil {
			out.Blobs = append(out.Blobs, adapter.Blob{
				OwnerID: page.Self.ID, Name: "result.json", Data: raw,
			})
		}
		out.Blobs = append(out.Blobs, adapter.Blob{
			OwnerID: page.Self.ID, Name: page.Self.ID + ".flight.txt",
			Data: body, Gzip: true, Archive: true,
		})
	}

	if t.Kind == KindCollection {
		col := Collection{
			Slug:  collectionSlug(t.URL),
			URL:   t.URL,
			Title: pageTitle(doc),
			Kind:  collectionKind(t.URL),
		}
		for _, c := range page.Cards {
			col.StyleIDs = append(col.StyleIDs, c.ID)
		}
		out.Records = append(out.Records, adapter.Record{Kind: "collection", ID: col.Slug, Data: col})
	}

	if a.opts.FollowRelated && t.Depth < a.opts.MaxDepth {
		for _, c := range page.Cards {
			out.Discover = append(out.Discover, adapter.Target{
				URL:   StyleURL(c.ID),
				Kind:  KindStyle,
				Depth: t.Depth + 1,
			})
		}
	}
	return out, nil
}

func collectionSlug(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return rawURL
	}
	return strings.Trim(u.Path, "/")
}

func collectionKind(rawURL string) string {
	u, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}
	parts := strings.Split(strings.Trim(u.Path, "/"), "/")
	if len(parts) > 1 {
		return parts[0]
	}
	return "root"
}

// pageTitle reads the rendered <title> element out of the flight metadata rows.
func pageTitle(doc *flight.Doc) string {
	for _, id := range doc.Order {
		v, ok := doc.JSON[id]
		if !ok {
			continue
		}
		if title := findTitle(v); title != "" {
			return strings.TrimSuffix(title, " | Refero Styles")
		}
	}
	return ""
}

func findTitle(n any) string {
	switch t := n.(type) {
	case []any:
		if len(t) >= 4 {
			if tag, _ := t[1].(string); tag == "title" {
				if props, ok := t[3].(map[string]any); ok {
					if s, ok := props["children"].(string); ok {
						return s
					}
				}
			}
		}
		for _, e := range t {
			if s := findTitle(e); s != "" {
				return s
			}
		}
	case map[string]any:
		for _, e := range t {
			if s := findTitle(e); s != "" {
				return s
			}
		}
	}
	return ""
}
