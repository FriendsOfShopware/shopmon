package httputil

import (
	"net/url"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestPHPFormArray(t *testing.T) {
	t.Parallel()
	for _, tt := range []struct {
		name  string
		body  string
		field string
		want  []string
	}{
		{"repeated", "tags[]=One&tags[]=Two", "tags", []string{"One", "Two"}},
		{"indexed", "ids[0]=a&ids[2]=b", "ids", []string{"a", "b"}},
		{"mixed", "ids[]=a&ids[1]=b", "ids", []string{"a", "b"}},
		{"encoded", "items%5B%5D=a%2Fb&items%5B1%5D=c+d", "items", []string{"a/b", "c d"}},
		{"duplicates and empty values", "tags[]=a&tags[]=a&tags[]=", "tags", []string{"a", "a", ""}},
		{"unrelated fields", "tags[]=a&other[]=b&prefix_tags[]=c", "tags", []string{"a"}},
		{"unsupported shapes", "tags=a&tags[key]=b&tags[0][name]=c&tags[-1]=d&tags[=e", "tags", nil},
		{"missing", "other[]=a", "tags", nil},
		{"empty", "", "tags", nil},
	} {
		t.Run(tt.name, func(t *testing.T) {
			form, err := url.ParseQuery(tt.body)
			require.NoError(t, err)
			assert.ElementsMatch(t, tt.want, PHPFormArray(form, tt.field))
		})
	}
}
