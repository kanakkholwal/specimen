package render

import (
	"sort"
	"strings"

	"github.com/kanak/design-supply/internal/sites/refero"
)

type lines []string

func (l *lines) add(s string)     { *l = append(*l, s) }
func (l lines) join() string      { return strings.Join(l, "\n") }
func (l *lines) blank()           { *l = append(*l, "") }
func (l *lines) section(s string) { l.blank(); l.add(s) }

type sizeStep struct {
	size       float64
	lineHeight float64
	frequency  int
}

// CSSVariables renders the ":root" custom-property sheet.
func CSSVariables(res refero.Result, v Variant) string {
	ds, raw := res.DesignSystem, res.Raw
	var out lines
	out.add(":root {")
	out.add("  /* Colors */")
	for _, c := range ds.Colors {
		out.add("  --color-" + slug(c.Name) + ": " + c.Hex + ";")
		if c.Gradient != "" {
			out.add("  --gradient-" + slug(c.Name) + ": " + c.Gradient + ";")
		}
	}
	out.section("  /* Typography — Font Families */")
	writeFontFamilies(&out, ds.Typography)

	if v != Extended {
		out.add("}")
		return out.join()
	}

	switch {
	case len(ds.TypeScale) > 0:
		out.section("  /* Typography — Scale */")
		for _, ts := range ds.TypeScale {
			n := slug(ts.Role)
			out.add("  --text-" + n + ": " + num(ts.Size) + "px;")
			out.add("  --leading-" + n + ": " + num(ts.LineHeight) + ";")
			if ts.LetterSpacing != 0 {
				out.add("  --tracking-" + n + ": " + num(ts.LetterSpacing) + "px;")
			}
		}
	case len(raw.Typography.Steps) > 0:
		out.section("  /* Typography — Scale */")
		// Highest-frequency step wins per size here; the Tailwind generator keeps the first.
		for _, s := range dedupeNames(stepNames(collapseSteps(raw.Typography.Steps, true))) {
			out.add("  --text-" + s.Name + ": " + num(s.Value) + "px;")
			out.add("  --leading-" + s.Name + ": " + num(s.Extra) + ";")
		}
	}

	if weights := sortedWeights(raw.Typography.Fonts); len(weights) > 0 {
		out.section("  /* Typography — Weights */")
		for _, w := range weights {
			out.add("  --font-weight-" + weightName(w) + ": " + itoa(w) + ";")
		}
	}

	if len(raw.Spacing.Tokens) > 0 {
		out.section("  /* Spacing */")
		if raw.Spacing.BaseUnit != nil && *raw.Spacing.BaseUnit != 0 {
			out.add("  --spacing-unit: " + num(*raw.Spacing.BaseUnit) + "px;")
		}
		for _, t := range filterSpacing(raw.Spacing.Tokens, raw.Spacing.BaseUnit) {
			out.add("  --spacing-" + num(t.Value) + ": " + num(t.Value) + "px;")
		}
	}

	out.section("  /* Layout */")
	if ds.Spacing.PageMaxWidth != "" {
		out.add("  --page-max-width: " + ds.Spacing.PageMaxWidth + ";")
	}
	if ds.Spacing.SectionGap != "" {
		out.add("  --section-gap: " + ds.Spacing.SectionGap + ";")
	}
	if ds.Spacing.CardPadding != "" {
		out.add("  --card-padding: " + ds.Spacing.CardPadding + ";")
	}
	if ds.Spacing.ElementGap != "" {
		out.add("  --element-gap: " + ds.Spacing.ElementGap + ";")
	}

	if len(raw.Shapes.Radii) > 0 {
		out.section("  /* Border Radius */")
		for _, r := range dedupeNames(radiusNames(raw.Shapes.Radii)) {
			out.add("  --radius-" + r.Name + ": " + num(r.Value) + "px;")
		}
	}
	if ds.Spacing.Radius.Len() > 0 {
		out.section("  /* Named Radii */")
		for _, k := range ds.Spacing.Radius.Keys {
			out.add("  --radius-" + slug(k) + ": " + ds.Spacing.Radius.Get(k) + ";")
		}
	}
	if len(raw.Shapes.Shadows) > 0 {
		out.section("  /* Shadows */")
		for _, s := range dedupeNames(shadowNames(raw.Shapes.Shadows)) {
			out.add("  --shadow-" + s.Name + ": " + s.Text + ";")
		}
	}
	if len(ds.Surfaces) > 0 {
		out.section("  /* Surfaces */")
		for _, s := range ds.Surfaces {
			out.add("  --surface-" + slug(s.Name) + ": " + s.Hex + ";")
		}
	}
	out.add("}")
	return out.join()
}

// TailwindTheme renders the Tailwind v4 "@theme" block.
func TailwindTheme(res refero.Result, v Variant) string {
	ds, raw := res.DesignSystem, res.Raw
	var out lines
	out.add("@theme {")
	out.add("  /* Colors */")
	for _, c := range ds.Colors {
		out.add("  --color-" + slug(c.Name) + ": " + c.Hex + ";")
	}
	out.section("  /* Typography */")
	writeFontFamilies(&out, ds.Typography)

	if v != Extended {
		out.add("}")
		return out.join()
	}

	switch {
	case len(ds.TypeScale) > 0:
		out.section("  /* Typography — Scale */")
		for _, ts := range ds.TypeScale {
			n := slug(ts.Role)
			out.add("  --text-" + n + ": " + num(ts.Size) + "px;")
			out.add("  --leading-" + n + ": " + num(ts.LineHeight) + ";")
			if ts.LetterSpacing != 0 {
				out.add("  --tracking-" + n + ": " + num(ts.LetterSpacing) + "px;")
			}
		}
	case len(raw.Typography.Steps) > 0:
		out.section("  /* Typography — Scale */")
		for _, s := range dedupeNames(stepNames(collapseSteps(raw.Typography.Steps, false))) {
			out.add("  --text-" + s.Name + ": " + num(s.Value) + "px;")
			out.add("  --leading-" + s.Name + ": " + num(s.Extra) + ";")
		}
	}

	if len(raw.Spacing.Tokens) > 0 {
		out.section("  /* Spacing */")
		for _, t := range filterSpacing(raw.Spacing.Tokens, raw.Spacing.BaseUnit) {
			out.add("  --spacing-" + num(t.Value) + ": " + num(t.Value) + "px;")
		}
	}
	if len(raw.Shapes.Radii) > 0 {
		out.section("  /* Border Radius */")
		for _, r := range dedupeNames(radiusNames(raw.Shapes.Radii)) {
			out.add("  --radius-" + r.Name + ": " + num(r.Value) + "px;")
		}
	}
	if len(raw.Shapes.Shadows) > 0 {
		out.section("  /* Shadows */")
		for _, s := range dedupeNames(shadowNames(raw.Shapes.Shadows)) {
			out.add("  --shadow-" + s.Name + ": " + s.Text + ";")
		}
	}
	out.add("}")
	return out.join()
}

func writeFontFamilies(out *lines, roles []refero.TypographyRole) {
	seen := map[string]bool{}
	for _, r := range roles {
		key := slug(r.Family)
		if seen[key] {
			continue
		}
		seen[key] = true
		out.add("  --font-" + key + ": '" + r.Family + "', " + fontStack(r.Family) + ";")
	}
}

// collapseSteps keeps one entry per size. preferFrequent picks the most common step for a
// size; otherwise the first one encountered wins, which is what the Tailwind path does.
func collapseSteps(steps []refero.TypeStep, preferFrequent bool) []sizeStep {
	order := []float64{}
	bySize := map[float64]sizeStep{}
	for _, s := range steps {
		if s.Size < 10 {
			continue
		}
		cur, ok := bySize[s.Size]
		if !ok {
			order = append(order, s.Size)
			bySize[s.Size] = sizeStep{size: s.Size, lineHeight: s.LineHeight, frequency: s.Frequency}
			continue
		}
		if preferFrequent && s.Frequency > cur.frequency {
			bySize[s.Size] = sizeStep{size: s.Size, lineHeight: s.LineHeight, frequency: s.Frequency}
		}
	}
	out := make([]sizeStep, 0, len(order))
	for _, size := range order {
		out = append(out, bySize[size])
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].size < out[j].size })
	if len(out) > 12 {
		out = out[:12]
	}
	return out
}

func stepNames(steps []sizeStep) []named {
	out := make([]named, 0, len(steps))
	for _, s := range steps {
		out = append(out, named{Name: textSizeName(s.size), Value: s.size, Extra: s.lineHeight})
	}
	return out
}

func radiusNames(radii []refero.RadiusToken) []named {
	sorted := append([]refero.RadiusToken(nil), radii...)
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].Value < sorted[j].Value })
	out := make([]named, 0, len(sorted))
	for _, r := range sorted {
		out = append(out, named{Name: radiusName(r.Value), Value: r.Value})
	}
	return out
}

func shadowNames(shadows []refero.ShadowToken) []named {
	if len(shadows) > 15 {
		shadows = shadows[:15]
	}
	out := make([]named, 0, len(shadows))
	for _, s := range shadows {
		out = append(out, named{Name: shadowName(s.Value), Text: s.Value})
	}
	return out
}

func itoa(n int) string { return num(float64(n)) }
