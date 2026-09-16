package render

import (
	"strings"

	"github.com/kanakkholwal/design-supply/internal/sites/refero"
)

// DesignMd renders the DESIGN.md export in either variant.
func DesignMd(res refero.Result, v Variant) string {
	if v == Compact {
		return designMdCompact(res)
	}
	return designMdExtended(res)
}

func designMdHeader(out *lines, res refero.Result) {
	ds := res.DesignSystem
	out.add("# " + res.Meta.SiteName + " — Style Reference")
	out.add("> " + ds.NorthStar)
	out.blank()
	out.add("**Theme:** " + ds.Theme)
	out.blank()
	out.add(ds.Description)
	out.blank()
}

func designMdCompact(res refero.Result) string {
	ds := res.DesignSystem
	var out lines
	designMdHeader(&out, res)

	out.add("## Colors")
	out.blank()
	out.add("| Name | Value | Role |")
	out.add("|------|-------|------|")
	for _, c := range ds.Colors {
		out.add("| " + c.Name + " | " + colorCell(c) + " | " + c.Role + " |")
	}

	out.blank()
	out.add("## Typography")
	out.blank()
	for _, tr := range ds.Typography {
		out.add("### " + tr.Family + " — " + tr.Role)
		writeTypographyBullets(&out, tr, false)
	}
	writeCompactTypeScale(&out, ds)

	out.add("## Spacing & Layout")
	out.blank()
	writeBaseUnitAndDensity(&out, res)
	writeLayoutBullets(&out, ds)
	out.blank()

	if ds.Spacing.Radius.Len() > 0 {
		out.add("### Border Radius")
		out.blank()
		for _, k := range ds.Spacing.Radius.Keys {
			out.add("- **" + k + ":** " + ds.Spacing.Radius.Get(k))
		}
		out.blank()
	}

	writeComponents(&out, ds)
	writeGuidelines(&out, ds)
	writeElevation(&out, ds)

	if len(ds.Surfaces) > 0 {
		out.add("## Surfaces")
		out.blank()
		for _, s := range ds.Surfaces {
			out.add("- **" + s.Name + "** (`" + s.Hex + "`) — " + s.Purpose)
		}
		out.blank()
	}

	writeProseSection(&out, "## Imagery", ds.Imagery)
	writeProseSection(&out, "## Layout", ds.Layout)
	writeSimilar(&out, ds)
	return out.join()
}

func designMdExtended(res refero.Result) string {
	ds, raw := res.DesignSystem, res.Raw
	var out lines
	designMdHeader(&out, res)

	out.add("## Tokens — Colors")
	out.blank()
	out.add("| Name | Value | Token | Role |")
	out.add("|------|-------|-------|------|")
	for _, c := range ds.Colors {
		token := "--color-" + slug(c.Name)
		out.add("| " + c.Name + " | " + colorCell(c) + " | `" + token + "` | " + c.Role + " |")
	}

	out.blank()
	out.add("## Tokens — Typography")
	out.blank()
	for _, tr := range ds.Typography {
		out.add("### " + tr.Family + " — " + tr.Role + " · `--font-" + slug(tr.Family) + "`")
		writeTypographyBullets(&out, tr, true)
	}

	if len(ds.TypeScale) > 0 {
		out.add("### Type Scale")
		out.blank()
		out.add("| Role | Size | Line Height | Letter Spacing | Token |")
		out.add("|------|------|-------------|----------------|-------|")
		for _, ts := range ds.TypeScale {
			out.add("| " + ts.Role + " | " + num(ts.Size) + "px | " + num(ts.LineHeight) +
				" | " + trackingCell(ts.LetterSpacing) + " | `--text-" + slug(ts.Role) + "` |")
		}
		out.blank()
	}

	out.add("## Tokens — Spacing & Shapes")
	out.blank()
	writeBaseUnitAndDensity(&out, res)

	if len(raw.Spacing.Tokens) > 0 {
		out.add("### Spacing Scale")
		out.blank()
		out.add("| Name | Value | Token |")
		out.add("|------|-------|-------|")
		for _, t := range filterSpacing(raw.Spacing.Tokens, raw.Spacing.BaseUnit) {
			n := num(t.Value)
			out.add("| " + n + " | " + n + "px | `--spacing-" + n + "` |")
		}
		out.blank()
	}

	if ds.Spacing.Radius.Len() > 0 {
		out.add("### Border Radius")
		out.blank()
		out.add("| Element | Value |")
		out.add("|---------|-------|")
		for _, k := range ds.Spacing.Radius.Keys {
			out.add("| " + k + " | " + ds.Spacing.Radius.Get(k) + " |")
		}
		out.blank()
	}

	if len(raw.Shapes.Shadows) > 0 {
		out.add("### Shadows")
		out.blank()
		out.add("| Name | Value | Token |")
		out.add("|------|-------|-------|")
		for _, s := range dedupeNames(shadowNames(raw.Shapes.Shadows)) {
			out.add("| " + s.Name + " | `" + truncate(s.Text, 60) + "` | `--shadow-" + s.Name + "` |")
		}
		out.blank()
	}

	out.add("### Layout")
	out.blank()
	writeLayoutBullets(&out, ds)
	out.blank()

	writeComponents(&out, ds)
	writeGuidelines(&out, ds)

	if len(ds.Surfaces) > 0 {
		out.add("## Surfaces")
		out.blank()
		out.add("| Level | Name | Value | Purpose |")
		out.add("|-------|------|-------|---------|")
		for _, s := range ds.Surfaces {
			out.add("| " + itoa(s.Level) + " | " + s.Name + " | `" + s.Hex + "` | " + s.Purpose + " |")
		}
		out.blank()
	}

	writeElevation(&out, ds)
	writeProseSection(&out, "## Imagery", ds.Imagery)
	writeProseSection(&out, "## Layout", ds.Layout)

	for _, cs := range ds.CustomSections {
		out.add("## " + cs.Title)
		out.blank()
		out.add(cs.Content)
		out.blank()
	}

	writeSimilar(&out, ds)

	out.add("## Quick Start")
	out.blank()
	out.add("### CSS Custom Properties")
	out.blank()
	out.add("```css")
	out.add(CSSVariables(res, Extended))
	out.add("```")
	out.blank()
	out.add("### Tailwind v4")
	out.blank()
	out.add("```css")
	out.add(TailwindTheme(res, Extended))
	out.add("```")
	out.blank()
	return out.join()
}

func colorCell(c refero.NamedColor) string {
	if c.Gradient != "" {
		return "`" + c.Gradient + "`"
	}
	return "`" + c.Hex + "`"
}

func trackingCell(ls float64) string {
	if ls != 0 {
		return num(ls) + "px"
	}
	return "—"
}

func writeTypographyBullets(out *lines, tr refero.TypographyRole, withRole bool) {
	if tr.Substitute != "" {
		out.add("- **Substitute:** " + tr.Substitute)
	}
	out.add("- **Weights:** " + tr.Weight)
	out.add("- **Sizes:** " + tr.Sizes)
	out.add("- **Line height:** " + tr.LineHeight)
	if tr.LetterSpacing != "" {
		out.add("- **Letter spacing:** " + tr.LetterSpacing)
	}
	if tr.FontFeatureSettings != "" {
		out.add("- **OpenType features:** `" + tr.FontFeatureSettings + "`")
	}
	if withRole {
		out.add("- **Role:** " + tr.Role)
	}
	out.blank()
}

func writeCompactTypeScale(out *lines, ds refero.DesignSystem) {
	if len(ds.TypeScale) == 0 {
		return
	}
	out.add("### Type Scale")
	out.blank()
	out.add("| Role | Size | Line Height | Letter Spacing |")
	out.add("|------|------|-------------|----------------|")
	for _, ts := range ds.TypeScale {
		out.add("| " + ts.Role + " | " + num(ts.Size) + "px | " + num(ts.LineHeight) +
			" | " + trackingCell(ts.LetterSpacing) + " |")
	}
	out.blank()
}

func writeBaseUnitAndDensity(out *lines, res refero.Result) {
	if b := res.Raw.Spacing.BaseUnit; b != nil && *b != 0 {
		out.add("**Base unit:** " + num(*b) + "px")
		out.blank()
	}
	out.add("**Density:** " + res.Raw.Spacing.Density)
	out.blank()
}

func writeLayoutBullets(out *lines, ds refero.DesignSystem) {
	if ds.Spacing.PageMaxWidth != "" {
		out.add("- **Page max-width:** " + ds.Spacing.PageMaxWidth)
	}
	if ds.Spacing.SectionGap != "" {
		out.add("- **Section gap:** " + ds.Spacing.SectionGap)
	}
	if ds.Spacing.CardPadding != "" {
		out.add("- **Card padding:** " + ds.Spacing.CardPadding)
	}
	if ds.Spacing.ElementGap != "" {
		out.add("- **Element gap:** " + ds.Spacing.ElementGap)
	}
}

func writeComponents(out *lines, ds refero.DesignSystem) {
	var described []refero.Component
	for _, c := range ds.Components {
		if strings.TrimSpace(c.Role) != "" && strings.TrimSpace(c.Description) != "" {
			described = append(described, c)
		}
	}
	if len(described) == 0 {
		return
	}
	out.add("## Components")
	out.blank()
	for _, c := range described {
		out.add("### " + c.Name)
		out.add("**Role:** " + c.Role)
		out.blank()
		out.add(c.Description)
		out.blank()
	}
}

func writeGuidelines(out *lines, ds refero.DesignSystem) {
	out.add("## Do's and Don'ts")
	out.blank()
	out.add("### Do")
	for _, d := range ds.Dos {
		out.add("- " + d)
	}
	out.blank()
	out.add("### Don't")
	for _, d := range ds.Donts {
		out.add("- " + d)
	}
	out.blank()
}

func writeElevation(out *lines, ds refero.DesignSystem) {
	if len(ds.Elevation) > 0 {
		out.add("## Elevation")
		out.blank()
		for _, e := range ds.Elevation {
			out.add("- **" + firstNonEmpty(e.Element, e.Name) + ":** `" + firstNonEmpty(e.Style, e.Shadow) + "`")
		}
		out.blank()
		return
	}
	writeProseSection(out, "## Elevation", ds.ElevationPhilosophy)
}

func writeSimilar(out *lines, ds refero.DesignSystem) {
	if len(ds.Similar) == 0 {
		return
	}
	out.add("## Similar Brands")
	out.blank()
	for _, s := range ds.Similar {
		if s.Business == "" {
			continue
		}
		out.add("- **" + s.Business + "** — " + s.Why)
	}
	out.blank()
}

func writeProseSection(out *lines, heading, body string) {
	if body == "" {
		return
	}
	out.add(heading)
	out.blank()
	out.add(body)
	out.blank()
}

func firstNonEmpty(a, b string) string {
	if a != "" {
		return a
	}
	return b
}

func truncate(s string, limit int) string {
	if len(s) <= limit {
		return s
	}
	return s[:limit-3] + "..."
}
