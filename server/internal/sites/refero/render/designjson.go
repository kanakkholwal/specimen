package render

import (
	"encoding/json"
	"strings"

	"github.com/kanak/design-supply/internal/sites/refero"
)

// DesignSystemJSON renders the Design JSON export. Sections that are copied through
// unchanged are spliced from the archived result bytes so their key order survives.
func DesignSystemJSON(res refero.Result, rawResult []byte) (string, error) {
	ds := res.DesignSystem
	dsRaw, err := designSystemFields(rawResult)
	if err != nil {
		return "", err
	}

	root := newObj()

	meta := newObj()
	meta.set("siteName", res.Meta.SiteName)
	meta.set("url", res.Meta.URL)
	meta.set("extractedAt", res.Meta.ExtractedAt)
	meta.set("thumbnailUrl", res.Screenshot.Thumbnail)
	meta.set("screenshotUrls", screenshotURLs(res, rawResult))
	root.set("meta", meta)

	root.set("description", ds.Description)
	root.set("northStar", ds.NorthStar)
	root.set("theme", ds.Theme)
	if ds.Industry != "" {
		root.set("industry", ds.Industry)
	}
	setRaw(root, "colors", dsRaw)
	root.set("typography", typographyEntries(res))
	setRaw(root, "typeScale", dsRaw)

	spacing := newObj()
	if b := res.Raw.Spacing.BaseUnit; b != nil {
		spacing.set("baseUnit", *b)
	} else {
		spacing.set("baseUnit", nil)
	}
	spacing.set("density", res.Raw.Spacing.Density)
	spacing.set("pageMaxWidth", ds.Spacing.PageMaxWidth)
	spacing.set("sectionGap", ds.Spacing.SectionGap)
	spacing.set("cardPadding", ds.Spacing.CardPadding)
	spacing.set("elementGap", ds.Spacing.ElementGap)
	if r, ok := rawField(dsRaw["spacing"], "radius"); ok {
		spacing.set("radius", r)
	}
	root.set("spacing", spacing)

	setRaw(root, "components", dsRaw)
	setRaw(root, "dos", dsRaw)
	setRaw(root, "donts", dsRaw)
	setRaw(root, "surfaces", dsRaw)
	setRaw(root, "elevation", dsRaw)
	if ds.ElevationPhilosophy != "" {
		root.set("elevationPhilosophy", ds.ElevationPhilosophy)
	}
	if ds.Imagery != "" {
		root.set("imagery", ds.Imagery)
	}
	if ds.Layout != "" {
		root.set("layout", ds.Layout)
	}
	setRaw(root, "customSections", dsRaw)
	setRaw(root, "similar", dsRaw)

	return root.marshal()
}

// setRaw copies a section through verbatim, skipping it when upstream omitted the key or
// set it to null, which is what the spread-conditional does.
func setRaw(o *jsonObj, key string, fields map[string]json.RawMessage) {
	v, ok := fields[key]
	if !ok || string(v) == "null" {
		return
	}
	o.set(key, rawJSON(v))
}

func designSystemFields(rawResult []byte) (map[string]json.RawMessage, error) {
	var top map[string]json.RawMessage
	if err := json.Unmarshal(rawResult, &top); err != nil {
		return nil, err
	}
	out := map[string]json.RawMessage{}
	if ds, ok := top["designSystem"]; ok {
		if err := json.Unmarshal(ds, &out); err != nil {
			return nil, err
		}
	}
	return out, nil
}

func rawField(obj json.RawMessage, key string) (rawJSON, bool) {
	if len(obj) == 0 {
		return nil, false
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(obj, &fields); err != nil {
		return nil, false
	}
	v, ok := fields[key]
	if !ok {
		return nil, false
	}
	return rawJSON(v), true
}

// screenshotURLs prefers the preview slice list from telemetry and falls back to the
// single screenshot, dropping blanks and duplicates.
func screenshotURLs(res refero.Result, rawResult []byte) rawJSON {
	var urls []string
	var tel struct {
		PreviewSliceURLs []string `json:"previewSliceUrls"`
	}
	if len(res.Meta.Telemetry) > 0 {
		_ = json.Unmarshal(res.Meta.Telemetry, &tel)
	}
	if tel.PreviewSliceURLs != nil {
		urls = tel.PreviewSliceURLs
	} else {
		urls = []string{res.Screenshot.URL}
	}
	seen := map[string]bool{}
	out := make([]string, 0, len(urls))
	for _, u := range urls {
		if u == "" || seen[u] {
			continue
		}
		seen[u] = true
		out = append(out, u)
	}
	b, err := json.Marshal(out)
	if err != nil {
		return rawJSON("[]")
	}
	return rawJSON(b)
}

// slugLoose is the looser key used only for matching a described font back to a measured
// one: lowercase with whitespace collapsed to dashes, nothing stripped.
func slugLoose(s string) string {
	return reSpace.ReplaceAllString(strings.ToLower(s), "-")
}

func typographyEntries(res refero.Result) rawJSON {
	arr := make([]any, 0, len(res.DesignSystem.Typography))
	for _, tr := range res.DesignSystem.Typography {
		o := newObj()
		o.set("family", tr.Family)
		o.set("source", fontSource(res.Raw.Typography.Fonts, tr.Family))
		if tr.Substitute != "" {
			o.set("substitute", tr.Substitute)
		}
		o.set("weight", tr.Weight)
		o.set("sizes", tr.Sizes)
		o.set("lineHeight", tr.LineHeight)
		if tr.LetterSpacing != "" {
			o.set("letterSpacing", tr.LetterSpacing)
		}
		if tr.FontFeatureSettings != "" {
			o.set("fontFeatureSettings", tr.FontFeatureSettings)
		}
		o.set("role", tr.Role)
		arr = append(arr, o)
	}
	return marshalObjArray(arr)
}

func fontSource(fonts []refero.Font, family string) string {
	want := slugLoose(family)
	for _, f := range fonts {
		if slugLoose(f.Family) == want {
			if f.Source == "" {
				return "custom"
			}
			return f.Source
		}
	}
	return "custom"
}

// marshalObjArray joins encoded objects into an array fragment; reindent fixes the
// depth when it is spliced into the document.
func marshalObjArray(items []any) rawJSON {
	var parts []string
	for _, it := range items {
		o, ok := it.(*jsonObj)
		if !ok {
			continue
		}
		s, err := o.marshal()
		if err != nil {
			return rawJSON("[]")
		}
		parts = append(parts, s)
	}
	if len(parts) == 0 {
		return rawJSON("[]")
	}
	return rawJSON("[" + strings.Join(parts, ",") + "]")
}
