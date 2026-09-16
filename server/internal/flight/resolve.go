package flight

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"
)

// RawValue returns the first value stored under key with source key order intact, which
// decoding through map[string]any would lose and the upstream generators depend on.
func (d *Doc) RawValue(key string) (json.RawMessage, bool) {
	needle := []byte(`"` + key + `":`)
	for _, id := range d.Order {
		raw, ok := d.Raw[id]
		if !ok {
			continue
		}
		for off := 0; ; {
			i := bytes.Index(raw[off:], needle)
			if i < 0 {
				break
			}
			start := off + i + len(needle)
			if v, end, err := decodeValue(raw, start); err == nil {
				if _, isObj := v.(map[string]any); isObj {
					return json.RawMessage(raw[start:end]), true
				}
			}
			off = start
		}
	}
	return nil, false
}

// ResolveJSON re-emits raw JSON with "$<hexid>" pointers replaced by the rows they name,
// preserving key order and number formatting. indent of 0 produces compact output.
func (d *Doc) ResolveJSON(raw json.RawMessage, indent int) ([]byte, error) {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	var buf bytes.Buffer
	if err := d.emit(dec, &buf, indent, 0); err != nil {
		return nil, err
	}
	if dec.More() {
		return nil, fmt.Errorf("flight: trailing data after value")
	}
	return buf.Bytes(), nil
}

func (d *Doc) emit(dec *json.Decoder, buf *bytes.Buffer, indent, depth int) error {
	tok, err := dec.Token()
	if err != nil {
		return err
	}
	switch t := tok.(type) {
	case json.Delim:
		switch t {
		case '{':
			return d.emitObject(dec, buf, indent, depth)
		case '[':
			return d.emitArray(dec, buf, indent, depth)
		}
		return fmt.Errorf("flight: unexpected delimiter %v", t)
	case string:
		return writeJSONString(buf, d.resolveString(t))
	case json.Number:
		buf.WriteString(t.String())
		return nil
	case bool:
		if t {
			buf.WriteString("true")
		} else {
			buf.WriteString("false")
		}
		return nil
	case nil:
		buf.WriteString("null")
		return nil
	}
	return fmt.Errorf("flight: unsupported token %T", tok)
}

// resolveString expands a pointer row, or unescapes a literal leading "$$".
func (d *Doc) resolveString(s string) string {
	if strings.HasPrefix(s, "$$") {
		return s[1:]
	}
	m := refRe.FindStringSubmatch(s)
	if m == nil {
		return s
	}
	if text, ok := d.Text[m[1]]; ok {
		return text
	}
	return s
}

func (d *Doc) emitObject(dec *json.Decoder, buf *bytes.Buffer, indent, depth int) error {
	buf.WriteByte('{')
	first := true
	for dec.More() {
		keyTok, err := dec.Token()
		if err != nil {
			return err
		}
		key, ok := keyTok.(string)
		if !ok {
			return fmt.Errorf("flight: non-string object key %v", keyTok)
		}
		if !first {
			buf.WriteByte(',')
		}
		first = false
		writeNewlineIndent(buf, indent, depth+1)
		if err := writeJSONString(buf, key); err != nil {
			return err
		}
		buf.WriteByte(':')
		if indent > 0 {
			buf.WriteByte(' ')
		}
		if err := d.emit(dec, buf, indent, depth+1); err != nil {
			return err
		}
	}
	if !first {
		writeNewlineIndent(buf, indent, depth)
	}
	if _, err := dec.Token(); err != nil {
		return err
	}
	buf.WriteByte('}')
	return nil
}

func (d *Doc) emitArray(dec *json.Decoder, buf *bytes.Buffer, indent, depth int) error {
	buf.WriteByte('[')
	first := true
	for dec.More() {
		if !first {
			buf.WriteByte(',')
		}
		first = false
		writeNewlineIndent(buf, indent, depth+1)
		if err := d.emit(dec, buf, indent, depth+1); err != nil {
			return err
		}
	}
	if !first {
		writeNewlineIndent(buf, indent, depth)
	}
	if _, err := dec.Token(); err != nil {
		return err
	}
	buf.WriteByte(']')
	return nil
}

func writeNewlineIndent(buf *bytes.Buffer, indent, depth int) {
	if indent <= 0 {
		return
	}
	buf.WriteByte('\n')
	for i := 0; i < indent*depth; i++ {
		buf.WriteByte(' ')
	}
}

// writeJSONString encodes without Go's HTML escaping, so output matches JSON.stringify.
func writeJSONString(buf *bytes.Buffer, s string) error {
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(s); err != nil {
		return err
	}
	buf.Truncate(buf.Len() - 1) // Encode appends a newline
	return nil
}
