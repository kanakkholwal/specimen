package normalize

import "testing"

func TestIdentify(t *testing.T) {
	cases := []struct{ in, origin, etld1, sub string }{
		{"https://www.11x.ai", "11x.ai", "11x.ai", ""},
		{"https://dub.sh", "dub.sh", "dub.sh", ""},
		{"https://shade.inc", "shade.inc", "shade.inc", ""},
		{"https://app.example.com/pricing", "app.example.com", "example.com", "app"},
		{"https://www.sequencehq.com/", "sequencehq.com", "sequencehq.com", ""},
		{"styles.refero.design", "styles.refero.design", "refero.design", "styles"},
		{"https://foo.co.uk", "foo.co.uk", "foo.co.uk", ""},
		{"https://a.b.foo.co.uk", "a.b.foo.co.uk", "foo.co.uk", "a.b"},
	}
	for _, c := range cases {
		got, err := Identify(c.in)
		if err != nil {
			t.Fatalf("Identify(%q): %v", c.in, err)
		}
		if got.Origin != c.origin || got.ETLD1 != c.etld1 || got.Subdomain != c.sub {
			t.Errorf("Identify(%q) = %+v, want origin=%s etld1=%s sub=%s",
				c.in, got, c.origin, c.etld1, c.sub)
		}
	}
}

func TestIdentifyRejectsEmpty(t *testing.T) {
	if _, err := Identify("   "); err == nil {
		t.Fatal("want error for empty host")
	}
}

func TestSlug(t *testing.T) {
	cases := map[string]string{
		"Charcoal Ink":     "charcoal-ink",
		"Inter Display":    "inter-display",
		"  Bone  ":         "bone",
		"Cutout Gray/2":    "cutout-gray-2",
		"PP Neue Montreal": "pp-neue-montreal",
		"---":              "",
	}
	for in, want := range cases {
		if got := Slug(in); got != want {
			t.Errorf("Slug(%q) = %q, want %q", in, got, want)
		}
	}
}
