// Package store composes the SQLite index and the on-disk archive into the single
// engine.Store the crawler writes through.
package store

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"github.com/kanakkholwal/design-supply/internal/adapter"
	"github.com/kanakkholwal/design-supply/internal/engine"
	"github.com/kanakkholwal/design-supply/internal/sites/refero"
	"github.com/kanakkholwal/design-supply/internal/sites/refero/render"
	"github.com/kanakkholwal/design-supply/internal/store/fsstore"
	"github.com/kanakkholwal/design-supply/internal/store/sqlite"
)

type Store struct {
	DB *sqlite.DB
	FS *fsstore.FS
}

func Open(dbPath, dataRoot string) (*Store, error) {
	fs, err := fsstore.New(dataRoot)
	if err != nil {
		return nil, err
	}
	db, err := sqlite.Open(dbPath)
	if err != nil {
		return nil, err
	}
	return &Store{DB: db, FS: fs}, nil
}

func (s *Store) Close() error { return s.DB.Close() }

func (s *Store) Push(ctx context.Context, ts []adapter.Target) (int, error) {
	return s.DB.Push(ctx, ts)
}

func (s *Store) Lease(ctx context.Context, n int) ([]adapter.Target, error) {
	return s.DB.Lease(ctx, n)
}

func (s *Store) Complete(ctx context.Context, rawURL string, meta engine.FetchMeta) error {
	return s.DB.Complete(ctx, rawURL, meta)
}

func (s *Store) Release(ctx context.Context, rawURL string) error {
	return s.DB.Release(ctx, rawURL)
}

func (s *Store) Pending(ctx context.Context) (int, error) { return s.DB.Pending(ctx) }

func (s *Store) Validators(ctx context.Context, rawURL string) (string, string, bool) {
	return s.DB.Validators(ctx, rawURL)
}

// Write commits to SQLite first, then lays the files down. A failed file write leaves the
// index row in place; a re-run rewrites the file.
func (s *Store) Write(ctx context.Context, t adapter.Target, res adapter.Result) error {
	if err := s.DB.Write(ctx, t, res); err != nil {
		return err
	}

	origins := map[string]string{}
	for _, rec := range res.Records {
		st, ok := rec.Data.(*refero.Style)
		if !ok {
			continue
		}
		origins[st.ID] = st.Origin
		body, err := json.MarshalIndent(styleFileFrom(st), "", "  ")
		if err != nil {
			return err
		}
		if _, err := s.FS.WriteFile(s.FS.StyleDir(st.Origin, st.ID), "style.json", body, false); err != nil {
			return err
		}
	}

	rawResults := map[string][]byte{}
	for _, b := range res.Blobs {
		if !b.Archive && b.Name == "result.json" {
			rawResults[b.OwnerID] = b.Data
		}
		dir := s.FS.RawDir()
		if !b.Archive {
			origin, ok := origins[b.OwnerID]
			if !ok {
				var err error
				if origin, err = s.originOf(ctx, b.OwnerID); err != nil {
					return fmt.Errorf("store: origin for %s: %w", b.OwnerID, err)
				}
			}
			dir = s.FS.StyleDir(origin, b.OwnerID)
		}
		if _, err := s.FS.WriteFile(dir, b.Name, b.Data, b.Gzip); err != nil {
			return err
		}
	}

	for id, raw := range rawResults {
		if err := s.RenderStyle(ctx, id, origins[id], raw); err != nil {
			return fmt.Errorf("store: render %s: %w", id, err)
		}
	}
	return nil
}

// RenderStyle writes every export format for one style and indexes the results. It never
// touches the network, so it can be replayed offline after a generator change.
func (s *Store) RenderStyle(ctx context.Context, styleID, origin string, rawResult []byte) error {
	if origin == "" {
		var err error
		if origin, err = s.originOf(ctx, styleID); err != nil {
			return err
		}
	}
	artifacts, err := render.FromRawResult(rawResult)
	if err != nil {
		return err
	}
	dir := s.FS.StyleDir(origin, styleID)
	for _, a := range artifacts {
		path, err := s.FS.WriteFile(dir, a.Name, a.Data, false)
		if err != nil {
			return err
		}
		sum := sha256.Sum256(a.Data)
		if err := s.DB.RecordArtifact(ctx, styleID, a.Name, render.Version,
			hex.EncodeToString(sum[:]), s.FS.Rel(path), int64(len(a.Data))); err != nil {
			return err
		}
	}
	return nil
}

func (s *Store) originOf(ctx context.Context, styleID string) (string, error) {
	var origin string
	err := s.DB.SQL().QueryRowContext(ctx, `
		SELECT COALESCE(si.origin, '') FROM styles st
		LEFT JOIN sites si ON si.id = st.site_id WHERE st.id = ?`, styleID).Scan(&origin)
	return origin, err
}

// styleFile is the human-facing summary written next to result.json. Every media URL is
// present here whether or not its bytes were ever downloaded.
type styleFile struct {
	ID        string       `json:"id"`
	SiteName  string       `json:"siteName"`
	URL       string       `json:"url"`
	Origin    string       `json:"origin"`
	ETLD1     string       `json:"etld1"`
	Subdomain string       `json:"subdomain,omitempty"`
	SourceURL string       `json:"sourceUrl"`
	Theme     string       `json:"theme,omitempty"`
	Industry  string       `json:"industry,omitempty"`
	NorthStar string       `json:"northStar,omitempty"`
	Extracted string       `json:"extractedAt,omitempty"`
	ResultSHA string       `json:"resultSha256"`
	Media     []mediaEntry `json:"media"`
}

type mediaEntry struct {
	Role       string `json:"role"`
	Kind       string `json:"kind"`
	URL        string `json:"url"`
	Width      int    `json:"width,omitempty"`
	Height     int    `json:"height,omitempty"`
	DurationMs int    `json:"durationMs,omitempty"`
	Downloaded bool   `json:"downloaded"`
}

func styleFileFrom(st *refero.Style) styleFile {
	c := st.Card
	entries := []mediaEntry{
		{"screenshot", "screenshot", c.ScreenshotURL, 0, 0, 0, false},
		{"thumbnail", "thumbnail", c.ThumbnailURL, 0, 0, 0, false},
		{"icon", "icon", c.IconURL, 0, 0, 0, false},
		{"preview_video", "video", c.PreviewVideoURL, c.PreviewVideoWidth, c.PreviewVideoHeight, c.PreviewVideoDurationMs, false},
		{"preview_video_poster", "video_poster", c.PreviewVideoPosterURL, c.PreviewVideoWidth, c.PreviewVideoHeight, 0, false},
		{"preview_video_detail", "video", c.PreviewVideoDetailURL, c.PreviewVideoDetailWidth, c.PreviewVideoDetailHeight, 0, false},
		{"preview_video_detail_poster", "video_poster", c.PreviewVideoDetailPosterURL, c.PreviewVideoDetailWidth, c.PreviewVideoDetailHeight, 0, false},
	}
	media := make([]mediaEntry, 0, len(entries))
	for _, e := range entries {
		if e.URL != "" {
			media = append(media, e)
		}
	}
	extracted := ""
	if ts, ok := st.Result.ExtractedAt(); ok {
		extracted = ts.Format("2006-01-02T15:04:05Z")
	}
	return styleFile{
		ID:        st.ID,
		SiteName:  st.Result.Meta.SiteName,
		URL:       st.Result.Meta.URL,
		Origin:    st.Origin,
		ETLD1:     st.ETLD1,
		Subdomain: st.Subdomain,
		SourceURL: st.SourceURL,
		Theme:     st.Result.DesignSystem.Theme,
		Industry:  st.Result.DesignSystem.Industry,
		NorthStar: st.Result.DesignSystem.NorthStar,
		Extracted: extracted,
		ResultSHA: st.ResultSHA,
		Media:     media,
	}
}

// StyleRow identifies one archived style for offline passes.
type StyleRow struct {
	ID     string
	Origin string
}

// StylesWithResult lists styles that have a stored extraction result.
func (s *Store) StylesWithResult(ctx context.Context, only []string) ([]StyleRow, error) {
	q := `SELECT st.id, COALESCE(si.origin,'') FROM styles st
	      LEFT JOIN sites si ON si.id = st.site_id
	      WHERE st.has_result = 1 ORDER BY st.id`
	rows, err := s.DB.SQL().QueryContext(ctx, q)
	if err != nil {
		return nil, err
	}
	defer rows.Close()
	want := map[string]bool{}
	for _, id := range only {
		want[id] = true
	}
	var out []StyleRow
	for rows.Next() {
		var r StyleRow
		if err := rows.Scan(&r.ID, &r.Origin); err != nil {
			return nil, err
		}
		if len(want) > 0 && !want[r.ID] {
			continue
		}
		out = append(out, r)
	}
	return out, rows.Err()
}

// ReadRawResult loads the archived result.json for a style.
func (s *Store) ReadRawResult(ctx context.Context, r StyleRow) ([]byte, error) {
	return os.ReadFile(filepath.Join(s.FS.StyleDir(r.Origin, r.ID), "result.json"))
}

// RefreshStyleFile rewrites style.json from the database, picking up media discovered on
// other pages after the style itself was crawled.
func (s *Store) RefreshStyleFile(ctx context.Context, r StyleRow) error {
	var f styleFile
	var subdomain, etld1, theme, industry, northStar, extracted, sourceURL, sha string
	err := s.DB.SQL().QueryRowContext(ctx, `
		SELECT st.id, COALESCE(st.site_name,''), COALESCE(st.url,''), COALESCE(si.origin,''),
		       COALESCE(si.etld1,''), COALESCE(si.subdomain,''), COALESCE(st.source_url,''),
		       COALESCE(st.theme,''), COALESCE(st.industry,''), COALESCE(st.north_star,''),
		       COALESCE(st.extracted_at,''), COALESCE(st.result_sha256,'')
		FROM styles st LEFT JOIN sites si ON si.id = st.site_id WHERE st.id = ?`, r.ID).
		Scan(&f.ID, &f.SiteName, &f.URL, &f.Origin, &etld1, &subdomain, &sourceURL,
			&theme, &industry, &northStar, &extracted, &sha)
	if err != nil {
		return err
	}
	f.ETLD1, f.Subdomain, f.SourceURL = etld1, subdomain, sourceURL
	f.Theme, f.Industry, f.NorthStar, f.Extracted, f.ResultSHA = theme, industry, northStar, extracted, sha

	rows, err := s.DB.SQL().QueryContext(ctx, `
		SELECT sm.role, m.kind, m.url, COALESCE(m.width,0), COALESCE(m.height,0),
		       COALESCE(m.duration_ms,0), m.downloaded
		FROM style_media sm JOIN media m ON m.id = sm.media_id
		WHERE sm.style_id = ? ORDER BY sm.role`, r.ID)
	if err != nil {
		return err
	}
	defer rows.Close()
	f.Media = []mediaEntry{}
	for rows.Next() {
		var e mediaEntry
		var downloaded int
		if err := rows.Scan(&e.Role, &e.Kind, &e.URL, &e.Width, &e.Height, &e.DurationMs, &downloaded); err != nil {
			return err
		}
		e.Downloaded = downloaded == 1
		f.Media = append(f.Media, e)
	}
	if err := rows.Err(); err != nil {
		return err
	}

	body, err := json.MarshalIndent(f, "", "  ")
	if err != nil {
		return err
	}
	_, err = s.FS.WriteFile(s.FS.StyleDir(f.Origin, f.ID), "style.json", body, false)
	return err
}
