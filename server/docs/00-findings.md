# Site recon: styles.refero.design

Recorded 2026-09-16. Everything below was verified against the live site, not assumed.

## Stack

Next.js (App Router) on Vercel. Server components, so every `/style/<uuid>` page ships the
complete extraction payload inside the RSC flight stream embedded in the HTML.

## The key finding

`GET /style/<uuid>` with header `RSC: 1` returns the raw flight payload as
`text/x-component`: **113 KB versus 278 KB** for the full HTML, same data. No API call, no
browser, no JS execution needed for the source data.

Measured: 0.60s to 1.92s per page, median ~0.77s, 200 OK throughout, no rate limiting
observed at 1 req/s over 12 sequential requests.

## robots.txt

```
User-Agent: anthropic-ai, Applebot-Extended, Bytespider, CCBot, Claude-Web, ClaudeBot,
            cohere-ai, Diffbot, FacebookBot, Google-Extended, GPTBot, ImagesiftBot,
            meta-externalagent, Omgilibot, PerplexityBot
Disallow: /

User-Agent: *
Allow: /
Disallow: /api/ /admin/ /extract/ /playground/
```

A normal browser UA falls under `*`. `/style/...` is allowed; the `RSC: 1` variant is the
same URL, not an `/api/` path. `images.refero.design/robots.txt` allows everything except
GPTBot. The scraper honours robots by default and requires an explicit flag to override.

## Inventory count

- `sitemaps/styles.xml`: **1290** `/style/<uuid>` URLs, each with `<lastmod>` and an
  `<image:loc>` thumbnail.
- `sitemaps/collections.xml`: 51 URLs. 10 are `/design-styles/<slug>` category listings,
  the rest are editorial `/ai-agents/*`, `/design-md/*`, `/examples/*` pages.
- The homepage markets "2,000+ AI-readable design systems".

Reconciliation: crawled 11 style pages, harvested 184 unique style ids from their
"More like this" sections, plus 24 ids from `/design-styles/dark-mode-websites`.
**Zero** ids fell outside the sitemap. The real public set looks like ~1290. The plan still
runs a related-graph BFS to closure so the true number is measured, not trusted.

## Per-style payload

One flight row carries a `result` object, ~33 KB of JSON:

- `meta` - url, siteName, extractedAt, durationMs, viewport, elementCount, telemetry
- `raw.colors` - token list with hex, oklch{l,c,h}, contexts, frequency, confidence,
  prominence, properties, per-`context/property` usageCounts; plus isDarkMode,
  colorfulness, and 20 contrastPairs with ratio and WCAG level
- `raw.shapes` - radii and shadows, each with contexts and frequency
- `raw.spacing` - value tokens with properties and propertyContextPairs, density, baseUnit
- `raw.gradients` - type, css value, component colors, frequency
- `raw.typography` - fonts (family, source system|google|custom, weights, frequency,
  fontFeatureSettings), detected scale (base, name, ratio, confidence), 30 type steps
- `designSystem` - the AI-authored layer: named colors with roles and groups, surfaces
  with levels and purposes, elevation per element, components with descriptions,
  typography roles, typeScale, spacing map with per-component radius, layout, imagery,
  dos, donts, northStar + northStarDetail, industry, theme, similar brands with reasons,
  description, customSections
- `screenshot` - url and thumbnail

The same page embeds ~20 style **cards**: id, url, siteName, screenshotUrl, thumbnailUrl,
iconUrl, previewVideoUrl, previewVideoPosterUrl, previewVideoDetailUrl,
previewVideoDetailPosterUrl, width/height for both videos, previewVideoDurationMs,
colorScheme, named colors with gradient strings, fonts, northStar, managementSignals,
createdAt. That covers the current style and its related set, so cards accumulate quickly.

## Flight format gotchas

The payload is newline-delimited `id:payload` rows in three shapes:

- `3:I[39756,["chunk.js",...],"ComponentName"]` client component manifest
- `19:T9df,<text>` length-prefixed text row, hex byte length
- `1c:T661,Quick Color Reference:...` same, and long strings are hoisted into these rows

Large strings inside the JSON are replaced by references like `"content":"$1c"`. The Go
parser must resolve `$<hex>` pointers against the text rows or fields silently come back
as literal `$1c`. Confirmed on `designSystem.customSections[0].content`.

## The artifact tabs are client-generated

DESIGN.md (Compact and Extended), Tailwind v4, CSS Variables and Design Tokens are built
in the browser from the same `result` object. They are not in the HTML and not fetched.

Located in two chunks, pinned under `reference/refero-js/` with sha256:

- `0jge9-tu1j3~c.js` (90 KB) - DESIGN.md generator. Section headings present verbatim:
  `## Quick Start`, `## Colors`, `## Surfaces`, `## Typography`, `## Spacing & Layout`,
  `## Components`, `## Elevation`, `## Layout`, `## Imagery`, `## Do's and Don'ts`,
  `## Similar Brands`, `## Tokens - Colors`, `## Tokens - Typography`,
  `## Tokens - Spacing & Shapes`, `### Type Scale`, `### Spacing Scale`,
  `### Border Radius`, `### Shadows`, `### CSS Custom Properties`, `### Tailwind v4`
- `0kxtp02n4q1qc.js` (57 KB) - CSS variable and Tailwind theme emitter, e.g.
  `--color-${slug(name)}: ${hex}` and the primary/secondary/background/foreground/muted/
  border role inference from luminance

Client JS also calls `/api/styles/<id>`, `/api/styles/<id>/event`,
`/api/styles/<id>/component-preview`. All under the robots-disallowed `/api/` prefix and
all unnecessary, since the page itself carries the data.

## Pagination

`?sort=newest` changes the server-rendered first page. `?page=N` and `?tab=x&page=N` are
ignored server-side; the grid pages client-side through `/api/`. Irrelevant, because the
sitemap plus the related-graph gives full coverage without touching `/api/`.

## Asset sizes (measured)

| kind | example bytes | x1290 |
|---|---|---|
| full screenshot jpg | 110,448 | ~140 MB |
| thumbnail jpg | 37,479 | ~48 MB |
| icon png | 6,888 | ~9 MB (often null) |
| preview video mp4 | 1,767,545 | ~2.3 GB |

## Identity

Site URLs are a mix of apex, www and short domains: `shade.inc`, `ballparkhq.com`,
`www.11x.ai`, `www.dimension.dev`, `dub.sh`, `midday.ai`. Normalisation needs
publicsuffix-aware eTLD+1 plus a preserved subdomain, per the chosen identity model.
