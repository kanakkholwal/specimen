# Specimen

A collection of design standards: colour tokens, type scales, spacing, components and
ready-to-copy DESIGN.md, CSS variables, Tailwind themes and design tokens.

SvelteKit on Cloudflare Workers, reading D1 and R2.

## Setup

```bash
bun install
```

## Run

```bash
bun run dev          # reads dist/specimen.db directly
bun run dev:worker   # runs the real Worker against local D1
```

`bun run dev` is the everyday one. Use `dev:worker` before a deploy to exercise the D1 binding.

## Index

The index is a SQLite bundle at `dist/specimen.db` for local work, and D1 in production. It is
loaded out of band and never committed.

```bash
bun run db:local     # load the index into local D1
bun run db:remote    # load the index into production D1
```

Local D1 is keyed by database name, so renaming it in `wrangler.jsonc` gives you a fresh empty
one. A Worker reporting `no such table` means exactly that. Re-run `db:local`.

## Before pushing

`bun run gate` runs everything CI runs, in the same order, so a green gate means a green pipeline.

```bash
bun run gate
```

Install the pre-push hook once and it runs itself:

```bash
bun run hooks
```

Individual checks:

```bash
bun run check    # svelte-check on TypeScript 7
bun test src     # unit tests
bun run lint     # biome
bun run build
```

The build step writes where the dev worker keeps its assets, so stop `bun run dev` if the gate
reports EBUSY.

## Layout

```
src/
  lib/server/db.ts   every database query
  components/        site primitives, application components, ui
  routes/            landing, explore, sites, collections, style detail
```

Design system and audit live in `.notes/`, local only.
