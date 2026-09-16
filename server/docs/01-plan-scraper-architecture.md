# Plan 1: the generic scraper

> Pre-build plan, kept for the record. The system as built is described in
> [03-implementation.md](03-implementation.md), which supersedes this where they differ.

A reusable Go crawling engine. The engine knows nothing about Refero; a site adapter
supplies seeds, request shaping, parsing and the typed records to persist.

## Module

`module github.com/kanak/design-supply`, Go 1.25.

## Layout

```
cmd/supply/                 CLI entrypoint
internal/
  engine/
    engine.go               pipeline wiring, worker pool, graceful shutdown
    frontier.go             deduped, persisted work queue
    fetch.go                HTTP client, retries, conditional GET
    robots.go               robots.txt fetch, cache, enforcement
    limiter.go              per-host token bucket + adaptive throttle
    checkpoint.go           resume state
  adapter/
    adapter.go              the Adapter contract and shared types
    registry.go             name to adapter
    spec.go                 optional declarative field spec for simple sites
  sites/refero/             see Plan 2
  store/
    store.go                Store contract
    sqlite/                 schema, migrations, upserts, queries
    fsstore/                canonical on-disk layout
  normalize/                domain, colour, font, slug helpers
  media/                    asset fetcher, content-addressed, resumable
  obs/                      slog setup, run metrics, progress reporting
reference/refero-js/        pinned upstream chunks + SHA256SUMS
testdata/fixtures/          saved payloads for golden tests
```

## The adapter contract

This is the "site as an argument" part. Adding a site means implementing one interface,
not touching the engine.

```go
type Kind string // adapter-defined: "index", "detail", "listing", ...

type Target struct {
    URL   string
    Kind  Kind
    Depth int
    Meta  map[string]string // e.g. sitemap lastmod, referring style id
}

type Record struct {
    Kind string // "style", "site", "collection"
    ID   string // stable primary key
    Data any    // adapter-owned struct, handed to the Store
}

type Asset struct {
    URL      string
    Kind     string // "screenshot", "thumbnail", "icon", "video", "video_poster"
    OwnerID  string
    Role     string
    Width    int
    Height   int
    Download bool // false = record the URL only
}

type Result struct {
    Records  []Record
    Discover []Target
    Assets   []Asset
}

type Adapter interface {
    Name() string
    Seeds(ctx context.Context, f Fetcher) ([]Target, error)
    Classify(u *url.URL) (Kind, bool)          // false = not ours, drop it
    Request(t Target) (*http.Request, error)   // adapter injects headers, e.g. RSC: 1
    Parse(ctx context.Context, t Target, body []byte, hdr http.Header) (Result, error)
}
```

Derived artifacts are a separate, network-free contract so they can be regenerated offline:

```go
type Renderer interface {
    Name() string    // "design.compact.md"
    Version() string // bumped when output changes
    Render(rec Record) ([]byte, error)
}
```

Persistence is pluggable:

```go
type Store interface {
    UpsertRecords(ctx context.Context, recs []Record) error
    PutArtifact(ctx context.Context, ownerID, name, version string, body []byte) error
    RecordAsset(ctx context.Context, a Asset, local, sha string, n int64) error
    MarkFetched(ctx context.Context, url string, meta FetchMeta) error
    Frontier() FrontierStore
    Close() error
}
```

### Declarative option for simple sites

For a future site that is plain HTML and does not deserve a Go file, `adapter/spec.go`
loads a YAML field spec and builds a generic adapter over goquery selectors:

```yaml
name: example
seeds: ["https://example.com/sitemap.xml"]
detail:
  match: "^/item/"
  fields:
    title:  { css: "h1", text: true }
    price:  { css: ".price", attr: "data-value" }
    images: { css: "img.gallery", attr: "src", all: true }
```

Refero does not use this path. Its payload is deeply structured JSON, so it gets typed Go
structs: faster, checked at compile time, testable with golden files.

## Pipeline

Stages connected by buffered channels, each with a bounded worker pool, all under one
`errgroup.WithContext`.

```
seeds -> [frontier] -> fetchers(N) -> [raw] -> parsers(M) -> [records] -> writer(1)
                           ^                        |
                           +---- discovered --------+
                                                    -> [assets] -> media workers(K)
```

- Frontier dedupes by normalised URL and persists to SQLite, so a run resumes exactly.
- One writer goroutine owns the SQLite connection. Batched transactions, default 50 records
  per commit. Avoids lock contention entirely.
- Parsers are CPU work, so M defaults to `runtime.NumCPU()`, independent of N.
- Media workers use their own limiter because assets live on a different host.

## Politeness and rate limiting

- Per-host `golang.org/x/time/rate` limiter. Default 2 rps, burst 2, plus 0 to 250 ms
  jitter so requests do not arrive in lockstep.
- Adaptive throttle: on 429 or 503, honour `Retry-After` when present, otherwise halve the
  host rate and set a cooldown. Rate recovers 10% per 20 consecutive successes, never above
  the configured ceiling.
- Retries: max 4 attempts, exponential backoff 1s/2s/4s/8s with full jitter. Retry only on
  429, 5xx and transport errors. Any other 4xx is terminal and logged.
- Per-host circuit breaker: 20 consecutive failures pauses that host for 60s and logs
  loudly instead of hammering.
- robots.txt fetched once per host, cached for the run, enforced by default.
  `--ignore-robots` exists, prints a warning, and is not the default.
- Descriptive UA with a contact URL, configurable via `--user-agent`.

## HTTP client

One shared `*http.Client` with a tuned transport: `MaxIdleConnsPerHost` matched to worker
count, HTTP/2 on, explicit dial (5s), TLS handshake (5s), response header (10s) and overall
request (60s) timeouts. Accepts gzip and brotli. No cross-host redirect following without
adapter consent.

## Incremental re-runs

- Store `ETag` and `Last-Modified` per URL; send `If-None-Match` / `If-Modified-Since` on
  re-crawl and treat 304 as "unchanged, skip parse".
- Compare the sitemap `lastmod` against the stored `fetched_at` and skip untouched pages
  before spending a request at all.
- SHA-256 of the extracted payload detects real change even when headers lie, and drives
  artifact re-render only when needed.

## Resilience

- `signal.NotifyContext` on SIGINT/SIGTERM: stop accepting new work, drain in-flight, flush
  the writer, checkpoint, exit non-zero.
- Worker panics are recovered, logged with the target URL, and the item is marked failed
  rather than killing the run.
- A `fetch_log` table records every attempt with status, bytes, duration and error, so a run
  is auditable afterwards.
- Raw response bodies are archived gzipped under `data/raw/`. Parsing bugs are then fixed by
  re-running the parser offline instead of re-crawling.

## Observability

`log/slog` JSON to a file, plus a human line summary in the terminal. Per-run counters:
fetched, cached-304, parsed, records written, assets downloaded, bytes, retries, errors,
effective rps. One `runs` row per invocation.

## CLI

```
supply crawl  --site refero [--rps 2] [--workers 4] [--limit N] [--only <id>]
              [--resume] [--refresh] [--media icons,posters] [--dry-run]
supply render --site refero [--only <id>]   # offline, re-emits artifacts from stored JSON
supply media  --site refero --kinds icons,posters,thumbnails
supply verify --site refero --sample 20     # headless fidelity check, see Plan 2
supply export --format jsonl|zip
supply stats
```

`crawl` and `render` are deliberately separate. When a generator improves, re-rendering all
1290 styles is a local, network-free operation that takes seconds.

## Dependencies

Small and boring:

- `modernc.org/sqlite` (pure Go, no cgo, no toolchain pain on Windows)
- `golang.org/x/time/rate`, `golang.org/x/sync/errgroup`, `golang.org/x/net/publicsuffix`
- `github.com/temoto/robotstxt`
- `github.com/spf13/cobra`
- `github.com/chromedp/chromedp`, only in the `verify` path and behind a build tag, so the
  normal binary does not carry it

## Testing

- Golden-file tests for the flight parser and every renderer, driven by
  `testdata/fixtures/`. No network in unit tests.
- `httptest` server tests for the limiter, retry, 304 and circuit-breaker paths.
- A fixture-refresh command re-downloads the handful of fixtures on demand.
