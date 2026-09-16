# design-supply Design System

A reading surface for 1,290 extracted design systems. White cards on a light gray canvas between
dashed column rails. Structure comes from 1px hairlines, generous spacing and one indigo accent.

> **Borders, not depth.** A container is a 1px hairline and a radius. If a surface needs a shadow
> to read, it needs better spacing instead.

This ports the [Orbit system](../../orbit/DESIGN.md) (layout, neutrals, dark palette, radii, type
scale, card grammar). Three things are this project's own:

| Exempt | Value | Why |
| --- | --- | --- |
| Brand accent | indigo, `#4338ca` / `#a5b4fc` | Orbit is emerald, college-ecosystem is teal. |
| `--info` | folded into the accent | A separate blue is perceptually identical to indigo. See below. |
| Brand panel | pure CSS, derived from the accent | No raster asset, and white text clears 4.5:1 everywhere on it. |

Tokens live in [src/app.css](../src/app.css). Never hardcode a hex in a component.

---

## The constraint that shapes everything

This app puts **12,568 distinct extracted colours** on screen. The chrome is not the subject; the
data is. So the accent is chosen to recede, not to compete, and no page element may use colour as
its only signal, because any swatch on the page can sit next to any token.

---

## Colour

### Surfaces and ink

| Role | Light | Dark | Token |
| --- | --- | --- | --- |
| Page canvas | `#f5f5f5` | `#0a0a0a` | `--canvas` |
| App background | `#ffffff` | `#0a0a0a` | `--background` |
| Card | `#ffffff` | `#171717` | `--card` |
| Muted fill, hover fill | `#f5f5f5` | `#262626` | `--muted`, `--paper` |
| Hairline | `#e5e5e5` | `rgb(255 255 255 / .10)` | `--border` |
| Strong border | `#d4d4d4` | `rgb(255 255 255 / .16)` | `--border-strong` |
| Ink | `#0a0a0a` | `#fafafa` | `--foreground` |
| Muted ink | `#6b6b6b` | `#a1a1a1` | `--muted-foreground` |
| Placeholder | `#737373` | `#737373` | `--placeholder` |

**Muted ink is `#6b6b6b`, not `#737373`.** `#737373` measures 4.35:1 on the gray canvas, under the
4.5:1 body floor. `#6b6b6b` measures 4.89:1 on the canvas and 5.33:1 on white.

**Cards on public pages** use `panel-card`: white in light, `--background` (`#0a0a0a`) in dark, so a
dark card is told apart from the canvas by its hairline alone.

### The accent

| Token | Light | Dark |
| --- | --- | --- |
| `--primary` | `#4338ca` (indigo-700) | `#a5b4fc` (indigo-300) |
| `--primary-active` | `#3730a3` | `#818cf8` |
| `--primary-foreground` | `#ffffff` | `#0a0a0a` |

Measured:

| Pair | Ratio |
| --- | --- |
| Light accent on white | 7.90:1 |
| Light accent on canvas `#f5f5f5` | 7.25:1 |
| Light accent on its own 10% tint over white (`#ecebfa`) | 6.71:1 |
| Light accent on its own 10% tint over canvas (`#e3e2f1`) | 6.18:1 |
| White on the light accent | 7.90:1 |
| Dark accent on `#0a0a0a` | 9.93:1 |
| Dark accent on card `#171717` | 8.99:1 |
| Dark accent on its 10% dark tint (`#1a1b22`) | 8.61:1 |

**Not indigo-600.** `#4f46e5` measures 6.29:1 on white, which passes, but indigo-700 buys another
1.6 points of headroom against the arbitrary colours it sits beside.

`--primary` is for: the accent words in a headline, the primary brand action, links in body copy,
active and selected states, focus rings, and progress fills. Never a large flat background fill.

### Why there is no separate `--info`

An indigo accent and a conventional info blue are the same colour to a large share of viewers:

| Pair | normal | protan | deutan | tritan | Luminance |
| --- | --- | --- | --- | --- | --- |
| `#4338ca` vs `#0060c9` | 0.091 | 0.093 | 0.060 | 0.087 | 1.32:1 |

Every row is under the 0.10 OKLab dE floor, and the luminance escape clause (3:1) fails at 1.32:1.
Retuning info to cyan `#0e7490`, sky `#075985` or teal `#0f766e` fixes normal, protan and deutan but
still collapses under tritanopia (0.085, 0.007, 0.077), because tritanopia flattens the entire
blue-cyan axis. **No cool hue can separate from an indigo accent under all vision types.**

So `--info` resolves to `--primary`. One cool token, one job. Informational states are told apart by
glyph and word, never hue.

### Semantic colour

| Token | Light | Dark |
| --- | --- | --- |
| `--destructive` | `#c0242a` | `#ff6b61` |
| `--success` | `#0a7d47` | `#30d158` |
| `--warning` | `#8a5c00` | `#ff9f0a` |
| `--info` | `var(--primary)` | `var(--primary)` |

**State is never colour alone.** Separation from the accent, measured:

| Pair | normal | deutan | tritan | Luminance |
| --- | --- | --- | --- | --- |
| accent vs success (light) | 0.307 | 0.267 | **0.089** | 1.52:1 |
| accent vs destructive (light) | 0.336 | 0.324 | 0.375 | 1.33:1 |
| accent vs warning (light) | 0.322 | 0.327 | 0.330 | 1.36:1 |
| accent vs success (dark) | 0.286 | 0.220 | **0.048** | 1.01:1 |

Success collides with the accent under tritanopia in both themes, and **no** semantic token reaches
3:1 luminance separation from the accent. Every status therefore carries a glyph or a word. Hue is
corroborating evidence, never the signal.

### Brand panel

`panel-brand` is CSS only, built from the accent ramp:

| Layer | Value |
| --- | --- |
| Base | `--brand-panel-base` `#312e81` (indigo-900) |
| Deep corner | `--brand-panel-deep` `#1e1b4b` (indigo-950) |
| Lift | `--brand-panel-lift` `#4338ca` (indigo-700) |
| Grid | `rgb(255 255 255 / 0.06)` at 80px |

White text measures **7.90:1 on the brightest point** of the panel (`#4338ca`), so unlike a raster
gradient there is no "keep text away from the light streak" rule. Any text size is safe.

### Contrast floors

| Thing | Floor |
| --- | --- |
| Body text on its surface | 4.5:1 |
| Focus ring, control boundary, meaningful icon | 3:1 |
| Two controls distinguished by colour | 3:1 luminance **or** >0.10 OKLab dE under CVD |

**Never fade a text token with an opacity modifier.** Tints on fills (`bg-primary/10`) are fine.

### Extracted colour is content, not chrome

Swatches rendered from the corpus are data. They are drawn as a filled chip with a 1px
`--border-strong` hairline and their hex printed beside them in `--font-mono`, because a swatch may
be any colour including the page background. Never rely on the swatch fill alone to be visible.

---

## Typography

Self-hosted through Fontsource, imported once in `src/routes/+layout.svelte`.

| Token | Face | Package |
| --- | --- | --- |
| `--font-heading` / `font-display` | Google Sans Variable | `@fontsource-variable/google-sans` |
| `--font-sans` | Inter Variable | `@fontsource-variable/inter` |
| `--font-mono` | Source Code Pro Variable | `@fontsource-variable/source-code-pro` |

Headings are weight **500**. h1 and h2 carry `-0.02em` tracking. Mono carries hex values, token
names, CSS output and font stacks.

### Scale

| Token | Size / line | Role |
| --- | --- | --- |
| `text-caption` | 12 / 16 | Chips, meta, tags, footer legal |
| `text-body` | 14 / 20 | Body copy, card copy, buttons |
| `text-body-lg` | 16 / 24 | Card titles (h3), FAQ questions, ledes from `md` |
| `text-body-xl` | 18 / 28 | Rare emphasis |
| `text-subheading` | 20 / 28 | Section titles inside a card |
| `text-heading-sm` | 24 / 32 | Prose h2, style detail section headings |
| `text-heading` | 30 / 36 | Prose h2 from `sm` |
| `text-heading-lg` | 36 / 40 | Page h1 (mobile), split-section h2 |
| `text-display` | 48 / 1 | Page h1 from `md`, brand panel h2 |
| `text-display-xl` | 60 / 1 | Landing hero h1 and closing CTA from `lg`, nothing else |

`text-7xl` and above clamp to 60px. **No ad-hoc `text-[Npx]`.**

### tailwind-merge

`cn()` uses `extendTailwindMerge` from [src/lib/utils.ts](../src/lib/utils.ts), which registers every
`text-*` role. Without it, tailwind-merge reads `text-body` as a colour and drops it when a
`text-foreground` follows. Add any new size token there.

---

## Shape

| Utility | Value | Use |
| --- | --- | --- |
| `rounded-sm` / `rounded-xs` | 6px | Swatch chips, keycaps |
| `rounded-md` | 8px | Buttons, controls, tilted chips |
| `rounded-lg` | 10px | List rows, nav items |
| `rounded-xl` | 14px | Inset visuals, search inputs |
| `rounded-2xl` | 18px | Cards (`panel-card`), FAQ cards, panels |
| `rounded-3xl` | 22px | Bento cards from `md`, brand panels, footer card |
| `rounded-full` | 9999px | Pills, status tags |

Nest radii inward: a 14px visual inside an 18 to 22px card, never the reverse.

---

## Elevation

Borders define containers.

| Use | Token |
| --- | --- |
| Buttons | `shadow-xs` |
| Inset visuals inside a bento card | `shadow-sm` |
| Bento card hover (pointer feedback only) | `shadow-lg` |
| Floating overlays | `shadow-md` / `shadow-lg` |

**Cards carry no shadow at rest. Inputs carry no shadow. No hover lift, no glow.**

---

## Layout

### The rail frame

Every public page renders inside `RailFrame`, which owns the navbar, `<main>`, the footer and the
rails. Pages never import the navbar or footer directly.

- **Column** (`rail-column`): `min(95vw, 1440px)`, `min(90vw, 1440px)` from `md`, `min(85vw, 1800px)`
  from 1536px.
- **Rails:** two fixed 2px dashed lines on the column edges in `--border`.
- **Rows:** `RailRow` is a full-bleed 2px dashed rule, then the column with `p-3 sm:p-6`. The first
  row passes `divider={false}`. `label` names the section for assistive tech.

### Page primitives (`src/components/site`)

| Component | Shape |
| --- | --- |
| `PageHero` | Tilted chip, h1 with an accent second line, lede, actions, optional aside |
| `SplitSection` | Sticky title + accent + description left, content right |
| `BrandPanel` | `panel-brand` 22px card, centred white h2, body, ink and light actions |
| `FaqList` | Numbered 18px cards on a bits-ui accordion, accent index, one open |

---

## Motion

One easing (`--ease-craft`), one hover duration (150ms). The global
`prefers-reduced-motion` block in `app.css` neutralises every animation and transition. Any JS-driven
motion (the stat count-up) checks `window.matchMedia('(prefers-reduced-motion: reduce)')` and renders
the final value immediately.

---

## Data honesty

Stats on the landing page are read from the database at build or request time, never typed in. The
corpus is 1,290 styles measured by a crawl that closed the related-style graph and found zero ids
outside the sitemap. The site does not claim "2,000+" because that number is not true of what is
stored here. No testimonials, no invented logos, no fake counters.

Empty states say what is missing and what to do, and never invent a placeholder record.

---

## Accessibility rules

- One `h1` per page.
- Skip link first in the layout, visible on focus.
- Keyboard focus is `outline-2 outline-offset-2 outline-ring`, never removed.
- Interactive targets are at least 40px, 44px on touch surfaces.
- Status is glyph or word plus colour, never colour alone.
- Overlays: right sheet on desktop, bottom drawer on mobile (`vaul-svelte`).
