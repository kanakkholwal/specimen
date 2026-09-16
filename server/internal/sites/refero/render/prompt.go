package render

import (
	"encoding/json"
	"strings"

	"github.com/kanak/design-supply/internal/sites/refero"
)

const promptMissing = "Prompt snapshot not stored for this extraction.\n\n" +
	"Re-extract this style to capture the exact LLM input for future review."

type promptSnapshot struct {
	SchemaVersion       json.Number `json:"schemaVersion"`
	Kind                string      `json:"kind"`
	CreatedAt           string      `json:"createdAt"`
	SystemPromptVersion string      `json:"systemPromptVersion"`
	Model               string      `json:"model"`
	TextChars           json.Number `json:"textChars"`
	ImageCount          json.Number `json:"imageCount"`
	ImageURLs           []string    `json:"imageUrls"`
	UserPromptText      string      `json:"userPromptText"`
}

// PromptText renders the stored LLM prompt snapshot, or the placeholder when the
// extraction predates snapshot capture.
func PromptText(res refero.Result) string {
	if len(res.Meta.Telemetry) == 0 {
		return promptMissing
	}
	var tel struct {
		PromptSnapshot *promptSnapshot `json:"promptSnapshot"`
	}
	if err := json.Unmarshal(res.Meta.Telemetry, &tel); err != nil || tel.PromptSnapshot == nil {
		return promptMissing
	}
	s := tel.PromptSnapshot

	out := []string{
		"Prompt snapshot v" + s.SchemaVersion.String(),
		"Kind: " + s.Kind,
		"Created: " + s.CreatedAt,
		"System prompt version: " + s.SystemPromptVersion,
	}
	if s.Model != "" {
		out = append(out, "Model: "+s.Model)
	}
	out = append(out,
		"Text chars: "+s.TextChars.String(),
		"Images sent: "+s.ImageCount.String(),
		"",
		"Image URLs:",
	)
	if len(s.ImageURLs) > 0 {
		for _, u := range s.ImageURLs {
			out = append(out, "- "+u)
		}
	} else {
		out = append(out, "- none stored")
	}
	out = append(out, "", "User prompt:", s.UserPromptText)
	return strings.Join(out, "\n")
}
