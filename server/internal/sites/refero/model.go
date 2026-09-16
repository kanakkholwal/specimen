package refero

import (
	"encoding/json"
	"strings"
	"time"
)

// Result is the extraction payload Refero embeds in every style page.
type Result struct {
	Meta         Meta         `json:"meta"`
	Raw          Raw          `json:"raw"`
	DesignSystem DesignSystem `json:"designSystem"`
	Screenshot   Screenshot   `json:"screenshot"`
}

type Meta struct {
	URL          string          `json:"url"`
	SiteName     string          `json:"siteName"`
	ExtractedAt  string          `json:"extractedAt"`
	DurationMs   int64           `json:"durationMs"`
	Viewport     Size            `json:"viewport"`
	ElementCount int             `json:"elementCount"`
	Telemetry    json.RawMessage `json:"telemetry,omitempty"`
}

type Size struct {
	Width  int `json:"width"`
	Height int `json:"height"`
}

type Screenshot struct {
	URL       string `json:"url"`
	Thumbnail string `json:"thumbnail"`
}

// Raw is the measured layer: what the extractor found in the live DOM.
type Raw struct {
	Colors     RawColors     `json:"colors"`
	Shapes     RawShapes     `json:"shapes"`
	Spacing    RawSpacing    `json:"spacing"`
	Gradients  []Gradient    `json:"gradients"`
	Typography RawTypography `json:"typography"`
}

type RawColors struct {
	Tokens        []ColorToken   `json:"tokens"`
	IsDarkMode    bool           `json:"isDarkMode"`
	Colorfulness  float64        `json:"colorfulness"`
	ContrastPairs []ContrastPair `json:"contrastPairs"`
}

type ColorToken struct {
	Hex         string         `json:"hex"`
	OKLCH       OKLCH          `json:"oklch"`
	Contexts    []string       `json:"contexts"`
	Frequency   int            `json:"frequency"`
	Confidence  float64        `json:"confidence"`
	Prominence  float64        `json:"prominence"`
	Properties  []string       `json:"properties"`
	UsageCounts map[string]int `json:"usageCounts"`
}

type OKLCH struct {
	L float64 `json:"l"`
	C float64 `json:"c"`
	H float64 `json:"h"`
}

type ContrastPair struct {
	Level      string  `json:"level"`
	Ratio      float64 `json:"ratio"`
	Background string  `json:"background"`
	Foreground string  `json:"foreground"`
}

type RawShapes struct {
	Radii   []RadiusToken `json:"radii"`
	Shadows []ShadowToken `json:"shadows"`
}

type RadiusToken struct {
	Value     float64  `json:"value"`
	Contexts  []string `json:"contexts"`
	Frequency int      `json:"frequency"`
}

type ShadowToken struct {
	Value     string   `json:"value"`
	Contexts  []string `json:"contexts"`
	Frequency int      `json:"frequency"`
}

type RawSpacing struct {
	Tokens   []SpacingToken `json:"tokens"`
	Density  string         `json:"density"`
	BaseUnit *float64       `json:"baseUnit"`
}

type SpacingToken struct {
	Value      float64  `json:"value"`
	Contexts   []string `json:"contexts"`
	Frequency  int      `json:"frequency"`
	Properties []string `json:"properties"`
	// PropertyContextPairs entries are "property|context".
	PropertyContextPairs []string `json:"propertyContextPairs"`
}

type Gradient struct {
	Type      string   `json:"type"`
	Value     string   `json:"value"`
	Colors    []string `json:"colors"`
	Frequency int      `json:"frequency"`
}

type RawTypography struct {
	Fonts []Font     `json:"fonts"`
	Scale TypeScale  `json:"scale"`
	Steps []TypeStep `json:"steps"`
}

type Font struct {
	Family              string   `json:"family"`
	Source              string   `json:"source"` // system | google | custom
	Weights             []int    `json:"weights"`
	Frequency           int      `json:"frequency"`
	FontFeatureSettings []string `json:"fontFeatureSettings"`
}

type TypeScale struct {
	Base       float64 `json:"base"`
	Name       string  `json:"name"`
	Ratio      float64 `json:"ratio"`
	Confidence float64 `json:"confidence"`
}

type TypeStep struct {
	Size          float64  `json:"size"`
	Family        string   `json:"family"`
	Weight        int      `json:"weight"`
	Contexts      []string `json:"contexts"`
	Frequency     int      `json:"frequency"`
	LineHeight    float64  `json:"lineHeight"`
	LetterSpacing float64  `json:"letterSpacing"`
	TextTransform string   `json:"textTransform"`
}

// DesignSystem is the authored layer: named, described and prescriptive.
type DesignSystem struct {
	Description     string           `json:"description"`
	NorthStar       string           `json:"northStar"`
	NorthStarDetail string           `json:"northStarDetail"`
	Industry        string           `json:"industry"`
	Theme           string           `json:"theme"`
	Colors          []NamedColor     `json:"colors"`
	Surfaces        []Surface        `json:"surfaces"`
	Typography      []TypographyRole `json:"typography"`
	TypeScale       []TypeScaleRole  `json:"typeScale"`
	Spacing         SystemSpacing    `json:"spacing"`
	Elevation       []Elevation      `json:"elevation"`
	Components      []Component      `json:"components"`
	Layout          string           `json:"layout"`
	Imagery         string           `json:"imagery"`
	Dos             []string         `json:"dos"`
	Donts           []string         `json:"donts"`
	Similar         []SimilarBrand   `json:"similar"`
	// Older extractions carry a prose paragraph instead of an elevation list.
	ElevationPhilosophy string          `json:"elevationPhilosophy,omitempty"`
	CustomSections      []CustomSection `json:"customSections"`
}

type NamedColor struct {
	Hex      string `json:"hex"`
	Name     string `json:"name"`
	Role     string `json:"role"`
	Group    string `json:"group"`
	Gradient string `json:"gradient,omitempty"`
}

type Surface struct {
	Hex     string `json:"hex"`
	Name    string `json:"name"`
	Level   int    `json:"level"`
	Purpose string `json:"purpose"`
}

type TypographyRole struct {
	Role       string `json:"role"`
	Family     string `json:"family"`
	Sizes      string `json:"sizes"`
	Weight     string `json:"weight"`
	LineHeight string `json:"lineHeight"`
	// The next three are absent on roles the extractor inferred without AI description.
	Substitute          string `json:"substitute,omitempty"`
	LetterSpacing       string `json:"letterSpacing,omitempty"`
	FontFeatureSettings string `json:"fontFeatureSettings,omitempty"`
}

type TypeScaleRole struct {
	Role          string  `json:"role"`
	Size          float64 `json:"size"`
	LineHeight    float64 `json:"lineHeight"`
	LetterSpacing float64 `json:"letterSpacing"`
}

// SystemSpacing carries a fixed set of layout values plus a per-component radius map
// whose keys vary by style (cards, buttons, inputs, tabs, smallChips, ...).
type SystemSpacing struct {
	Radius       OrderedStrings `json:"radius"`
	ElementGap   string         `json:"elementGap"`
	SectionGap   string         `json:"sectionGap"`
	CardPadding  string         `json:"cardPadding"`
	PageMaxWidth string         `json:"pageMaxWidth"`
}

// Elevation names vary across extraction versions; the generators fall back to the
// older name/shadow spelling when element/style are absent.
type Elevation struct {
	Element string `json:"element"`
	Style   string `json:"style"`
	Name    string `json:"name,omitempty"`
	Shadow  string `json:"shadow,omitempty"`
}

type Component struct {
	Name        string `json:"name"`
	Role        string `json:"role"`
	Description string `json:"description"`
}

type SimilarBrand struct {
	Business string `json:"business"`
	Why      string `json:"why"`
}

type CustomSection struct {
	Title   string `json:"title"`
	Content string `json:"content"`
}

// Card is the summary object Refero embeds for a style and each of its related styles.
// Every media URL is recorded here whether or not the bytes are ever downloaded.
type Card struct {
	ID                          string          `json:"id"`
	URL                         string          `json:"url"`
	SiteName                    string          `json:"siteName"`
	ScreenshotURL               string          `json:"screenshotUrl"`
	ScreenshotAlt               string          `json:"screenshotAlt"`
	ThumbnailURL                string          `json:"thumbnailUrl"`
	IconURL                     string          `json:"iconUrl"`
	PreviewVideoURL             string          `json:"previewVideoUrl"`
	PreviewVideoPosterURL       string          `json:"previewVideoPosterUrl"`
	PreviewVideoWidth           int             `json:"previewVideoWidth"`
	PreviewVideoHeight          int             `json:"previewVideoHeight"`
	PreviewVideoDurationMs      int             `json:"previewVideoDurationMs"`
	PreviewVideoDetailURL       string          `json:"previewVideoDetailUrl"`
	PreviewVideoDetailPosterURL string          `json:"previewVideoDetailPosterUrl"`
	PreviewVideoDetailWidth     int             `json:"previewVideoDetailWidth"`
	PreviewVideoDetailHeight    int             `json:"previewVideoDetailHeight"`
	ColorScheme                 string          `json:"colorScheme"`
	Colors                      []CardColor     `json:"colors"`
	Fonts                       []string        `json:"fonts"`
	NorthStar                   string          `json:"northStar"`
	ManagementSignals           json.RawMessage `json:"managementSignals"`
	CreatedAt                   string          `json:"createdAt"`
}

type CardColor struct {
	Name     string `json:"name"`
	Hex      string `json:"hex"`
	Gradient string `json:"gradient"`
}

// Style is one archived record: the card, the full result, and derived identity.
type Style struct {
	ID        string    `json:"id"`
	Card      Card      `json:"card"`
	Result    Result    `json:"result"`
	Origin    string    `json:"origin"`
	ETLD1     string    `json:"etld1"`
	Subdomain string    `json:"subdomain"`
	SourceURL string    `json:"sourceUrl"`
	ResultSHA string    `json:"resultSha256"`
	FetchedAt time.Time `json:"fetchedAt"`
}

// ExtractedAt parses the extraction timestamp, which is RFC3339 with milliseconds.
func (r Result) ExtractedAt() (time.Time, bool) {
	return parseTime(r.Meta.ExtractedAt)
}

// CreatedAt parses the card timestamp, which is a space-separated SQL datetime.
func (c Card) CreatedTime() (time.Time, bool) { return parseTime(c.CreatedAt) }

func parseTime(s string) (time.Time, bool) {
	s = strings.TrimSpace(s)
	if s == "" {
		return time.Time{}, false
	}
	layouts := []string{
		time.RFC3339Nano,
		time.RFC3339,
		"2006-01-02 15:04:05.999999",
		"2006-01-02 15:04:05",
	}
	for _, l := range layouts {
		if t, err := time.Parse(l, s); err == nil {
			return t.UTC(), true
		}
	}
	return time.Time{}, false
}

// Collection is an editorial or category listing page that groups styles.
type Collection struct {
	Slug     string   `json:"slug"`
	URL      string   `json:"url"`
	Title    string   `json:"title"`
	Kind     string   `json:"kind"`
	StyleIDs []string `json:"styleIds"`
}
