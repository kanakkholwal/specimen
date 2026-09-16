# Specimen

An archive of design systems extracted from real product websites: colour tokens, type scales,
spacing, components, and ready-to-copy DESIGN.md, CSS variables, Tailwind themes and design tokens.

- **Web app**: SvelteKit on Cloudflare Workers, reading D1 and R2.
- **Scraper**: Go, in [`server/`](server), run locally. Its output is never committed.

## Why nothing heavy is in git

The crawl produces about 224 MB and the media about 350 MB. None of it belongs in history, so
`server/data/`, `dist/` and `.wrangler/` are ignored. You regenerate them with the commands below.
What *is* committed: source, the two pinned upstream JS chunks used to verify the export
generators, and small test fixtures. The whole repository is under 1 MB.

## One-time setup

```bash
bun install                 # web app
cd server && go build ./... # scraper
choco install webp          # cwebp, for the image pipeline
```

## The pipeline

Four steps. Each is idempotent and safe to re-run.

### 1. Crawl

Fetches every style page, parses the embedded extraction payload, and writes both a normalised
SQLite index and per-style files.

```bash
cd server
go run ./cmd/supply crawl --site refero --workers 4 --rps 2
```

About 11 minutes for 1,290 styles at the default polite rate. Resumable with `--resume`; a second
run mostly collects 304s. Output lands in `server/data/`:

```
server/data/
  index.db                     normalised index, 311k rows
  sites/<origin>/<uuid>/       result.json plus ten generated exports
  raw/<uuid>.flight.txt.gz     verbatim response, so parser fixes need no re-crawl
```

### 2. Render (optional)

Regenerates the ten export artifacts from stored results without touching the network. Run it after
changing a generator.

```bash
go run ./cmd/supply render
```

### 3. Export the index

Emits SQL for D1, including the artifact bodies so the app can serve copy actions.

```bash
go run ./cmd/supply export --out ../dist/specimen.sql
cd ..
bunx wrangler d1 execute specimen --local  --file=dist/specimen.sql   # local dev
bunx wrangler d1 execute specimen --remote --file=dist/specimen.sql   # production
```

About 135 MB of SQL. Statements are capped at 32 KB and artifact bodies are chunked, because D1
rejects anything larger.

### 4. Transcode media

Downloads the source images, encodes WebP with `cwebp`, and writes a manifest. Keys are the source
content hash, so an unchanged image is never re-encoded or re-uploaded.

```bash
cd server
go run ./cmd/supply media --out ../dist/media --manifest ../dist/media-manifest.json
```

Roughly 4,900 images. Videos are deliberately skipped: 2,076 objects at 5.3 GB, and the poster
frames already give a still. Their URLs stay in the database, so a later backfill needs no re-crawl.

Upload to R2:

```bash
cd dist/media && bunx wrangler r2 object put specimen-assets/<key> --file=<key> --remote
```

## Verify

```bash
cd server
go test ./...                       # includes byte-for-byte export tests
go run ./cmd/supply verify          # are the upstream generators unchanged?
go run ./cmd/supply stats           # what the archive holds
```

`verify` re-fetches the JavaScript chunks a live style page references and compares them against
`server/reference/refero-js/SHA256SUMS`. A changed hash means the upstream generators may have moved
and the golden files should be regenerated.

## Develop

```bash
bun run dev      # needs step 3 run with --local at least once
bun run check    # svelte-check on TypeScript 7
bun run gate     # comment, dash, lint and type gates
bun run build
```

## Layout

```
src/                 SvelteKit app
  lib/server/db.ts   every D1 query
  components/        site primitives, application components, shadcn-svelte ui
server/              Go scraper, exporter and media pipeline
  cmd/supply/        crawl, render, export, media, verify, stats
  internal/          flight parser, refero adapter, export generators, store
.notes/              DESIGN.md, design-audit.md
  docs/              site recon and the original build plans
```

Design system: [.notes/DESIGN.md](.notes/DESIGN.md). Audit: [.notes/design-audit.md](.notes/design-audit.md).
How the scraper was built and what the site recon found: [.notes/docs/](.notes/docs/).
