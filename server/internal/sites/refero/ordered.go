package refero

import (
	"bytes"
	"encoding/json"
	"fmt"
)

// OrderedStrings is a string map that remembers JSON key order. The upstream generators
// iterate these objects with Object.entries, so key order is part of the output.
type OrderedStrings struct {
	Keys   []string
	Values map[string]string
}

func (o OrderedStrings) Len() int { return len(o.Keys) }

func (o OrderedStrings) Get(k string) string { return o.Values[k] }

func (o *OrderedStrings) UnmarshalJSON(b []byte) error {
	o.Keys = nil
	o.Values = map[string]string{}
	if bytes.Equal(bytes.TrimSpace(b), []byte("null")) {
		return nil
	}
	dec := json.NewDecoder(bytes.NewReader(b))
	tok, err := dec.Token()
	if err != nil {
		return err
	}
	if d, ok := tok.(json.Delim); !ok || d != '{' {
		return fmt.Errorf("refero: expected object, got %v", tok)
	}
	for dec.More() {
		keyTok, err := dec.Token()
		if err != nil {
			return err
		}
		key, ok := keyTok.(string)
		if !ok {
			return fmt.Errorf("refero: non-string key %v", keyTok)
		}
		var raw json.RawMessage
		if err := dec.Decode(&raw); err != nil {
			return err
		}
		var val string
		if err := json.Unmarshal(raw, &val); err != nil {
			// Tolerate a non-string value rather than losing the whole object.
			val = string(raw)
		}
		if _, seen := o.Values[key]; !seen {
			o.Keys = append(o.Keys, key)
		}
		o.Values[key] = val
	}
	_, err = dec.Token()
	return err
}

func (o OrderedStrings) MarshalJSON() ([]byte, error) {
	if o.Keys == nil && o.Values == nil {
		return []byte("null"), nil
	}
	var buf bytes.Buffer
	buf.WriteByte('{')
	for i, k := range o.Keys {
		if i > 0 {
			buf.WriteByte(',')
		}
		key, err := json.Marshal(k)
		if err != nil {
			return nil, err
		}
		val, err := json.Marshal(o.Values[k])
		if err != nil {
			return nil, err
		}
		buf.Write(key)
		buf.WriteByte(':')
		buf.Write(val)
	}
	buf.WriteByte('}')
	return buf.Bytes(), nil
}
