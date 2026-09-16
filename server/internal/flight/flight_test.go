package flight

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func fixture(t *testing.T, name string) []byte {
	t.Helper()
	b, err := os.ReadFile(filepath.Join("..", "..", "testdata", "fixtures", name))
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	return b
}

func TestParseRSCAndHTMLAgree(t *testing.T) {
	rsc, err := Parse(fixture(t, "style-shade.rsc.txt"))
	if err != nil {
		t.Fatalf("parse rsc: %v", err)
	}
	html, err := Parse(fixture(t, "style-shade.html"))
	if err != nil {
		t.Fatalf("parse html: %v", err)
	}
	for _, d := range []*Doc{rsc, html} {
		if _, ok := d.FindObject("meta", "raw", "designSystem"); !ok {
			t.Fatal("result object not found")
		}
	}
}

func TestTextRowLengthPrefix(t *testing.T) {
	d, err := Parse(fixture(t, "style-shade.rsc.txt"))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	got, ok := textRowContaining(d, "Quick Color Reference")
	if !ok {
		t.Fatal("length-prefixed text row not recovered")
	}
	if !strings.HasPrefix(got, "Quick Color Reference") {
		t.Fatalf("text row starts with %q", got[:min(40, len(got))])
	}
	if len(got) != 0x661 {
		t.Fatalf("text row length = %d, want %d", len(got), 0x661)
	}
	if len(d.Order) < 60 {
		t.Fatalf("recovered %d rows, want at least 60; rows after a text row were lost", len(d.Order))
	}
}

func TestResolveReferences(t *testing.T) {
	d, err := Parse(fixture(t, "style-shade.rsc.txt"))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	res, ok := d.FindObject("meta", "raw", "designSystem")
	if !ok {
		t.Fatal("result object not found")
	}
	ds, _ := res["designSystem"].(map[string]any)
	secs, _ := ds["customSections"].([]any)
	if len(secs) == 0 {
		t.Fatal("no customSections")
	}
	first, _ := secs[0].(map[string]any)
	content, _ := first["content"].(string)
	if strings.HasPrefix(content, "$") {
		t.Fatalf("customSections[0].content left unresolved: %q", content)
	}
	if !strings.Contains(content, "Quick Color Reference") {
		t.Fatalf("customSections[0].content = %q", content[:min(60, len(content))])
	}
}

func TestFindObjectsCards(t *testing.T) {
	d, err := Parse(fixture(t, "style-shade.rsc.txt"))
	if err != nil {
		t.Fatalf("parse: %v", err)
	}
	cards := d.FindObjects("id", "id", "siteName", "screenshotUrl")
	if len(cards) < 15 {
		t.Fatalf("found %d cards, want at least 15", len(cards))
	}
	for _, c := range cards {
		if _, ok := c["url"].(string); !ok {
			t.Fatalf("card %v missing url", c["id"])
		}
	}
}

func textRowContaining(d *Doc, sub string) (string, bool) {
	for _, v := range d.Text {
		if strings.Contains(v, sub) {
			return v, true
		}
	}
	return "", false
}
