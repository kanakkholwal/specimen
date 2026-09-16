# design-supply

A generic Go scraper plus a Refero Styles adapter, building a private, queryable archive of
AI-readable design systems.

```
go build ./cmd/supply
supply crawl --site refero      # fetch and index
supply render                   # regenerate export artifacts offline
supply verify                   # check upstream generators have not moved
supply stats                    # what the archive holds
```

Each style lands in `data/sites/<origin>/<uuid>/` with the full extraction payload and ten
export artifacts (DESIGN.md, CSS variables, Tailwind theme, design tokens, design JSON,
prompt snapshot) that are byte-identical to what styles.refero.design generates in the
browser. No media bytes are downloaded; every media URL is recorded.

## Docs

- [docs/03-implementation.md](docs/03-implementation.md) - what exists, start here
- [docs/00-findings.md](docs/00-findings.md) - site recon, measured against the live site
- [docs/01-plan-scraper-architecture.md](docs/01-plan-scraper-architecture.md) - the generic engine plan
- [docs/02-plan-refero-adapter.md](docs/02-plan-refero-adapter.md) - extraction and schema plan

`reference/refero-js/` pins the upstream generator chunks with their hashes.
`testdata/golden/` holds output produced by running those chunks, which the Go renderers are
diffed against.
