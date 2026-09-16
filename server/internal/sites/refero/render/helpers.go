// Package render reproduces the export artifacts styles.refero.design generates in the
// browser, ported from reference/refero-js and diffed against it by the golden tests.
package render

import (
	"regexp"
	"sort"
	"strconv"
	"strings"

	"github.com/kanak/design-supply/internal/sites/refero"
)

// Variant selects how much detail an artifact carries.
type Variant string

const (
	Compact  Variant = "compact"
	Extended Variant = "extended"
)

var (
	reApostrophe = regexp.MustCompile(`'`)
	reSpace      = regexp.MustCompile(`\s+`)
	reNotSlug    = regexp.MustCompile(`[^a-z0-9-]`)
	reDashes     = regexp.MustCompile(`-+`)
	reTrimDash   = regexp.MustCompile(`^-|-$`)
)

// slug matches the upstream rule exactly: apostrophes vanish before whitespace becomes a
// dash, so "Don't Go" is "dont-go" rather than "don-t-go".
func slug(s string) string {
	s = strings.ToLower(s)
	s = reApostrophe.ReplaceAllString(s, "")
	s = reSpace.ReplaceAllString(s, "-")
	s = reNotSlug.ReplaceAllString(s, "")
	s = reDashes.ReplaceAllString(s, "-")
	return reTrimDash.ReplaceAllString(s, "")
}

// num renders a float the way JavaScript renders a number in a template literal.
func num(f float64) string { return strconv.FormatFloat(f, 'f', -1, 64) }

func fontStack(family string) string {
	t := strings.ToLower(family)
	switch {
	case strings.Contains(t, "mono"), strings.Contains(t, "code"), strings.Contains(t, "consola"):
		return "ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace"
	case strings.Contains(t, "serif") && !strings.Contains(t, "sans"):
		return `ui-serif, Georgia, Cambria, "Times New Roman", Times, serif`
	default:
		return `ui-sans-serif, system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif`
	}
}

func radiusName(v float64) string {
	switch {
	case v <= 0:
		return "none"
	case v <= 3:
		return "sm"
	case v <= 7:
		return "md"
	case v <= 11:
		return "lg"
	case v <= 15:
		return "xl"
	case v <= 23:
		return "2xl"
	case v <= 47:
		return "3xl"
	default:
		return "full"
	}
}

var reShadowPx = regexp.MustCompile(`-?[\d.]+px`)
var reInsetPrefix = regexp.MustCompile(`(?i)^inset\s+`)

// shadowName buckets a shadow by its blur radius, the third pixel value in the string.
func shadowName(value string) string {
	s := strings.TrimSpace(reInsetPrefix.ReplaceAllString(value, ""))
	m := reShadowPx.FindAllString(s, -1)
	blur := 0.0
	if len(m) >= 3 {
		blur, _ = strconv.ParseFloat(strings.TrimSuffix(m[2], "px"), 64)
	}
	switch {
	case blur < 4:
		return "subtle"
	case blur <= 8:
		return "sm"
	case blur <= 16:
		return "md"
	case blur <= 24:
		return "lg"
	default:
		return "xl"
	}
}

func textSizeName(v float64) string {
	switch {
	case v <= 12:
		return "xs"
	case v <= 14:
		return "sm"
	case v <= 16:
		return "base"
	case v <= 19:
		return "lg"
	case v <= 23:
		return "xl"
	case v <= 29:
		return "2xl"
	case v <= 35:
		return "3xl"
	case v <= 47:
		return "4xl"
	default:
		return "5xl"
	}
}

type named struct {
	Name  string
	Value float64
	Text  string
	Extra float64
}

// dedupeNames suffixes repeats as name-2, name-3, matching the upstream counter.
func dedupeNames(items []named) []named {
	counts := map[string]int{}
	for i := range items {
		base := items[i].Name
		counts[base]++
		if n := counts[base]; n > 1 {
			items[i].Name = base + "-" + strconv.Itoa(n)
		}
	}
	return items
}

// filterSpacing keeps plausible spacing tokens, prefers multiples of the base unit when
// that leaves enough of them, and caps the set at the sixteen most frequent.
func filterSpacing(tokens []refero.SpacingToken, baseUnit *float64) []refero.SpacingToken {
	out := make([]refero.SpacingToken, 0, len(tokens))
	for _, t := range tokens {
		if t.Value >= 4 && t.Value <= 240 && t.Frequency >= 2 {
			out = append(out, t)
		}
	}
	if baseUnit != nil && *baseUnit > 1 {
		var multiples []refero.SpacingToken
		for _, t := range out {
			if modZero(t.Value, *baseUnit) {
				multiples = append(multiples, t)
			}
		}
		if len(multiples) >= 5 {
			out = multiples
		}
	}
	if len(out) > 16 {
		sort.SliceStable(out, func(i, j int) bool { return out[i].Frequency > out[j].Frequency })
		out = out[:16]
	}
	sort.SliceStable(out, func(i, j int) bool { return out[i].Value < out[j].Value })
	return out
}

func modZero(v, base float64) bool {
	if base == 0 {
		return false
	}
	r := v - base*float64(int(v/base))
	return r == 0
}

var weightNames = map[int]string{
	100: "thin", 200: "extralight", 300: "light", 400: "regular", 500: "medium",
	600: "semibold", 700: "bold", 800: "extrabold", 900: "black",
}

func weightName(w int) string {
	if n, ok := weightNames[w]; ok {
		return n
	}
	return "w" + strconv.Itoa(w)
}

// sortedWeights collects every declared font weight across the measured fonts.
func sortedWeights(fonts []refero.Font) []int {
	seen := map[int]bool{}
	var out []int
	for _, f := range fonts {
		for _, w := range f.Weights {
			if !seen[w] {
				seen[w] = true
				out = append(out, w)
			}
		}
	}
	sort.Ints(out)
	return out
}
