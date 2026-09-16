# What was built

Authoritative description of the working system. `01` and `02` are the plans it was built
from; where they disagree with this file, this file is right.

## Commands

```
supply crawl  --site refero [--workers 4] [--rps 2] [--limit N] [--only <id>...]
              [--resume] [--refresh] [--dry-run] [--follow-related] [--collections]
              [--ignore-robots] [--user-agent ...] [--max-depth 6]
supply render [--only <id>...] [--skip-style-file]   # offline, no network
supply verify [--page <url>]                          # upstream generator drift check
supply stats
```

Global flags: `--data` (archive root, default `data`), `--db` (default `<data>/index.db`),
`--quiet`.

## Fetch strategy

`GET https://styles.refero.design/style/<uuid>` with header `RSC: 1`. That returns the
React flight payload directly, around 110 KB instead of 278 KB of HTML, from the same URL a
browser uses. No `/api/` path is touched, and robots.txt is enforced by default.

## Packages

| package | role |
|---|---|
| `internal/flight` | Next.js RSC wire-format parser, site-agnostic |
| `internal/adapter` | the `Adapter`, `Record`, `Blob`, `Result` contracts |
| `internal/engine` | worker pool, frontier, limiter, robots, retries, resume |
| `internal/normalize` | publicsuffix identity, slugs, path safety |
| `internal/sites/refero` | seeds, classify, extract, typed model |
| `internal/sites/refero/render` | the ported export generators |
| `internal/store` | composes SQLite and the file archive |
| `internal/store/sqlite` | schema, frontier, normalised corpus |
| `internal/store/fsstore` | atomic on-disk layout |
| `tools/fixtures` | regenerates the result fixture from a saved payload |
| `scripts/golden.mjs` | runs the real upstream JS to produce golden output |

## Flight parser notes

Two things the format does that a naive parser gets wrong, both covered by tests:

- Rows are **not** reliably newline-delimited. A JSON row ends where its JSON value ends,
  so the boundary comes from `json.Decoder.InputOffset`, not from `\n`.
- Text rows are length-prefixed as `<id>:T<hexlen>,<bytes>` and long strings elsewhere in
  the payload are replaced by `"$<hexid>"` pointers into them. Unresolved, fields such as
  `designSystem.customSections[0].content` come back as the literal `"$18"`.

`result.json` is re-emitted from the original bytes with **key order preserved**, because
the upstream generators iterate `designSystem.spacing.radius` with `Object.entries`. Round
tripping through a Go map would sort those keys and change the Border Radius table.

## A style page does not carry its own card

Each `/style/<uuid>` page embeds the extraction `result` plus about twenty **related**
style cards, and never its own. So icon URLs, video dimensions, `colorScheme` and
`createdAt` for style X only ever arrive from other pages that link to X. The database
tracks this with `styles.has_result` and `styles.has_card`; `supply stats` reports both,
and `supply render` rewrites `style.json` from the merged database state.

## Export artifacts

Ten files per style, all generated offline from `result.json`. Ported from the pinned
chunks and verified byte for byte against the upstream JavaScript by the golden tests in
`internal/sites/refero/render`.

| file | upstream format key |
|---|---|
| `DESIGN.compact.md`, `DESIGN.extended.md` | `design-md` |
| `variables.compact.css`, `variables.extended.css` | `css-variables` |
| `theme.compact.css`, `theme.extended.css` | `tailwind-config` |
| `tokens.compact.json`, `tokens.extended.json` | `tokens-json` |
| `design-system.json` | `design-json` |
| `prompt.txt` | `prompt` |

The plans said four or five formats. The site's own `FORMAT_META` table lists six, and four
of them take a compact/extended variant, hence ten files.

### How fidelity is proven

`scripts/golden.mjs` loads the pinned Turbopack chunks in a Node VM with a stubbed module
context, pulls out the real `generateDesignMd`, `generateCssVariables`,
`generateTailwindConfig`, `generateTokensJson`, `generateDesignJson` and
`generatePromptSnapshotText`, and runs them against the fixture to produce
`testdata/golden/shade/`. The Go renderers are then diffed against that output. No headless
browser is involved.

Regenerate after re-pinning chunks:

```
go run ./tools/fixtures testdata/fixtures/style-shade.rsc.txt testdata/fixtures/style-shade.result.json
node scripts/golden.mjs reference/refero-js/0jge9-tu1j3~c.js reference/refero-js/0kxtp02n4q1qc.js \
  testdata/fixtures/style-shade.result.json testdata/golden/shade
go test ./...
```

`supply verify` re-fetches the chunk files a live style page references and compares them
against `reference/refero-js/SHA256SUMS`. A changed hash means the generators may have
moved and the goldens should be regenerated.

## On-disk layout

```
data/
  sites/<origin>/<uuid>/
    style.json            identity plus every media URL
    result.json           full extraction payload, upstream key order
    DESIGN.compact.md  DESIGN.extended.md
    variables.compact.css  variables.extended.css
    theme.compact.css  theme.extended.css
    tokens.compact.json  tokens.extended.json
    design-system.json
    prompt.txt
  raw/<uuid>.flight.txt.gz   verbatim response, for re-parsing without re-crawling
  index.db
```

Files are written to a temporary name and renamed, so an interrupted run never leaves a
half-written artifact.

## Media

**Nothing is downloaded.** Every media URL is still recorded in the `media` table with its
kind, width, height and duration where known, linked to its style through `style_media`,
and mirrored into `style.json`. Rows carry `downloaded = 0` and a null `local_path`, so
adding a downloader later is additive and needs no re-crawl. The CDN can be hotlinked from
day one.

Kinds recorded: `screenshot`, `thumbnail`, `icon`, `video` (preview and detail),
`video_poster` (both posters).

## Deduplication

`colors`, `fonts`, `shadows`, `radii`, `spacings` and `gradients` hold one row per distinct
value across the entire corpus; styles reference them through join tables that carry the
per-style facts (frequency, context, role, prominence). `#ffffff` is one row no matter how
many styles use it.

Identity: `sites` is keyed by normalised origin with `www.` stripped and any other
subdomain preserved, plus the publicsuffix `etld1` for grouping. `dub.sh` and `dub.co` stay
separate sites; `app.foo.com` and `foo.com` stay separate but share an `etld1`.

## Politeness

Per-host token bucket at 2 rps with a 250 ms jitter, adaptive halving on 429 or 503 with
`Retry-After` honoured, 10% recovery per twenty clean responses, four attempts with
exponential backoff and full jitter, a circuit breaker at twenty consecutive failures, and
robots.txt enforced unless `--ignore-robots` is passed. The User-Agent identifies the
crawler rather than impersonating a browser.

## Resuming

The frontier lives in SQLite. `--resume` returns stranded `leased` rows to `pending` and
continues. ETag and Last-Modified are stored per URL and replayed as conditional GETs, so a
second run mostly collects 304s.

## Tests

- `internal/flight`: RSC and HTML inputs agree; length-prefixed rows recovered; `$ref`
  resolution; card discovery.
- `internal/sites/refero`: extraction, plus `TestResultModelIsLossless`, which round-trips
  the payload through the typed model and fails naming any field the model drops. It caught
  three real fields during the build.
- `internal/sites/refero/render`: all ten artifacts diffed against upstream JavaScript
  output, reporting the first differing line.
- `internal/engine`: retry on 5xx, 304 short-circuit, robots enforcement, exact `--limit`,
  adaptive throttle, circuit breaker.
- `internal/normalize`: domain identity and slugging.
