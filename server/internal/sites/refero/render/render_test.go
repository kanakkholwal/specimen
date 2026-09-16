package render

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/kanakkholwal/design-supply/internal/sites/refero"
)

func loadResult(t *testing.T) refero.Result {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "..", "..",
		"testdata", "fixtures", "style-shade.result.json"))
	if err != nil {
		t.Fatalf("read result fixture: %v", err)
	}
	var r refero.Result
	if err := json.Unmarshal(b, &r); err != nil {
		t.Fatalf("decode result: %v", err)
	}
	return r
}

func golden(t *testing.T, name string) string {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "..", "..", "testdata", "golden", "shade", name))
	if err != nil {
		t.Fatalf("read golden %s: %v", name, err)
	}
	return string(b)
}

// assertSame reports the first differing line, which is far easier to act on than a
// wall of diff when porting a generator.
func assertSame(t *testing.T, name, got, want string) {
	t.Helper()
	if got == want {
		return
	}
	g, w := strings.Split(got, "\n"), strings.Split(want, "\n")
	for i := 0; i < len(g) || i < len(w); i++ {
		gl, wl := "", ""
		if i < len(g) {
			gl = g[i]
		}
		if i < len(w) {
			wl = w[i]
		}
		if gl != wl {
			t.Fatalf("%s differs at line %d of %d/%d\n  got:  %q\n  want: %q", name, i+1, len(g), len(w), gl, wl)
		}
	}
	t.Fatalf("%s differs in length only: got %d bytes, want %d", name, len(got), len(want))
}

func TestCSSVariablesMatchesUpstream(t *testing.T) {
	res := loadResult(t)
	assertSame(t, "variables.compact.css", CSSVariables(res, Compact), golden(t, "variables.compact.css"))
	assertSame(t, "variables.extended.css", CSSVariables(res, Extended), golden(t, "variables.extended.css"))
}

func TestTailwindThemeMatchesUpstream(t *testing.T) {
	res := loadResult(t)
	assertSame(t, "theme.compact.css", TailwindTheme(res, Compact), golden(t, "theme.compact.css"))
	assertSame(t, "theme.extended.css", TailwindTheme(res, Extended), golden(t, "theme.extended.css"))
}

func TestDesignMdMatchesUpstream(t *testing.T) {
	res := loadResult(t)
	assertSame(t, "DESIGN.compact.md", DesignMd(res, Compact), golden(t, "DESIGN.compact.md"))
	assertSame(t, "DESIGN.extended.md", DesignMd(res, Extended), golden(t, "DESIGN.extended.md"))
}

func TestTokensJSONMatchesUpstream(t *testing.T) {
	res := loadResult(t)
	for _, tc := range []struct {
		variant Variant
		file    string
	}{{Compact, "tokens.compact.json"}, {Extended, "tokens.extended.json"}} {
		got, err := TokensJSON(res, tc.variant)
		if err != nil {
			t.Fatalf("%s: %v", tc.file, err)
		}
		assertSame(t, tc.file, got, golden(t, tc.file))
	}
}

func TestDesignSystemJSONMatchesUpstream(t *testing.T) {
	res := loadResult(t)
	raw, err := os.ReadFile(filepath.Join("..", "..", "..", "..",
		"testdata", "fixtures", "style-shade.result.json"))
	if err != nil {
		t.Fatal(err)
	}
	got, err := DesignSystemJSON(res, raw)
	if err != nil {
		t.Fatalf("design-system.json: %v", err)
	}
	assertSame(t, "design-system.json", got, golden(t, "design-system.json"))
}

func TestPromptTextMatchesUpstream(t *testing.T) {
	assertSame(t, "prompt.txt", PromptText(loadResult(t)), golden(t, "prompt.txt"))
}
