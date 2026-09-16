package render

import (
	"encoding/json"
	"fmt"

	"github.com/kanakkholwal/design-supply/internal/sites/refero"
)

// Version tags generated artifacts. Bump it whenever a generator's output changes so
// stale files are identifiable in the artifacts table.
const Version = "1"

type Artifact struct {
	Name string
	Data []byte
}

// All renders every export format for one style. rawResult is the archived result.json,
// needed because the Design JSON export splices sections through verbatim.
func All(res refero.Result, rawResult []byte) ([]Artifact, error) {
	designJSON, err := DesignSystemJSON(res, rawResult)
	if err != nil {
		return nil, fmt.Errorf("design-system.json: %w", err)
	}
	tokensCompact, err := TokensJSON(res, Compact)
	if err != nil {
		return nil, fmt.Errorf("tokens.compact.json: %w", err)
	}
	tokensExtended, err := TokensJSON(res, Extended)
	if err != nil {
		return nil, fmt.Errorf("tokens.extended.json: %w", err)
	}
	return []Artifact{
		{"DESIGN.compact.md", []byte(DesignMd(res, Compact))},
		{"DESIGN.extended.md", []byte(DesignMd(res, Extended))},
		{"variables.compact.css", []byte(CSSVariables(res, Compact))},
		{"variables.extended.css", []byte(CSSVariables(res, Extended))},
		{"theme.compact.css", []byte(TailwindTheme(res, Compact))},
		{"theme.extended.css", []byte(TailwindTheme(res, Extended))},
		{"tokens.compact.json", []byte(tokensCompact)},
		{"tokens.extended.json", []byte(tokensExtended)},
		{"design-system.json", []byte(designJSON)},
		{"prompt.txt", []byte(PromptText(res))},
	}, nil
}

// FromRawResult renders every export straight from an archived result.json.
func FromRawResult(rawResult []byte) ([]Artifact, error) {
	var res refero.Result
	if err := json.Unmarshal(rawResult, &res); err != nil {
		return nil, fmt.Errorf("decode result: %w", err)
	}
	return All(res, rawResult)
}
