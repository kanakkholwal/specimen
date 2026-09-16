// Package normalize derives stable identity and slug keys from scraped values.
package normalize

import (
	"net/url"
	"strings"

	"golang.org/x/net/publicsuffix"
)

// Origin is the grouping key for a scraped site. Subdomains other than www are kept
// distinct, so app.foo.com and foo.com remain separate sites sharing an ETLD1.
type Origin struct {
	Origin    string // 11x.ai, app.example.com
	ETLD1     string // 11x.ai, example.com
	Subdomain string // "", "app"
	Scheme    string
}

// Identify parses a site URL into its grouping key. A bare host is accepted.
func Identify(raw string) (Origin, error) {
	s := strings.TrimSpace(raw)
	if s == "" {
		return Origin{}, &url.Error{Op: "parse", URL: raw, Err: errEmpty{}}
	}
	if !strings.Contains(s, "//") {
		s = "https://" + s
	}
	u, err := url.Parse(s)
	if err != nil {
		return Origin{}, err
	}
	host := strings.ToLower(u.Hostname())
	host = strings.TrimSuffix(host, ".")
	host = strings.TrimPrefix(host, "www.")
	if host == "" {
		return Origin{}, &url.Error{Op: "parse", URL: raw, Err: errEmpty{}}
	}
	etld1, err := publicsuffix.EffectiveTLDPlusOne(host)
	if err != nil {
		etld1 = host
	}
	sub := strings.TrimSuffix(strings.TrimSuffix(host, etld1), ".")
	scheme := u.Scheme
	if scheme == "" {
		scheme = "https"
	}
	return Origin{Origin: host, ETLD1: etld1, Subdomain: sub, Scheme: scheme}, nil
}

type errEmpty struct{}

func (errEmpty) Error() string { return "empty host" }

// Slug lowercases and reduces a label to its alphanumeric runs joined by dashes.
// It mirrors the slug rule used by the upstream CSS variable generator.
func Slug(s string) string {
	var b strings.Builder
	b.Grow(len(s))
	prevDash := false
	for _, r := range strings.ToLower(s) {
		switch {
		case r >= 'a' && r <= 'z', r >= '0' && r <= '9':
			b.WriteRune(r)
			prevDash = false
		default:
			if !prevDash && b.Len() > 0 {
				b.WriteByte('-')
				prevDash = true
			}
		}
	}
	return strings.Trim(b.String(), "-")
}

// PathSafe reduces a value to something usable as a directory name on any OS.
func PathSafe(s string) string {
	s = strings.Map(func(r rune) rune {
		if strings.ContainsRune(`<>:"/\|?*`, r) || r < 0x20 {
			return '-'
		}
		return r
	}, s)
	s = strings.Trim(strings.TrimSpace(s), ". ")
	if s == "" {
		return "unknown"
	}
	return s
}
