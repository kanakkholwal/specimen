package render

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// rawJSON is a pre-encoded fragment spliced in verbatim and re-indented to its new depth.
type rawJSON []byte

// jsonObj is an insertion-ordered JSON object. Re-setting a key overwrites the value but
// keeps its original position, which is how JavaScript objects behave.
type jsonObj struct {
	keys []string
	vals map[string]any
}

func newObj() *jsonObj { return &jsonObj{vals: map[string]any{}} }

func (o *jsonObj) set(k string, v any) {
	if _, seen := o.vals[k]; !seen {
		o.keys = append(o.keys, k)
	}
	o.vals[k] = v
}

func (o *jsonObj) len() int { return len(o.keys) }

// marshal renders the object the way JSON.stringify(value, null, 2) would.
func (o *jsonObj) marshal() (string, error) {
	var buf bytes.Buffer
	if err := writeValue(&buf, o, 0); err != nil {
		return "", err
	}
	return buf.String(), nil
}

func writeValue(buf *bytes.Buffer, v any, depth int) error {
	switch t := v.(type) {
	case *jsonObj:
		return writeObj(buf, t, depth)
	case rawJSON:
		return reindent(buf, t, depth)
	case string:
		return writeStr(buf, t)
	case float64:
		buf.WriteString(num(t))
		return nil
	case int:
		buf.WriteString(num(float64(t)))
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
	return fmt.Errorf("render: unsupported json value %T", v)
}

func writeObj(buf *bytes.Buffer, o *jsonObj, depth int) error {
	if o.len() == 0 {
		buf.WriteString("{}")
		return nil
	}
	buf.WriteString("{\n")
	for i, k := range o.keys {
		if i > 0 {
			buf.WriteString(",\n")
		}
		pad(buf, depth+1)
		if err := writeStr(buf, k); err != nil {
			return err
		}
		buf.WriteString(": ")
		if err := writeValue(buf, o.vals[k], depth+1); err != nil {
			return err
		}
	}
	buf.WriteByte('\n')
	pad(buf, depth)
	buf.WriteByte('}')
	return nil
}

func pad(buf *bytes.Buffer, depth int) {
	for i := 0; i < depth*2; i++ {
		buf.WriteByte(' ')
	}
}

func writeStr(buf *bytes.Buffer, s string) error {
	enc := json.NewEncoder(buf)
	enc.SetEscapeHTML(false)
	if err := enc.Encode(s); err != nil {
		return err
	}
	buf.Truncate(buf.Len() - 1)
	return nil
}

// reindent re-emits an encoded fragment at a new depth, preserving key order.
func reindent(buf *bytes.Buffer, raw []byte, depth int) error {
	dec := json.NewDecoder(bytes.NewReader(raw))
	dec.UseNumber()
	return reindentValue(dec, buf, depth)
}

func reindentValue(dec *json.Decoder, buf *bytes.Buffer, depth int) error {
	tok, err := dec.Token()
	if err != nil {
		return err
	}
	switch t := tok.(type) {
	case json.Delim:
		switch t {
		case '{':
			return reindentObject(dec, buf, depth)
		case '[':
			return reindentArray(dec, buf, depth)
		}
		return fmt.Errorf("render: unexpected delimiter %v", t)
	case string:
		return writeStr(buf, t)
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
	return fmt.Errorf("render: unsupported token %T", tok)
}

func reindentObject(dec *json.Decoder, buf *bytes.Buffer, depth int) error {
	if !dec.More() {
		if _, err := dec.Token(); err != nil {
			return err
		}
		buf.WriteString("{}")
		return nil
	}
	buf.WriteString("{\n")
	first := true
	for dec.More() {
		if !first {
			buf.WriteString(",\n")
		}
		first = false
		keyTok, err := dec.Token()
		if err != nil {
			return err
		}
		key, ok := keyTok.(string)
		if !ok {
			return fmt.Errorf("render: non-string key %v", keyTok)
		}
		pad(buf, depth+1)
		if err := writeStr(buf, key); err != nil {
			return err
		}
		buf.WriteString(": ")
		if err := reindentValue(dec, buf, depth+1); err != nil {
			return err
		}
	}
	buf.WriteByte('\n')
	pad(buf, depth)
	buf.WriteByte('}')
	if _, err := dec.Token(); err != nil {
		return err
	}
	return nil
}

func reindentArray(dec *json.Decoder, buf *bytes.Buffer, depth int) error {
	if !dec.More() {
		if _, err := dec.Token(); err != nil {
			return err
		}
		buf.WriteString("[]")
		return nil
	}
	buf.WriteString("[\n")
	first := true
	for dec.More() {
		if !first {
			buf.WriteString(",\n")
		}
		first = false
		pad(buf, depth+1)
		if err := reindentValue(dec, buf, depth+1); err != nil {
			return err
		}
	}
	buf.WriteByte('\n')
	pad(buf, depth)
	buf.WriteByte(']')
	if _, err := dec.Token(); err != nil {
		return err
	}
	return nil
}
