package notify

import (
	"embed"
	"encoding/json"
	"fmt"
	"log/slog"
	"path/filepath"
	"strings"
)

// DefaultLocale is used both as the seed locale and as the fallback whenever a
// recipient's locale or a specific key is missing.
const DefaultLocale = "en"

//go:embed locales/*.json
var localeFS embed.FS

// Translator renders translation keys with {param} interpolation. It mirrors the
// frontend vue-i18n catalog (same keys, same {name} placeholder syntax) so the
// wording stays consistent between in-app notifications and emails.
type Translator struct {
	// messages maps locale -> key -> template.
	messages map[string]map[string]string
}

// NewTranslator loads the embedded locale catalogs. It never returns an error:
// a malformed or missing catalog degrades to returning the raw key, which keeps
// notifications flowing rather than failing the scrape.
func NewTranslator() *Translator {
	t := &Translator{messages: map[string]map[string]string{}}

	entries, err := localeFS.ReadDir("locales")
	if err != nil {
		slog.Error("notify: failed to read embedded locales", "error", err)
		return t
	}

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".json") {
			continue
		}
		locale := strings.TrimSuffix(e.Name(), ".json")
		data, err := localeFS.ReadFile(filepath.Join("locales", e.Name()))
		if err != nil {
			slog.Error("notify: failed to read locale file", "file", e.Name(), "error", err)
			continue
		}
		var msgs map[string]string
		if err := json.Unmarshal(data, &msgs); err != nil {
			slog.Error("notify: failed to parse locale file", "file", e.Name(), "error", err)
			continue
		}
		t.messages[locale] = msgs
	}

	return t
}

// T renders key in the given locale, interpolating {param} placeholders. It
// falls back to DefaultLocale and finally to the raw key when a translation is
// missing.
func (t *Translator) T(locale, key string, params map[string]any) string {
	tmpl, ok := t.lookup(locale, key)
	if !ok {
		tmpl, ok = t.lookup(DefaultLocale, key)
	}
	if !ok {
		return key
	}
	return interpolate(tmpl, params)
}

// RenderCheck renders an environment check message in the given locale. It
// translates the key, degrades to the raw "snippet" param when the key is
// unknown (rather than showing a bare key), and appends the recommendation
// clause when both current/recommended params are present.
func (t *Translator) RenderCheck(locale, key string, params map[string]any) string {
	msg := t.T(locale, key, params)
	if msg == key {
		if s, ok := params["snippet"].(string); ok && s != "" {
			msg = s
		}
	}
	if nonEmptyParam(params, "current") && nonEmptyParam(params, "recommended") {
		msg += t.T(locale, "check.recommendationSuffix", params)
	}
	return msg
}

func nonEmptyParam(params map[string]any, k string) bool {
	v, ok := params[k]
	if !ok {
		return false
	}
	s, ok := v.(string)
	return ok && s != ""
}

func (t *Translator) lookup(locale, key string) (string, bool) {
	msgs, ok := t.messages[locale]
	if !ok {
		return "", false
	}
	tmpl, ok := msgs[key]
	return tmpl, ok
}

// interpolate replaces {key} tokens with the matching param value. Unknown
// tokens are left untouched so missing data is visible rather than silently
// dropped.
func interpolate(tmpl string, params map[string]any) string {
	if len(params) == 0 || !strings.Contains(tmpl, "{") {
		return tmpl
	}
	for k, v := range params {
		tmpl = strings.ReplaceAll(tmpl, "{"+k+"}", fmt.Sprintf("%v", v))
	}
	return tmpl
}
