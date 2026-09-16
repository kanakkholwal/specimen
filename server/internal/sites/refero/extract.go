package refero

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/kanak/design-supply/internal/flight"
	"github.com/kanak/design-supply/internal/normalize"
)

// PageData is everything one style page yields.
type PageData struct {
	// Self is the page's own style, present only on /style/<uuid> pages.
	Self *Style
	// Cards are the related styles the page embeds. A style page never carries its own
	// card, so a card for X is only ever found on other pages that link to X.
	Cards []Card
}

// ExtractPage pulls the result object and every embedded card out of a flight document.
func ExtractPage(doc *flight.Doc, styleID string) (*PageData, error) {
	out := &PageData{}
	for _, c := range extractCards(doc) {
		if c.ID == "" {
			continue
		}
		out.Cards = append(out.Cards, c)
	}
	if styleID == "" {
		return out, nil
	}
	self, err := extractSelf(doc, styleID)
	if err != nil {
		return out, err
	}
	out.Self = self
	return out, nil
}

var errNoResult = errors.New("refero: no result object in payload")

// extractSelf reads the props object that carries both the extraction result and the
// page's own media URLs, which sit as siblings of result rather than inside it.
func extractSelf(doc *flight.Doc, styleID string) (*Style, error) {
	host, ok := doc.FindObject("result", "screenshotUrl")
	if !ok {
		if host, ok = doc.FindObject("meta", "raw", "designSystem"); !ok {
			return nil, errNoResult
		}
		host = map[string]any{"result": host}
	}
	rawResult, ok := host["result"]
	if !ok {
		return nil, errNoResult
	}
	var res Result
	if err := flight.Decode(rawResult, &res); err != nil {
		return nil, fmt.Errorf("refero: decode result: %w", err)
	}
	if res.Meta.URL == "" {
		return nil, errNoResult
	}

	canonical, err := json.Marshal(rawResult)
	if err != nil {
		return nil, err
	}
	sum := sha256.Sum256(canonical)

	origin, err := normalize.Identify(res.Meta.URL)
	if err != nil {
		return nil, fmt.Errorf("refero: identify %q: %w", res.Meta.URL, err)
	}

	card := Card{
		ID:                          styleID,
		URL:                         res.Meta.URL,
		SiteName:                    res.Meta.SiteName,
		ScreenshotURL:               str(host["screenshotUrl"]),
		ScreenshotAlt:               str(host["screenshotAlt"]),
		ThumbnailURL:                res.Screenshot.Thumbnail,
		PreviewVideoURL:             str(host["previewVideoUrl"]),
		PreviewVideoPosterURL:       str(host["previewVideoPosterUrl"]),
		PreviewVideoDetailURL:       str(host["previewVideoDetailUrl"]),
		PreviewVideoDetailPosterURL: str(host["previewVideoDetailPosterUrl"]),
		NorthStar:                   res.DesignSystem.NorthStar,
	}
	if card.ScreenshotURL == "" {
		card.ScreenshotURL = res.Screenshot.URL
	}

	return &Style{
		ID:        styleID,
		Card:      card,
		Result:    res,
		Origin:    origin.Origin,
		ETLD1:     origin.ETLD1,
		Subdomain: origin.Subdomain,
		ResultSHA: hex.EncodeToString(sum[:]),
	}, nil
}

// extractCards finds every related-style card. Cards are identified by carrying an id
// alongside a siteName and a screenshot, which no other object in the payload does.
func extractCards(doc *flight.Doc) []Card {
	objs := doc.FindObjects("id", "id", "siteName", "screenshotUrl")
	cards := make([]Card, 0, len(objs))
	for _, o := range objs {
		var c Card
		if err := flight.Decode(o, &c); err != nil {
			continue
		}
		if c.ID == "" || c.URL == "" {
			continue
		}
		cards = append(cards, c)
	}
	return cards
}

func str(v any) string {
	s, _ := v.(string)
	return s
}

// RawResultJSON re-emits the result object for archival with reference rows expanded and
// upstream key order kept, because the exports are sensitive to that order.
func RawResultJSON(doc *flight.Doc) ([]byte, error) {
	raw, ok := doc.RawValue("result")
	if !ok {
		return nil, errNoResult
	}
	return doc.ResolveJSON(raw, 2)
}
