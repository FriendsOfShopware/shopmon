package shopware

import "github.com/friendsofshopware/shopmon/api/internal/ptr"

// ContentLanguage normalizes an optional store language to a supported code.
// Unsupported or missing values default to English; callers already fall back
// to English for any field missing in the requested language.
func ContentLanguage(language *string) string {
	if ptr.Deref(language) == "de" {
		return "de"
	}
	return "en"
}
