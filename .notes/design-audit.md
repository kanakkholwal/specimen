# Design audit: design-supply vs the Orbit system (Sept 2026)

Measured against [DESIGN.md](DESIGN.md). Contrast is WCAG 2.x relative luminance; CVD separation is
OKLab dE through protan, deutan and tritan simulation. Status: **Fixed** in this pass, **Open** still
to do, **N/A** where the finding cannot apply yet.

This project was a Go scraper with zero UI when the audit started, so section 1 records decisions
made against the Orbit baseline rather than defects found in shipped code. Sections 5 and later stay
Open until the routes they describe exist.

## 1. Colour and tokens

| # | Finding | Measured | Status |
| --- | --- | --- | --- |
| 1.1 | Accent needed that does not compete with 12,568 extracted colours on screen | indigo-700 `#4338ca` 7.90:1 white, 7.25:1 canvas; dark `#a5b4fc` 9.93:1 | Fixed |
| 1.2 | indigo-600 `#4f46e5` considered as the accent | 6.29:1 white, 5.77:1 canvas | Rejected: indigo-700 keeps 1.6 more points of headroom |
| 1.3 | Accent on its own 10% tint, the usual silent failure | 6.71:1 on white tint, 6.18:1 on canvas tint, 8.61:1 dark | Fixed |
| 1.4 | **Token model bug inherited from Orbit:** `--info` blue is perceptually identical to an indigo accent | dE 0.091 / 0.093 / 0.060 / 0.087, luminance 1.32:1 | Fixed: `--info` resolves to `--primary` |
| 1.5 | Retuning info to cyan, sky or teal instead of folding it in | tritan dE 0.085 / 0.007 / 0.077 | Rejected: tritanopia flattens the whole blue-cyan axis |
| 1.6 | No semantic token separates from the accent by luminance | success 1.52:1, destructive 1.33:1, warning 1.36:1 | Fixed by rule: glyph or word mandatory, documented |
| 1.7 | Success collides with the accent under tritanopia | 0.089 light, 0.048 dark | Accepted with the glyph rule; no green escapes it |
| 1.8 | Orbit's muted ink `#737373` fails the body floor on the canvas | 4.35:1 | Fixed: `#6b6b6b`, 4.89:1 canvas, 5.33:1 white |
| 1.9 | Orbit's brand panel is a raster whose brightest streak measures 3.12:1 against white text | n/a | Fixed: CSS panel from the accent ramp, 7.90:1 at its brightest point |
| 1.10 | Extracted swatches can be any colour, including the page background | n/a | Fixed by rule: swatches carry a `--border-strong` hairline and a printed hex |

## 2. Typography

| # | Finding | Status |
| --- | --- | --- |
| 2.1 | No type system existed | Fixed: Fontsource Google Sans, Inter, Source Code Pro; roles `text-caption` to `text-display-xl` |
| 2.2 | tailwind-merge silently drops role sizes next to colour classes | Fixed: `extendTailwindMerge` in `cn()` registers all ten roles |
| 2.3 | `text-7xl` and above could outgrow the scale | Fixed: clamped to 60px in `@theme` |
| 2.4 | Ad-hoc `text-[Npx]` | Fixed: none in the tree. The footer wordmark uses SVG text, so it scales without a font-size at all |
| 2.5 | Copied shadcn primitives used raw Tailwind sizes (`text-sm`, `text-xs`, `text-base`, `text-lg`), a second vocabulary next to the roles | Fixed: 24 occurrences converted to `text-body`, `text-caption`, `text-body-lg`, `text-body-xl` |

## 3. Surface, depth and motion

| # | Finding | Status |
| --- | --- | --- |
| 3.1 | Reduced motion | Fixed: global `prefers-reduced-motion` block neutralises animation and transition |
| 3.2 | Forced colours and high contrast | Fixed: `prefers-contrast: more` raises border and muted-ink tokens in both themes |
| 3.3 | Shadow at rest on cards | Fixed by rule and by `panel-card` carrying border only |
| 3.4 | JS-driven motion ignoring the OS setting | Open: applies when the count-up stat lands |

## 4. Toolchain and platform

| # | Finding | Measured | Status |
| --- | --- | --- | --- |
| 4.1 | TypeScript 7 with `svelte-check` | `svelte-check` 4.7.6 refuses TS 7 alone; needs TS 6 plus `@typescript/native` alias and `--tsgo` | Fixed: dual install, check reports 0 errors |
| 4.2 | Static bundling of the corpus | 15,480 artifact files + 4,914 image files = 20,394, over the 20,000 asset limit | Fixed: D1 for the index, R2 for artifacts and images |
| 4.3 | Index size for D1 | 311k rows, 64 MB, against a 10 GB ceiling | Fixed: full schema goes to D1 unslimmed |
| 4.4 | Preview videos | 2,076 objects, 2.6 MB average, 5.3 GB total | Deferred by decision: posters only, URLs retained for backfill |
| 4.5 | Em and en dashes in source and docs | n/a | Fixed: `scripts/check-dashes.mjs` gate, currently clean |
| 4.6 | Comment discipline | n/a | Fixed: `scripts/check-comments.mjs` gate over `src` and `scripts`, clean over 134 files |
| 4.7 | A third upstream chunk was pinned but contains no generator, so it would raise false drift alarms | 0 generator references | Fixed: unpinned, 117 KB removed |
| 4.8 | Repository hygiene | 203 files, 1.0 MB would be committed | Fixed: `.gitignore` covers node_modules, build output, `.wrangler`, env files, binaries and the 224 MB `server/data` |

## 5. Primitives

| # | Finding | Status |
| --- | --- | --- |
| 5.1 | Button variants on tokens, old names kept as aliases | Fixed: ported from Orbit. `default` near-black action, `primary` accent, `outline`, `ghost`, `ink`, `light`; `dark`, `secondary`, `link`, `brand`, `*_soft`, `raw` kept as aliases |
| 5.2 | Card, badge, input on tokens with no rest shadow | Fixed: 17 shadcn-svelte primitive groups ported (badge, button, checkbox, command, dialog, drawer, dropdown-menu, input, label, radio-group, select, separator, sheet, skeleton, sonner, switch, textarea, tooltip) |
| 5.3 | 40px minimum target, 44px on touch | Partly fixed: button `default` and `icon` are 40px, mobile nav rows 44px. Open: sweep once the app routes exist |
| 5.4 | Copied components referenced utilities the token layer did not define (`ease-snappy`, `shadow-subtle`, `shadow-brand`, `bg-brand`), which fail silently | Fixed: all four added to the token layer |
| 5.5 | Orbit-only components dragged in by the copy (`UploadArea`, `ui/sidebar`, Orbit navbar/footer/ErrorState) | Fixed: removed; navbar, footer, logo and theme toggle rewritten for this product |
| 5.6 | A button comment referenced `.band-dark`, a class this project does not have | Fixed |

## 6. Public pages

| # | Finding | Status |
| --- | --- | --- |
| 6.1 | Rail frame owning navbar, main, footer, rails | Fixed: `RailFrame` ported, `main` carries `id="main"` for the skip link |
| 6.6 | Footer card with spotlight wordmark | Fixed: SVG text with a pointer-tracked accent mask, hidden from assistive tech |
| 6.2 | Landing: tilted chip hero, accent second line, isometric illustration, real count-up stats | Open |
| 6.3 | Search launcher, bento, numbered FAQ on a bits-ui accordion, brand panel CTA | Open |
| 6.4 | Error and 404 states | Open |
| 6.5 | Boot loader | Open |

## 7. App routes

| # | Finding | Status |
| --- | --- | --- |
| 7.1 | `/explore` search, filter sheet, result grid | Open |
| 7.2 | `/style/[id]` tokens, exports, preview, copy actions | Open |
| 7.3 | `/sites/[origin]` grouping by publicsuffix identity | Open |
| 7.4 | `/collections/[slug]` for the 14 category pages | Open |

## 8. Security

| # | Finding | Status |
| --- | --- | --- |
| 8.1 | No write paths, no auth, no user data | N/A by design in this pass; read-only reader |
| 8.2 | Secrets in `NEXT_PUBLIC_`-style client vars | N/A: no secrets exist; D1 and R2 reach the worker through bindings |
| 8.3 | Admin surface | Open when it is asked for; schema and R2 keys are shaped so it needs no migration |

## Not verified

- Nothing has been rendered in a browser. The build compiles and type-checks; visual verification is
  outstanding.
- No SVG illustration exists yet, so the render-to-PNG check has not run.
- Lighthouse, axe and real-device checks have not run.
