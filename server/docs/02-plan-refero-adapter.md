# Plan 2: the Refero adapter, schema and storage

> Pre-build plan, kept for the record. The system as built is described in
> [03-implementation.md](03-implementation.md), which supersedes this where they differ.

What gets extracted from styles.refero.design, how it is identified, deduplicated and
written. Assumes the findings in `00-findings.md`.

## Crawl stages

1. **Discover.** Fetch `sitemaps/styles.xml` (1290 style URLs with `lastmod` and a
   thumbnail `image:loc`) and `sitemaps/collections.xml` (51 URLs, of which 10 are
   `/design-styles/<slug>` category listings worth crawling for tag edges).
2. **Fetch.** `GET /style/<uuid>` with `RSC: 1`. Archive the body gzipped to
   `data/raw/<uuid>.flight.txt.gz`.
3. **Parse.** Flight parser resolves `T<hexlen>,` text rows and `$<hex>` references, then
   pulls the `result` object and every embedded style card.
4. **Close the graph.** Every card id found is pushed back to the frontier. Loop until no
   new ids appear. This measures the true inventory instead of trusting either the
   sitemap's 1290 or the homepage's "2,000+".
5. **Normalise.** Domain identity, colour and font dedup, slugging.
6. **Render.** Emit the five artifacts from the stored `result`.
7. **Media.** Record every URL; download the selected kinds.
8. **Store.** Write files, upsert SQLite, in one batched transaction per page.

Budget at the default 2 rps with 4 workers: ~1290 pages in roughly 11 minutes, ~145 MB of
flight payload archived gzipped (~25 MB on disk).

## Flight parser

`internal/sites/refero/flight/` is generic Next.js RSC, not Refero-specific, and is the
piece most likely to be reused for another Next.js site.

```go
type Doc struct {
    rows map[string]json.RawMessage // id -> parsed row
    text map[string]string          // id -> T-row text
}

func Parse(body []byte) (*Doc, error)          // accepts raw RSC or HTML with __next_f
func (d *Doc) Find(key string) (json.RawMessage, bool) // first object carrying a key
func (d *Doc) Resolve(v any) error             // rewrites "$<hex>" to the text row
```

Two input shapes are supported: the `RSC: 1` body directly, and HTML with
`self.__next_f.push([1,"..."])` chunks, for when a response arrives as HTML. Golden tests
run off `testdata/fixtures/style-shade.flight.txt`.

## Typed model

```go
type Style struct {
    ID         string    // refero uuid, primary key
    URL        string    // https://shade.inc
    SiteName   string
    ExtractedAt time.Time
    CreatedAt   time.Time
    ColorScheme string   // light | dark
    Viewport    Size
    ElementCount int
    Raw         Raw          // raw.colors, shapes, spacing, gradients, typography
    System      DesignSystem // the AI-authored layer
    Card        Card         // media URLs, gradients, northStar, managementSignals
    ResultSHA   string
}
```

`Raw` and `DesignSystem` mirror the payload field for field, per the inventory in
`00-findings.md`. Nothing is dropped. Unknown future fields land in a `Extra
map[string]json.RawMessage` catch-all so a schema change upstream never silently loses data.

## Identity and grouping

Chosen model: the style is the record, the site is a grouping key. Nothing is merged.

```go
func Identify(rawURL string) Origin
// https://www.11x.ai      -> origin=11x.ai        etld1=11x.ai   sub=""
// https://dub.sh          -> origin=dub.sh        etld1=dub.sh   sub=""
// https://app.example.com -> origin=app.example.com etld1=example.com sub="app"
```

`www.` is stripped, everything else is preserved, eTLD+1 comes from
`golang.org/x/net/publicsuffix`. Two Refero entries for `dub.sh` and `dub.co` stay two
sites; `app.foo.com` and `foo.com` stay two sites with a shared `etld1` you can group on in
a query when you want to.

## On-disk layout

```
data/
  sites/<origin>/<uuid>/
    style.json              card metadata, media URLs, normalised identity
    result.json             the full extraction payload, source of truth
    design.compact.md
    design.extended.md
    tailwind.v4.css
    variables.css
    tokens.json
    media/
      icon.png
      video-poster.jpg
      video-detail-poster.jpg
  raw/<uuid>.flight.txt.gz
  index.db
```

`data/sites/shade.inc/e549766e-.../`. Origin-first means the directory tree is already the
"uniquely identified by website" view, and a company with several entries groups naturally.

## SQLite schema

Repeated values live once in a shared table and styles reference them. That is the single
source of truth you asked for: `Inter` is one row, `#ffffff` is one row, a given shadow
string is one row, no matter how many of the 1290 styles use it.

```sql
-- identity
sites(id PK, origin UNIQUE, etld1, subdomain, site_name, first_seen, last_seen)
styles(id PK /*uuid*/, site_id FK, url, site_name, color_scheme, theme, industry,
       north_star, north_star_detail, description, layout, imagery,
       is_dark_mode, colorfulness, density, base_unit,
       scale_base, scale_name, scale_ratio, scale_confidence,
       viewport_w, viewport_h, element_count,
       extracted_at, created_at, fetched_at, etag, last_modified, result_sha256)

-- shared value tables (dedup)
colors(id PK, hex UNIQUE, oklch_l, oklch_c, oklch_h)
fonts(id PK, family, source, UNIQUE(family, source))
shadows(id PK, value UNIQUE)
radii(id PK, value UNIQUE)
spacings(id PK, value UNIQUE)
gradients(id PK, value UNIQUE, type)

-- style to value edges, carrying the per-style facts
style_colors(style_id, color_id, origin /*raw|system*/, name, role, group_name,
             frequency, confidence, prominence, PRIMARY KEY(style_id, color_id, origin))
color_usage(style_id, color_id, context, property, count)
contrast_pairs(style_id, fg_color_id, bg_color_id, ratio, level)
style_fonts(style_id, font_id, weights_json, frequency, features_json)
style_shadows(style_id, shadow_id, context, frequency)
style_radii(style_id, radius_id, context, frequency)
style_spacings(style_id, spacing_id, property, context, frequency)
style_gradients(style_id, gradient_id, frequency)

-- the descriptive layer
type_steps(style_id, size, family, weight, line_height, letter_spacing, text_transform,
           context, frequency)
type_scale(style_id, role, size, line_height, letter_spacing)
typography_roles(style_id, role, family, sizes, weight, line_height)
surfaces(style_id, level, hex, name, purpose)
elevation(style_id, element, style)
components(style_id, name, role, description)
guidelines(style_id, kind /*do|dont*/, ord, text)
similar(style_id, business, why)
spacing_map(style_id, key /*cards|buttons|inputs|tabs|elementGap|...*/, value)
custom_sections(style_id, title, content)

-- media: every URL recorded, download optional
media(id PK, url UNIQUE, kind, sha256, bytes, width, height, duration_ms,
      local_path, downloaded INT DEFAULT 0, checked_at)
style_media(style_id, media_id, role)

-- collections and tags
collections(id PK, slug UNIQUE, url, title, kind)
collection_styles(collection_id, style_id)

-- artifacts and run bookkeeping
artifacts(style_id, name, version, sha256, bytes, rendered_at,
          PRIMARY KEY(style_id, name))
runs(id PK, adapter, started_at, finished_at, pages, bytes, errors, notes)
fetch_log(id PK, run_id, url, status, bytes, ms, attempt, error)
```

Indexes on `styles(site_id)`, `sites(etld1)`, `style_colors(color_id)`,
`style_fonts(font_id)`, `media(kind, downloaded)`. FTS5 virtual table over
`styles(site_name, north_star, description)` plus `components(name, description)` so the
private viewer gets real search for free.

Queries this unlocks: every style using Inter, the most common radius across the corpus,
all dark-mode fintech systems, colours that appear in more than fifty systems, styles whose
palette is within a distance of a given hex.

## Media policy

Every URL is recorded for every kind, always, with dimensions and duration where the payload
supplies them, so a deployment can hotlink the Refero CDN on day one and backfill local
bytes later. `downloaded` and `local_path` track what is actually on disk.

| kind | URL recorded | bytes downloaded by default |
|---|---|---|
| icon / favicon | yes | yes (~9 MB total) |
| preview video poster | yes | yes |
| preview video detail poster | yes | yes |
| thumbnail | yes | no |
| full screenshot | yes | no |
| preview video mp4 | yes | no |
| detail video mp4 | yes | no |

`--media` changes the download set without re-crawling, because the URLs are already in the
database: `supply media --site refero --kinds thumbnails` backfills later.

### Pushback on skipping screenshots and thumbnails

Runtime generation (microlink, urlbox and similar) screenshots the site **as it is today**.
The tokens in this dataset were extracted on a specific date, and `extractedAt` values in
the corpus already range across months. A regenerated screenshot will drift away from the
palette and typography stored next to it, and for a site that has since redesigned it will
contradict them outright. The Refero screenshot is the exact frame the extraction was
derived from, which makes it provenance, not decoration.

Cost of keeping both: ~140 MB screenshots plus ~48 MB thumbnails, one time, versus a
per-request dependency on a third-party service with its own rate limits and a live-site
round trip per card.

Recommendation: download thumbnails at minimum for the grid, and screenshots for
provenance. Runtime generation is still the right tool for sites that are *not* in this
dataset. The default in the plan stays as chosen; flipping it is one flag:

```
supply media --site refero --kinds thumbnails,screenshots
```

## Artifact generation

Five outputs per style, all derived from `result.json`, all produced offline by
`supply render`:

| file | ported from |
|---|---|
| `design.compact.md` | `reference/refero-js/0jge9-tu1j3~c.js` |
| `design.extended.md` | same chunk, extended branch |
| `tailwind.v4.css` | `0kxtp02n4q1qc.js` theme emitter |
| `variables.css` | `0kxtp02n4q1qc.js` custom-properties emitter |
| `tokens.json` | `0jge9-tu1j3~c.js` tokens sections |

Porting approach:

1. Beautify the pinned chunks and extract the generator functions, including the slug rule
   (`lowercase, non-alnum to dash, collapse, trim`), the font-stack fallback rules
   (`/mono|code|consola/i` to a monospace stack) and the luminance-based role inference that
   picks primary, secondary, background, foreground, muted and border.
2. Reimplement in `internal/sites/refero/render/`, one file per output, pure functions over
   the typed model, no globals.
3. Golden tests against captured browser output for the fixture styles.

### Fidelity verification

`supply verify --site refero --sample 20` launches chromedp, loads N sampled style pages,
clicks each of the five tabs, reads the generated text, and diffs it against the Go
renderer. Success criterion is byte-identical output; any diff is printed with context and
fails the command. This runs once during the port and thereafter on demand.

### Drift detection

`reference/refero-js/SHA256SUMS` pins the upstream chunk hashes. `supply verify` re-fetches
the current chunk filenames from a live page and flags a hash change, which is the signal
that the generators may have moved and the sample diff should be re-run. Chunk filenames are
content-hashed, so a rename alone is the tell.

`artifacts.version` records which generator version produced each file, so a re-render after
a port fix is traceable rather than silent.

## Collections as tags

The 10 `/design-styles/<slug>` pages (clean-saas, editorial-websites, dark-mode-websites,
fintech-websites, devtools-websites, ai-startup-websites, ecommerce-websites,
minimal-websites, agency-websites, productivity-apps) are server-rendered and each embeds
its member style cards. Crawling them fills `collection_styles`, giving every style a set of
human-meaningful tags on top of the `industry` field in the payload.

## Build order

1. Skeleton: module, cobra CLI, slog, config.
2. Flight parser plus golden test on the saved fixture.
3. Typed model plus `result` extraction, verified against the fixture.
4. SQLite store, schema, migrations, upserts.
5. Engine: frontier, limiter, robots, retries, resume.
6. Full crawl of the 1290, graph closure, inventory report.
7. Renderers, one at a time, each validated by `verify` before the next.
8. Media downloader.
9. Export and stats, ready for the private viewer.

Stages 1 to 6 give a complete, queryable dataset even before any renderer is ported, because
`result.json` is retained verbatim.
