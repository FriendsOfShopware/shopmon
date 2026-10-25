package sitespeed

import (
	"errors"
	"fmt"
	"net/url"
	"strings"
)

var ErrInvalidURL = errors.New("sitespeed URLs must be non-blank HTTP or HTTPS URLs without credentials")

// NormalizeURLs validates every supplied row. An empty list uses the environment URL.
func NormalizeURLs(values []string) ([]string, error) {
	normalized := make([]string, 0, len(values))
	for i, value := range values {
		value = strings.TrimSpace(value)
		parsed, err := url.Parse(value)
		if err != nil || (parsed.Scheme != "http" && parsed.Scheme != "https") || parsed.Hostname() == "" || parsed.User != nil {
			return nil, fmt.Errorf("%w (entry %d)", ErrInvalidURL, i+1)
		}
		normalized = append(normalized, value)
	}
	return normalized, nil
}
