package webui

import (
	"io/fs"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"testing/fstest"
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

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	if !strings.Contains(rec.Body.String(), "frontend") {
		t.Fatalf("expected frontend index, got %q", rec.Body.String())
	}
}

func TestNewHandlerServesStaticAssets(t *testing.T) {
	handler := NewHandler(testFS())

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/assets/app.js", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	if cacheControl := rec.Header().Get("Cache-Control"); !strings.Contains(cacheControl, "immutable") {
		t.Fatalf("expected immutable cache header, got %q", cacheControl)
	}
}

func TestNewHandlerCompressesAssets(t *testing.T) {
	handler := NewHandler(testFS())

	for _, encoding := range []string{"zstd", "gzip"} {
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/assets/large.js", nil)
		req.Header.Set("Accept-Encoding", encoding)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusOK {
			t.Fatalf("%s: expected 200, got %d", encoding, rec.Code)
		}

		if got := rec.Header().Get("Content-Encoding"); got != encoding {
			t.Fatalf("%s: expected Content-Encoding %q, got %q", encoding, encoding, got)
		}

		if rec.Header().Get("Cache-Control") == "" {
			t.Fatalf("%s: expected cache headers to survive compression", encoding)
		}
	}
}

func TestNewHandlerSkipsTinyAssets(t *testing.T) {
	handler := NewHandler(testFS())

	req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, "/assets/app.js", nil)
	req.Header.Set("Accept-Encoding", "gzip, zstd")
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	// app.js is below MinSize, so it must be served uncompressed.
	if got := rec.Header().Get("Content-Encoding"); got != "" {
		t.Fatalf("expected no Content-Encoding for tiny asset, got %q", got)
	}
}

func TestNewHandlerDoesNotHijackAPIOrMissingAssets(t *testing.T) {
	handler := NewHandler(testFS())

	for _, requestPath := range []string{"/api/health", "/assets/missing.js"} {
		req := httptest.NewRequestWithContext(t.Context(), http.MethodGet, requestPath, nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("%s: expected 404, got %d", requestPath, rec.Code)
		}
	}
}
