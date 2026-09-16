package refero

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"testing"

	"github.com/kanakkholwal/design-supply/internal/flight"
)

const shadeID = "e549766e-b8b1-48a2-bd72-8cc04e9e4e9d"

func loadDoc(t *testing.T, name string) *flight.Doc {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "..", "testdata", "fixtures", name))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	d, err := flight.Parse(b)
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	return d
}

func TestExtractPage(t *testing.T) {
	page, err := ExtractPage(loadDoc(t, "style-shade.rsc.txt"), shadeID)
	if err != nil {
		t.Fatalf("extract: %v", err)
	}
	s := page.Self
	if s == nil {
		t.Fatal("no self style")
	}
	if s.Result.Meta.SiteName != "Shade" || s.Result.Meta.URL != "https://shade.inc" {
		t.Fatalf("meta = %+v", s.Result.Meta)
	}
	if s.Origin != "shade.inc" || s.ETLD1 != "shade.inc" {
		t.Fatalf("origin = %s etld1 = %s", s.Origin, s.ETLD1)
	}
	if len(s.ResultSHA) != 64 {
		t.Fatalf("result sha = %q", s.ResultSHA)
	}
	if got := len(s.Result.Raw.Colors.Tokens); got != 22 {
		t.Errorf("raw colour tokens = %d, want 22", got)
	}
	if got := len(s.Result.Raw.Typography.Steps); got != 30 {
		t.Errorf("type steps = %d, want 30", got)
	}
	if got := len(s.Result.DesignSystem.Components); got != 11 {
		t.Errorf("components = %d, want 11", got)
	}
	if got := s.Result.DesignSystem.Spacing.Radius.Get("buttons"); got != "35px" {
		t.Errorf("radius.buttons = %q, want 35px", got)
	}
	if !strings.HasPrefix(s.Card.PreviewVideoURL, "https://images.refero.design/") {
		t.Errorf("own preview video url = %q", s.Card.PreviewVideoURL)
	}
	if !strings.HasPrefix(s.Card.ScreenshotURL, "https://images.refero.design/") {
		t.Errorf("own screenshot url = %q", s.Card.ScreenshotURL)
	}
	if len(page.Cards) < 15 {
		t.Fatalf("related cards = %d, want at least 15", len(page.Cards))
	}
	for _, c := range page.Cards {
		if c.ID == shadeID {
			t.Fatal("a style page should not carry its own card")
		}
	}
}

func TestCustomSectionReferenceResolved(t *testing.T) {
	page, err := ExtractPage(loadDoc(t, "style-shade.rsc.txt"), shadeID)
	if err != nil {
		t.Fatalf("extract: %v", err)
	}
	secs := page.Self.Result.DesignSystem.CustomSections
	if len(secs) == 0 {
		t.Fatal("no custom sections")
	}
	if strings.HasPrefix(secs[0].Content, "$") || len(secs[0].Content) < 100 {
		t.Fatalf("custom section content not resolved: %q", secs[0].Content)
	}
}

// TestResultModelIsLossless round-trips the payload through the typed model and fails on
// any field the model silently drops. This is the guard against upstream schema drift.
func TestResultModelIsLossless(t *testing.T) {
	doc := loadDoc(t, "style-shade.rsc.txt")
	host, ok := doc.FindObject("meta", "raw", "designSystem")
	if !ok {
		t.Fatal("result object not found")
	}
	original, err := json.Marshal(host)
	if err != nil {
		t.Fatal(err)
	}
	var res Result
	if err := json.Unmarshal(original, &res); err != nil {
		t.Fatalf("unmarshal into model: %v", err)
	}
	round, err := json.Marshal(res)
	if err != nil {
		t.Fatal(err)
	}

	var a, b any
	if err := json.Unmarshal(original, &a); err != nil {
		t.Fatal(err)
	}
	if err := json.Unmarshal(round, &b); err != nil {
		t.Fatal(err)
	}
	missing := map[string]bool{}
	diffPaths("", a, b, missing)
	if len(missing) > 0 {
		paths := make([]string, 0, len(missing))
		for p := range missing {
			paths = append(paths, p)
		}
		sort.Strings(paths)
		t.Fatalf("model drops %d field(s): %s", len(paths), strings.Join(paths, ", "))
	}
}

// diffPaths records paths present in want but absent from got. Array elements collapse to
// [] so a single missing field is reported once rather than per element.
func diffPaths(path string, want, got any, missing map[string]bool) {
	switch w := want.(type) {
	case map[string]any:
		g, ok := got.(map[string]any)
		if !ok {
			missing[path] = true
			return
		}
		for k, wv := range w {
			child := k
			if path != "" {
				child = path + "." + k
			}
			gv, ok := g[k]
			if !ok {
				missing[child] = true
				continue
			}
			diffPaths(child, wv, gv, missing)
		}
	case []any:
		g, ok := got.([]any)
		if !ok || len(g) != len(w) {
			missing[path+"[]"] = true
			return
		}
		for i := range w {
			diffPaths(path+"[]", w[i], g[i], missing)
		}
	}
}
