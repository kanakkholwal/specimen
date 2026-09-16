// Package flight parses the React Server Components wire format used by Next.js.
package flight

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

// Doc is a parsed flight stream: text rows, JSON rows, and their original order.
type Doc struct {
	Text  map[string]string
	JSON  map[string]any
	Raw   map[string][]byte
	Order []string
}

var errShort = errors.New("flight: truncated row")

// Parse accepts either a raw text/x-component body or HTML embedding __next_f chunks.
func Parse(body []byte) (*Doc, error) {
	s := string(body)
	if strings.Contains(s, "__next_f.push") {
		joined, err := joinHTMLChunks(s)
		if err != nil {
			return nil, err
		}
		s = joined
	}
	return parseStream(s)
}

var pushRe = regexp.MustCompile(`self\.__next_f\.push\(\[1,("(?:[^"\\]|\\.)*")\]\)`)

func joinHTMLChunks(html string) (string, error) {
	var b strings.Builder
	for _, m := range pushRe.FindAllStringSubmatch(html, -1) {
		var chunk string
		if err := json.Unmarshal([]byte(m[1]), &chunk); err != nil {
			return "", fmt.Errorf("flight: bad __next_f chunk: %w", err)
		}
		b.WriteString(chunk)
	}
	if b.Len() == 0 {
		return "", errors.New("flight: no __next_f chunks found")
	}
	return b.String(), nil
}

// parseStream walks "<hexid>:<payload>" rows. Text rows are length-prefixed "T<hexlen>,";
// others end where their JSON ends, since rows are not reliably newline-separated.
func parseStream(s string) (*Doc, error) {
	d := &Doc{Text: map[string]string{}, JSON: map[string]any{}, Raw: map[string][]byte{}}
	b := []byte(s)
	for i := 0; i < len(b); {
		if b[i] == '\n' {
			i++
			continue
		}
		colon := indexByteFrom(b, i, ':')
		if colon < 0 {
			break
		}
		id := string(b[i:colon])
		if id == "" || len(id) > 8 || !isHex(id) {
			i = resync(b, i)
			continue
		}
		p := colon + 1
		if p >= len(b) {
			return nil, errShort
		}
		if b[p] == 'T' {
			comma := indexByteFrom(b, p, ',')
			if comma < 0 {
				return nil, errShort
			}
			n, err := strconv.ParseInt(string(b[p+1:comma]), 16, 64)
			if err != nil {
				return nil, fmt.Errorf("flight: bad text length in row %s: %w", id, err)
			}
			end := min(comma+1+int(n), len(b))
			d.set(id, string(b[comma+1:end]), nil)
			i = end
			continue
		}
		// I/H/E/P rows are tagged; the JSON value begins after the single-letter tag.
		start := p
		if b[start] >= 'A' && b[start] <= 'Z' {
			start++
		}
		v, end, err := decodeValue(b, start)
		if err != nil {
			i = resync(b, i)
			continue
		}
		d.set(id, "", v)
		d.Raw[id] = b[start:end]
		i = end
	}
	if len(d.Order) == 0 {
		return nil, errors.New("flight: no rows parsed")
	}
	return d, nil
}

// decodeValue reads exactly one JSON value starting at off and reports where it ended.
func decodeValue(b []byte, off int) (any, int, error) {
	if off >= len(b) {
		return nil, off, errShort
	}
	dec := json.NewDecoder(bytes.NewReader(b[off:]))
	dec.UseNumber()
	var v any
	if err := dec.Decode(&v); err != nil {
		return nil, off, err
	}
	return v, off + int(dec.InputOffset()), nil
}

// resync skips to the next plausible row header after a malformed row.
func resync(b []byte, from int) int {
	for i := from + 1; i < len(b); i++ {
		if b[i] != '\n' {
			continue
		}
		if c := indexByteFrom(b, i+1, ':'); c > i+1 && isHex(string(b[i+1:c])) {
			return i + 1
		}
	}
	return len(b)
}

func (d *Doc) set(id, text string, v any) {
	if _, seen := d.Text[id]; !seen {
		if _, seen := d.JSON[id]; !seen {
			d.Order = append(d.Order, id)
		}
	}
	if v != nil {
		d.JSON[id] = v
		return
	}
	d.Text[id] = text
}

func indexByteFrom(b []byte, from int, c byte) int {
	for i := from; i < len(b); i++ {
		if b[i] == c {
			return i
		}
		if c == ':' && b[i] == '\n' {
			return -1
		}
	}
	return -1
}

func isHex(s string) bool {
	for i := 0; i < len(s); i++ {
		c := s[i]
		if !(c >= '0' && c <= '9' || c >= 'a' && c <= 'f') {
			return false
		}
	}
	return true
}

var refRe = regexp.MustCompile(`^\$([0-9a-f]{1,8})$`)

// Resolve rewrites "$<hexid>" pointers to the row they name and unescapes "$$" literals.
func (d *Doc) Resolve(v any) any {
	return d.resolve(v, 0)
}

func (d *Doc) resolve(v any, depth int) any {
	if depth > 32 {
		return v
	}
	switch t := v.(type) {
	case string:
		if strings.HasPrefix(t, "$$") {
			return t[1:]
		}
		m := refRe.FindStringSubmatch(t)
		if m == nil {
			return t
		}
		if s, ok := d.Text[m[1]]; ok {
			return s
		}
		if j, ok := d.JSON[m[1]]; ok {
			return d.resolve(j, depth+1)
		}
		return t
	case []any:
		out := make([]any, len(t))
		for i, e := range t {
			out[i] = d.resolve(e, depth+1)
		}
		return out
	case map[string]any:
		out := make(map[string]any, len(t))
		for k, e := range t {
			out[k] = d.resolve(e, depth+1)
		}
		return out
	default:
		return v
	}
}

// Find returns every value stored under key across all JSON rows, refs resolved,
// in document order. Objects are matched by key name only.
func (d *Doc) Find(key string) []any {
	var out []any
	for _, id := range d.Order {
		v, ok := d.JSON[id]
		if !ok {
			continue
		}
		collect(v, key, &out)
	}
	for i := range out {
		out[i] = d.Resolve(out[i])
	}
	return out
}

// FindObject returns the first object carrying all of the given keys.
func (d *Doc) FindObject(keys ...string) (map[string]any, bool) {
	for _, id := range d.Order {
		v, ok := d.JSON[id]
		if !ok {
			continue
		}
		if m, found := searchObject(v, keys); found {
			r, _ := d.Resolve(m).(map[string]any)
			return r, true
		}
	}
	return nil, false
}

// FindObjects returns every object carrying all of the given keys, de-duplicated by
// the value of dedupeKey when that key is a string.
func (d *Doc) FindObjects(dedupeKey string, keys ...string) []map[string]any {
	var out []map[string]any
	seen := map[string]bool{}
	for _, id := range d.Order {
		v, ok := d.JSON[id]
		if !ok {
			continue
		}
		var walk func(any)
		walk = func(n any) {
			switch t := n.(type) {
			case map[string]any:
				if hasAll(t, keys) {
					if k, _ := t[dedupeKey].(string); k == "" || !seen[k] {
						seen[k] = true
						r, _ := d.Resolve(t).(map[string]any)
						out = append(out, r)
					}
				}
				for _, e := range t {
					walk(e)
				}
			case []any:
				for _, e := range t {
					walk(e)
				}
			}
		}
		walk(v)
	}
	return out
}

func hasAll(m map[string]any, keys []string) bool {
	for _, k := range keys {
		if _, ok := m[k]; !ok {
			return false
		}
	}
	return true
}

func searchObject(n any, keys []string) (map[string]any, bool) {
	switch t := n.(type) {
	case map[string]any:
		if hasAll(t, keys) {
			return t, true
		}
		for _, e := range t {
			if m, ok := searchObject(e, keys); ok {
				return m, true
			}
		}
	case []any:
		for _, e := range t {
			if m, ok := searchObject(e, keys); ok {
				return m, true
			}
		}
	}
	return nil, false
}

func collect(n any, key string, out *[]any) {
	switch t := n.(type) {
	case map[string]any:
		if v, ok := t[key]; ok {
			*out = append(*out, v)
		}
		for _, e := range t {
			collect(e, key, out)
		}
	case []any:
		for _, e := range t {
			collect(e, key, out)
		}
	}
}

// Decode marshals a resolved value into a typed destination.
func Decode(v any, dst any) error {
	b, err := json.Marshal(v)
	if err != nil {
		return err
	}
	return json.Unmarshal(b, dst)
}
