package webui

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func testFS() fs.FS {
	return fstest.MapFS{
		"index.html":              {Data: []byte("<html>frontend</html>")},
		"assets/app.js":           {Data: []byte("console.log('ok')")},
		"assets/large.js":         {Data: []byte(strings.Repeat("console.log('ok');\n", 512))},
		"favicon.ico":             {Data: []byte("ico")},
		"site.webmanifest":        {Data: []byte("{}")},
		"nested/route/index.html": {Data: []byte("<html>nested</html>")},
	}
}

func TestNewHandlerServesIndexForSPARoutes(t *testing.T) {
	handler := NewHandler(testFS())

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/app/dashboard", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	assert.Contains(t, rec.Body.String(), "frontend")
}

func TestNewHandlerServesStaticAssets(t *testing.T) {
	handler := NewHandler(testFS())

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/assets/app.js", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	require.Equal(t, http.StatusOK, rec.Code)

	assert.Contains(t, rec.Header().Get("Cache-Control"), "immutable")
}

func TestNewHandlerCompressesAssets(t *testing.T) {
	handler := NewHandler(testFS())

	for _, encoding := range []string{"zstd", "gzip"} {
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/assets/large.js", nil)
		req.Header.Set("Accept-Encoding", encoding)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		require.Equal(t, http.StatusOK, rec.Code, encoding)

		assert.Equal(t, encoding, rec.Header().Get("Content-Encoding"), encoding)

		assert.NotEmpty(t, rec.Header().Get("Cache-Control"), "%s: expected cache headers to survive compression", encoding)
	}
}

func TestNewHandlerSkipsTinyAssets(t *testing.T) {
	handler := NewHandler(testFS())

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/assets/app.js", nil)
	req.Header.Set("Accept-Encoding", "gzip, zstd")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// app.js is below MinSize, so it must be served uncompressed.
	assert.Empty(t, rec.Header().Get("Content-Encoding"), "expected no Content-Encoding for tiny asset")
}

func TestNewHandlerDoesNotHijackAPIOrMissingAssets(t *testing.T) {
	handler := NewHandler(testFS())

	for _, requestPath := range []string{"/api/health", "/assets/missing.js"} {
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, requestPath, nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		assert.Equal(t, http.StatusNotFound, rec.Code, requestPath)
	}
}
