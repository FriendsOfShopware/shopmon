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
		"favicon.ico":             {Data: []byte("ico")},
		"site.webmanifest":        {Data: []byte("{}")},
		"nested/route/index.html": {Data: []byte("<html>nested</html>")},
	}
}

func TestNewHandlerServesIndexForSPARoutes(t *testing.T) {
	handler := NewHandler(testFS())

	req := httptest.NewRequest(http.MethodGet, "/app/dashboard", nil)
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

	req := httptest.NewRequest(http.MethodGet, "/assets/app.js", nil)
	rec := httptest.NewRecorder()

	handler.ServeHTTP(rec, req)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	if cacheControl := rec.Header().Get("Cache-Control"); !strings.Contains(cacheControl, "immutable") {
		t.Fatalf("expected immutable cache header, got %q", cacheControl)
	}
}

func TestNewHandlerDoesNotHijackAPIOrMissingAssets(t *testing.T) {
	handler := NewHandler(testFS())

	for _, requestPath := range []string{"/api/health", "/assets/missing.js"} {
		req := httptest.NewRequest(http.MethodGet, requestPath, nil)
		rec := httptest.NewRecorder()

		handler.ServeHTTP(rec, req)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("%s: expected 404, got %d", requestPath, rec.Code)
		}
	}
}
