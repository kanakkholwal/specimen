package render

import (
	"sort"

	"github.com/kanakkholwal/design-supply/internal/sites/refero"
)

// TokensJSON renders the Design Tokens export in DTCG shape.
func TokensJSON(res refero.Result, v Variant) (string, error) {
	ds, raw := res.DesignSystem, res.Raw
	root := newObj()

	colors := newObj()
	for _, c := range ds.Colors {
		colors.set(slug(c.Name), token(c.Hex, "color", c.Name+" — "+c.Role))
	}
	root.set("color", colors)

	fonts := newObj()
	for _, tr := range ds.Typography {
		fonts.set(slug(tr.Family), token(tr.Family, "fontFamily", tr.Role))
	}
	root.set("font", fonts)

	if v == Extended {
		if len(raw.Typography.Steps) > 0 {
			root.set("typography", typographyTokens(raw.Typography.Steps))
		}
		if len(raw.Spacing.Tokens) > 0 {
			spacing := newObj()
			for _, t := range filterSpacing(raw.Spacing.Tokens, raw.Spacing.BaseUnit) {
				n := num(t.Value)
				spacing.set(n, token(n+"px", "dimension", "Spacing "+n+"px"))
			}
			if b := raw.Spacing.BaseUnit; b != nil && *b != 0 {
				spacing.set("unit", token(num(*b)+"px", "dimension", "Base spacing unit"))
			}
			root.set("spacing", spacing)
		}
		if len(raw.Shapes.Radii) > 0 {
			radii := newObj()
			for _, r := range dedupeNames(radiusNames(raw.Shapes.Radii)) {
				radii.set(r.Name, token(num(r.Value)+"px", "dimension", "Border radius "+r.Name))
			}
			root.set("radius", radii)
		}
		if len(raw.Shapes.Shadows) > 0 {
			shadows := newObj()
			for _, s := range dedupeNames(shadowNames(raw.Shapes.Shadows)) {
				shadows.set(s.Name, token(s.Text, "shadow", "Shadow elevation "+s.Name))
			}
			root.set("shadow", shadows)
		}
		if len(ds.Surfaces) > 0 {
			surfaces := newObj()
			for _, s := range ds.Surfaces {
				desc := "Surface level " + itoa(s.Level) + ": " + s.Purpose
				surfaces.set(slug(s.Name), token(s.Hex, "color", desc))
			}
			root.set("surface", surfaces)
		}
	}

	extraction := newObj()
	extraction.set("url", res.Meta.URL)
	extraction.set("siteName", res.Meta.SiteName)
	extraction.set("extractedAt", res.Meta.ExtractedAt)
	extraction.set("variant", string(v))
	extensions := newObj()
	extensions.set("com.refero.extraction", extraction)
	root.set("$extensions", extensions)

	return root.marshal()
}

func token(value any, typ, description string) *jsonObj {
	o := newObj()
	o.set("$value", value)
	o.set("$type", typ)
	o.set("$description", description)
	return o
}

// typographyTokens keeps every measured step, unlike the CSS generators which collapse
// to one entry per size.
func typographyTokens(steps []refero.TypeStep) *jsonObj {
	sorted := append([]refero.TypeStep(nil), steps...)
	sort.SliceStable(sorted, func(i, j int) bool { return sorted[i].Size < sorted[j].Size })

	names := make([]named, 0, len(sorted))
	for _, s := range sorted {
		names = append(names, named{Name: textSizeName(s.Size)})
	}
	names = dedupeNames(names)

	out := newObj()
	for i, s := range sorted {
		name := names[i].Name
		value := newObj()
		value.set("fontFamily", s.Family)
		value.set("fontSize", num(s.Size)+"px")
		value.set("fontWeight", s.Weight)
		value.set("lineHeight", s.LineHeight)

		entry := newObj()
		entry.set("$value", value)
		entry.set("$type", "typography")
		entry.set("$description", "Typography step "+name+" at "+num(s.Size)+"px")
		out.set(name, entry)
	}
	return out
}
