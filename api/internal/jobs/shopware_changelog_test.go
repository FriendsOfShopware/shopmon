package jobs

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestParsePHPVersions(t *testing.T) {
	tests := []struct {
		name string
		body string
		want []string
	}{
		{
			name: "modern release with php list",
			body: "<h2>System requirements</h2>\n<ul>\n<li>tested on PHP <code>8.2</code>, <code>8.4</code> and <code>8.5</code></li>\n<li>tested on <code>MySQL 8</code> and <code>MariaDB 11</code></li>\n</ul>",
			want: []string{"8.2", "8.4", "8.5"},
		},
		{
			name: "single php version",
			body: "<li>tested on PHP <code>8.1</code></li>",
			want: []string{"8.1"},
		},
		{
			name: "older release without php info",
			body: "<h1>Changelog</h1><p>Some fixes</p>",
			want: []string{},
		},
		{
			name: "empty body",
			body: "",
			want: []string{},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, parsePHPVersions(tt.body))
		})
	}
}
