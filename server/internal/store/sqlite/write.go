package sqlite

import (
	"context"
	"database/sql"
	"fmt"
	"strings"

	"github.com/kanak/design-supply/internal/adapter"
	"github.com/kanak/design-supply/internal/normalize"
	"github.com/kanak/design-supply/internal/sites/refero"
)

// Write persists one parsed page in a single transaction. Records are adapter-owned, so
// this is the one place that knows concrete site types; a second site adds a case here.
func (d *DB) Write(ctx context.Context, t adapter.Target, res adapter.Result) error {
	if len(res.Records) == 0 {
		return nil
	}
	tx, err := d.db.BeginTx(ctx, nil)
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			tx.Rollback()
			d.dropCaches() // interned ids from a rolled-back tx never existed
		}
	}()

	for _, rec := range res.Records {
		switch v := rec.Data.(type) {
		case *refero.Style:
			if err := d.writeStyle(ctx, tx, v, t.URL); err != nil {
				return fmt.Errorf("write style %s: %w", v.ID, err)
			}
		case refero.Card:
			if err := d.writeCard(ctx, tx, v); err != nil {
				return fmt.Errorf("write card %s: %w", v.ID, err)
			}
		case refero.Collection:
			if err := d.writeCollection(ctx, tx, v); err != nil {
				return fmt.Errorf("write collection %s: %w", v.Slug, err)
			}
		default:
			return fmt.Errorf("sqlite: unknown record type %T", rec.Data)
		}
	}
	if err := tx.Commit(); err != nil {
		return err
	}
	committed = true
	return nil
}

func (d *DB) siteID(ctx context.Context, tx *sql.Tx, o normalize.Origin, name string) (int64, error) {
	if o.Origin == "" {
		return 0, nil
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO sites(origin, etld1, subdomain, site_name) VALUES(?,?,?,?)
		ON CONFLICT(origin) DO UPDATE SET
			last_seen=datetime('now'),
			site_name=COALESCE(NULLIF(excluded.site_name,''), sites.site_name)`,
		o.Origin, o.ETLD1, o.Subdomain, name); err != nil {
		return 0, err
	}
	var id int64
	err := tx.QueryRowContext(ctx, `SELECT id FROM sites WHERE origin=?`, o.Origin).Scan(&id)
	return id, err
}

// writeCard upserts the summary record. It never clears fields the result page owns,
// because a card and its result arrive from different pages.
func (d *DB) writeCard(ctx context.Context, tx *sql.Tx, c refero.Card) error {
	if c.ID == "" {
		return nil
	}
	var siteID any
	if c.URL != "" {
		o, err := normalize.Identify(c.URL)
		if err == nil {
			id, err := d.siteID(ctx, tx, o, c.SiteName)
			if err != nil {
				return err
			}
			siteID = id
		}
	}
	created := ""
	if ts, ok := c.CreatedTime(); ok {
		created = ts.Format("2006-01-02T15:04:05Z")
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO styles(id, site_id, url, site_name, color_scheme, north_star, created_at, has_card)
		VALUES(?,?,?,?,?,?,?,1)
		ON CONFLICT(id) DO UPDATE SET
			site_id      = COALESCE(excluded.site_id, styles.site_id),
			url          = COALESCE(NULLIF(excluded.url,''), styles.url),
			site_name    = COALESCE(NULLIF(excluded.site_name,''), styles.site_name),
			color_scheme = COALESCE(NULLIF(excluded.color_scheme,''), styles.color_scheme),
			north_star   = COALESCE(NULLIF(excluded.north_star,''), styles.north_star),
			created_at   = COALESCE(NULLIF(excluded.created_at,''), styles.created_at),
			has_card     = 1`,
		c.ID, siteID, c.URL, c.SiteName, c.ColorScheme, c.NorthStar, created); err != nil {
		return err
	}

	for i, cc := range c.Colors {
		id, err := d.colorID(ctx, tx, cc.Hex, 0, 0, 0)
		if err != nil {
			return err
		}
		if id == 0 {
			continue
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO style_colors(style_id, color_id, origin, name, gradient, ord)
			VALUES(?,?,'card',?,?,?)
			ON CONFLICT(style_id, color_id, origin) DO UPDATE SET
				name=excluded.name, gradient=excluded.gradient, ord=excluded.ord`,
			c.ID, id, cc.Name, cc.Gradient, i); err != nil {
			return err
		}
	}

	return d.writeCardMedia(ctx, tx, c)
}

type mediaRef struct {
	url      string
	kind     string
	role     string
	w, h     int
	duration int
}

func (d *DB) writeCardMedia(ctx context.Context, tx *sql.Tx, c refero.Card) error {
	refs := []mediaRef{
		{c.ScreenshotURL, "screenshot", "screenshot", 0, 0, 0},
		{c.ThumbnailURL, "thumbnail", "thumbnail", 0, 0, 0},
		{c.IconURL, "icon", "icon", 0, 0, 0},
		{c.PreviewVideoURL, "video", "preview_video", c.PreviewVideoWidth, c.PreviewVideoHeight, c.PreviewVideoDurationMs},
		{c.PreviewVideoPosterURL, "video_poster", "preview_video_poster", c.PreviewVideoWidth, c.PreviewVideoHeight, 0},
		{c.PreviewVideoDetailURL, "video", "preview_video_detail", c.PreviewVideoDetailWidth, c.PreviewVideoDetailHeight, 0},
		{c.PreviewVideoDetailPosterURL, "video_poster", "preview_video_detail_poster", c.PreviewVideoDetailWidth, c.PreviewVideoDetailHeight, 0},
	}
	for _, r := range refs {
		if err := d.linkMedia(ctx, tx, c.ID, r); err != nil {
			return err
		}
	}
	return nil
}

func (d *DB) linkMedia(ctx context.Context, tx *sql.Tx, styleID string, r mediaRef) error {
	if r.url == "" || !strings.HasPrefix(r.url, "http") {
		return nil
	}
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO media(url, kind, width, height, duration_ms) VALUES(?,?,?,?,?)
		ON CONFLICT(url) DO UPDATE SET
			width       = COALESCE(NULLIF(excluded.width,0), media.width),
			height      = COALESCE(NULLIF(excluded.height,0), media.height),
			duration_ms = COALESCE(NULLIF(excluded.duration_ms,0), media.duration_ms)`,
		r.url, r.kind, nullIfZero(r.w), nullIfZero(r.h), nullIfZero(r.duration)); err != nil {
		return err
	}
	var id int64
	if err := tx.QueryRowContext(ctx, `SELECT id FROM media WHERE url=?`, r.url).Scan(&id); err != nil {
		return err
	}
	_, err := tx.ExecContext(ctx, `
		INSERT INTO style_media(style_id, media_id, role) VALUES(?,?,?)
		ON CONFLICT(style_id, role) DO UPDATE SET media_id=excluded.media_id`,
		styleID, id, r.role)
	return err
}

func nullIfZero(n int) any {
	if n == 0 {
		return nil
	}
	return n
}

func (d *DB) writeStyle(ctx context.Context, tx *sql.Tx, s *refero.Style, sourceURL string) error {
	res, ds, raw := s.Result, s.Result.DesignSystem, s.Result.Raw

	siteID, err := d.siteID(ctx, tx,
		normalize.Origin{Origin: s.Origin, ETLD1: s.ETLD1, Subdomain: s.Subdomain}, res.Meta.SiteName)
	if err != nil {
		return err
	}
	extracted := ""
	if ts, ok := res.ExtractedAt(); ok {
		extracted = ts.Format("2006-01-02T15:04:05Z")
	}
	var baseUnit any
	if raw.Spacing.BaseUnit != nil {
		baseUnit = *raw.Spacing.BaseUnit
	}

	if _, err := tx.ExecContext(ctx, `
		INSERT INTO styles(id, site_id, url, site_name, source_url, theme, industry, north_star,
			north_star_detail, description, layout, imagery, is_dark_mode, colorfulness, density,
			base_unit, scale_base, scale_name, scale_ratio, scale_confidence, viewport_w, viewport_h,
			element_count, duration_ms, extracted_at, fetched_at, result_sha256, has_result)
		VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,1)
		ON CONFLICT(id) DO UPDATE SET
			site_id=excluded.site_id, url=excluded.url, site_name=excluded.site_name,
			source_url=excluded.source_url, theme=excluded.theme, industry=excluded.industry,
			north_star=excluded.north_star, north_star_detail=excluded.north_star_detail,
			description=excluded.description, layout=excluded.layout, imagery=excluded.imagery,
			is_dark_mode=excluded.is_dark_mode, colorfulness=excluded.colorfulness,
			density=excluded.density, base_unit=excluded.base_unit, scale_base=excluded.scale_base,
			scale_name=excluded.scale_name, scale_ratio=excluded.scale_ratio,
			scale_confidence=excluded.scale_confidence, viewport_w=excluded.viewport_w,
			viewport_h=excluded.viewport_h, element_count=excluded.element_count,
			duration_ms=excluded.duration_ms, extracted_at=excluded.extracted_at,
			fetched_at=excluded.fetched_at, result_sha256=excluded.result_sha256, has_result=1`,
		s.ID, siteID, res.Meta.URL, res.Meta.SiteName, sourceURL, ds.Theme, ds.Industry,
		ds.NorthStar, ds.NorthStarDetail, ds.Description, ds.Layout, ds.Imagery,
		boolInt(raw.Colors.IsDarkMode), raw.Colors.Colorfulness, raw.Spacing.Density,
		baseUnit, raw.Typography.Scale.Base, raw.Typography.Scale.Name,
		raw.Typography.Scale.Ratio, raw.Typography.Scale.Confidence,
		res.Meta.Viewport.Width, res.Meta.Viewport.Height, res.Meta.ElementCount,
		res.Meta.DurationMs, extracted, now(), s.ResultSHA); err != nil {
		return err
	}

	for _, clear := range []string{
		"DELETE FROM color_usage WHERE style_id=?",
		"DELETE FROM contrast_pairs WHERE style_id=?",
		"DELETE FROM type_steps WHERE style_id=?",
		"DELETE FROM type_scale WHERE style_id=?",
		"DELETE FROM typography_roles WHERE style_id=?",
		"DELETE FROM elevation WHERE style_id=?",
		"DELETE FROM components WHERE style_id=?",
		"DELETE FROM guidelines WHERE style_id=?",
		"DELETE FROM similar WHERE style_id=?",
		"DELETE FROM spacing_map WHERE style_id=?",
		"DELETE FROM custom_sections WHERE style_id=?",
		"DELETE FROM style_colors WHERE style_id=? AND origin IN ('raw','system','surface')",
	} {
		if _, err := tx.ExecContext(ctx, clear, s.ID); err != nil {
			return err
		}
	}

	if err := d.writeColors(ctx, tx, s); err != nil {
		return err
	}
	if err := d.writeTypography(ctx, tx, s); err != nil {
		return err
	}
	if err := d.writeShapes(ctx, tx, s); err != nil {
		return err
	}
	if err := d.writeDescriptive(ctx, tx, s); err != nil {
		return err
	}
	return d.writeCardMedia(ctx, tx, s.Card)
}

func (d *DB) writeColors(ctx context.Context, tx *sql.Tx, s *refero.Style) error {
	raw, ds := s.Result.Raw, s.Result.DesignSystem

	for i, tok := range raw.Colors.Tokens {
		id, err := d.colorID(ctx, tx, tok.Hex, tok.OKLCH.L, tok.OKLCH.C, tok.OKLCH.H)
		if err != nil || id == 0 {
			if err != nil {
				return err
			}
			continue
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO style_colors(style_id, color_id, origin, frequency, confidence, prominence, ord)
			VALUES(?,?,'raw',?,?,?,?)
			ON CONFLICT(style_id, color_id, origin) DO UPDATE SET
				frequency=excluded.frequency, confidence=excluded.confidence,
				prominence=excluded.prominence, ord=excluded.ord`,
			s.ID, id, tok.Frequency, tok.Confidence, tok.Prominence, i); err != nil {
			return err
		}
		for pair, count := range tok.UsageCounts {
			ctxName, prop, ok := strings.Cut(pair, "/")
			if !ok {
				continue
			}
			if _, err := tx.ExecContext(ctx, `
				INSERT INTO color_usage(style_id, color_id, context, property, count) VALUES(?,?,?,?,?)
				ON CONFLICT(style_id, color_id, context, property) DO UPDATE SET count=excluded.count`,
				s.ID, id, ctxName, prop, count); err != nil {
				return err
			}
		}
	}

	for i, c := range ds.Colors {
		id, err := d.colorID(ctx, tx, c.Hex, 0, 0, 0)
		if err != nil {
			return err
		}
		if id == 0 {
			continue
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO style_colors(style_id, color_id, origin, name, role, group_name, ord)
			VALUES(?,?,'system',?,?,?,?)
			ON CONFLICT(style_id, color_id, origin) DO UPDATE SET
				name=excluded.name, role=excluded.role, group_name=excluded.group_name, ord=excluded.ord`,
			s.ID, id, c.Name, c.Role, c.Group, i); err != nil {
			return err
		}
	}

	for i, sf := range ds.Surfaces {
		id, err := d.colorID(ctx, tx, sf.Hex, 0, 0, 0)
		if err != nil {
			return err
		}
		if id == 0 {
			continue
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO style_colors(style_id, color_id, origin, name, level, purpose, ord)
			VALUES(?,?,'surface',?,?,?,?)
			ON CONFLICT(style_id, color_id, origin) DO UPDATE SET
				name=excluded.name, level=excluded.level, purpose=excluded.purpose, ord=excluded.ord`,
			s.ID, id, sf.Name, sf.Level, sf.Purpose, i); err != nil {
			return err
		}
	}

	for _, p := range raw.Colors.ContrastPairs {
		fg, err := d.colorID(ctx, tx, p.Foreground, 0, 0, 0)
		if err != nil {
			return err
		}
		bg, err := d.colorID(ctx, tx, p.Background, 0, 0, 0)
		if err != nil {
			return err
		}
		if fg == 0 || bg == 0 {
			continue
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO contrast_pairs(style_id, fg_color_id, bg_color_id, ratio, level) VALUES(?,?,?,?,?)
			ON CONFLICT(style_id, fg_color_id, bg_color_id) DO UPDATE SET
				ratio=excluded.ratio, level=excluded.level`,
			s.ID, fg, bg, p.Ratio, p.Level); err != nil {
			return err
		}
	}

	for _, g := range s.Result.Raw.Gradients {
		id, err := d.valueID(ctx, tx, "gradients", g.Value, g.Type)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO style_gradients(style_id, gradient_id, frequency) VALUES(?,?,?)
			ON CONFLICT(style_id, gradient_id) DO UPDATE SET frequency=excluded.frequency`,
			s.ID, id, g.Frequency); err != nil {
			return err
		}
	}
	return nil
}

func (d *DB) writeTypography(ctx context.Context, tx *sql.Tx, s *refero.Style) error {
	raw, ds := s.Result.Raw, s.Result.DesignSystem
	for _, f := range raw.Typography.Fonts {
		id, err := d.fontID(ctx, tx, f.Family, f.Source)
		if err != nil {
			return err
		}
		if id == 0 {
			continue
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO style_fonts(style_id, font_id, weights, frequency, features) VALUES(?,?,?,?,?)
			ON CONFLICT(style_id, font_id) DO UPDATE SET
				weights=excluded.weights, frequency=excluded.frequency, features=excluded.features`,
			s.ID, id, jsonOrEmpty(f.Weights), f.Frequency, jsonOrEmpty(f.FontFeatureSettings)); err != nil {
			return err
		}
	}
	for i, st := range raw.Typography.Steps {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO type_steps(style_id, ord, size, family, weight, line_height, letter_spacing,
				text_transform, contexts, frequency) VALUES(?,?,?,?,?,?,?,?,?,?)`,
			s.ID, i, st.Size, st.Family, st.Weight, st.LineHeight, st.LetterSpacing,
			st.TextTransform, jsonOrEmpty(st.Contexts), st.Frequency); err != nil {
			return err
		}
	}
	for i, ts := range ds.TypeScale {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO type_scale(style_id, ord, role, size, line_height, letter_spacing)
			VALUES(?,?,?,?,?,?)`,
			s.ID, i, ts.Role, ts.Size, ts.LineHeight, ts.LetterSpacing); err != nil {
			return err
		}
	}
	for i, tr := range ds.Typography {
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO typography_roles(style_id, ord, role, family, sizes, weight, line_height,
				substitute, letter_spacing, font_features) VALUES(?,?,?,?,?,?,?,?,?,?)`,
			s.ID, i, tr.Role, tr.Family, tr.Sizes, tr.Weight, tr.LineHeight,
			tr.Substitute, tr.LetterSpacing, tr.FontFeatureSettings); err != nil {
			return err
		}
	}
	return nil
}

func (d *DB) writeShapes(ctx context.Context, tx *sql.Tx, s *refero.Style) error {
	raw := s.Result.Raw
	for _, r := range raw.Shapes.Radii {
		id, err := d.valueID(ctx, tx, "radii", r.Value)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO style_radii(style_id, radius_id, contexts, frequency) VALUES(?,?,?,?)
			ON CONFLICT(style_id, radius_id) DO UPDATE SET
				contexts=excluded.contexts, frequency=excluded.frequency`,
			s.ID, id, jsonOrEmpty(r.Contexts), r.Frequency); err != nil {
			return err
		}
	}
	for _, sh := range raw.Shapes.Shadows {
		id, err := d.valueID(ctx, tx, "shadows", sh.Value)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO style_shadows(style_id, shadow_id, contexts, frequency) VALUES(?,?,?,?)
			ON CONFLICT(style_id, shadow_id) DO UPDATE SET
				contexts=excluded.contexts, frequency=excluded.frequency`,
			s.ID, id, jsonOrEmpty(sh.Contexts), sh.Frequency); err != nil {
			return err
		}
	}
	for _, sp := range raw.Spacing.Tokens {
		id, err := d.valueID(ctx, tx, "spacings", sp.Value)
		if err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO style_spacings(style_id, spacing_id, contexts, properties, pairs, frequency)
			VALUES(?,?,?,?,?,?)
			ON CONFLICT(style_id, spacing_id) DO UPDATE SET
				contexts=excluded.contexts, properties=excluded.properties,
				pairs=excluded.pairs, frequency=excluded.frequency`,
			s.ID, id, jsonOrEmpty(sp.Contexts), jsonOrEmpty(sp.Properties),
			jsonOrEmpty(sp.PropertyContextPairs), sp.Frequency); err != nil {
			return err
		}
	}
	return nil
}

func (d *DB) writeDescriptive(ctx context.Context, tx *sql.Tx, s *refero.Style) error {
	ds := s.Result.DesignSystem
	for i, e := range ds.Elevation {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO elevation(style_id, ord, element, style) VALUES(?,?,?,?)`,
			s.ID, i, e.Element, e.Style); err != nil {
			return err
		}
	}
	for i, c := range ds.Components {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO components(style_id, ord, name, role, description) VALUES(?,?,?,?,?)`,
			s.ID, i, c.Name, c.Role, c.Description); err != nil {
			return err
		}
	}
	for i, g := range ds.Dos {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO guidelines(style_id, kind, ord, text) VALUES(?,'do',?,?)`, s.ID, i, g); err != nil {
			return err
		}
	}
	for i, g := range ds.Donts {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO guidelines(style_id, kind, ord, text) VALUES(?,'dont',?,?)`, s.ID, i, g); err != nil {
			return err
		}
	}
	for i, sb := range ds.Similar {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO similar(style_id, ord, business, why) VALUES(?,?,?,?)`,
			s.ID, i, sb.Business, sb.Why); err != nil {
			return err
		}
	}
	for i, cs := range ds.CustomSections {
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO custom_sections(style_id, ord, title, content) VALUES(?,?,?,?)`,
			s.ID, i, cs.Title, cs.Content); err != nil {
			return err
		}
	}
	sp := ds.Spacing
	pairs := map[string]string{
		"elementGap":   sp.ElementGap,
		"sectionGap":   sp.SectionGap,
		"cardPadding":  sp.CardPadding,
		"pageMaxWidth": sp.PageMaxWidth,
	}
	for _, k := range sp.Radius.Keys {
		pairs["radius."+k] = sp.Radius.Get(k)
	}
	for k, v := range pairs {
		if v == "" {
			continue
		}
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO spacing_map(style_id, key, value) VALUES(?,?,?)
			 ON CONFLICT(style_id, key) DO UPDATE SET value=excluded.value`, s.ID, k, v); err != nil {
			return err
		}
	}
	return nil
}

func (d *DB) writeCollection(ctx context.Context, tx *sql.Tx, c refero.Collection) error {
	if _, err := tx.ExecContext(ctx, `
		INSERT INTO collections(slug, url, title, kind) VALUES(?,?,?,?)
		ON CONFLICT(slug) DO UPDATE SET
			url=excluded.url,
			title=COALESCE(NULLIF(excluded.title,''), collections.title),
			kind=excluded.kind`,
		c.Slug, c.URL, c.Title, c.Kind); err != nil {
		return err
	}
	var id int64
	if err := tx.QueryRowContext(ctx, `SELECT id FROM collections WHERE slug=?`, c.Slug).Scan(&id); err != nil {
		return err
	}
	for i, sid := range c.StyleIDs {
		// A collection can name a style this crawl has not reached yet; stub it so the
		// membership edge survives and the real row fills in later.
		if _, err := tx.ExecContext(ctx,
			`INSERT INTO styles(id) VALUES(?) ON CONFLICT(id) DO NOTHING`, sid); err != nil {
			return err
		}
		if _, err := tx.ExecContext(ctx, `
			INSERT INTO collection_styles(collection_id, style_id, ord) VALUES(?,?,?)
			ON CONFLICT(collection_id, style_id) DO UPDATE SET ord=excluded.ord`,
			id, sid, i); err != nil {
			return err
		}
	}
	return nil
}

func boolInt(b bool) int {
	if b {
		return 1
	}
	return 0
}

// RecordArtifact notes a generated export file for a style.
func (d *DB) RecordArtifact(ctx context.Context, styleID, name, version, sha, localPath string, n int64) error {
	_, err := d.db.ExecContext(ctx, `
		INSERT INTO artifacts(style_id, name, version, sha256, bytes, local_path, rendered_at)
		VALUES(?,?,?,?,?,?,datetime('now'))
		ON CONFLICT(style_id, name) DO UPDATE SET
			version=excluded.version, sha256=excluded.sha256, bytes=excluded.bytes,
			local_path=excluded.local_path, rendered_at=excluded.rendered_at`,
		styleID, name, version, sha, n, localPath)
	return err
}
