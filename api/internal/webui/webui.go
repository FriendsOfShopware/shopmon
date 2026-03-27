package webui

import (
	"io/fs"
	"mime"
	"net/http"
	"path"
	"strings"
)

func Available() bool {
	_, ok := FS()
	return ok
}

func FS() (fs.FS, bool) {
	embedded, ok := embeddedFS()
	if !ok {
		return nil, false
	}

	dist, err := fs.Sub(embedded, "dist")
	if err != nil {
		return nil, false
	}

	if _, err := fs.Stat(dist, "index.html"); err != nil {
		return nil, false
	}

	return dist, true
}

func Handler() (http.Handler, bool) {
	dist, ok := FS()
	if !ok {
		return nil, false
	}

	return NewHandler(dist), true
}

func NewHandler(dist fs.FS) http.Handler {
	files := http.FileServerFS(dist)

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "" {
			r.URL.Path = "/"
		}

		if strings.HasPrefix(r.URL.Path, "/api/") || r.URL.Path == "/api" {
			http.NotFound(w, r)
			return
		}

		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			http.NotFound(w, r)
			return
		}

		w.Header().Set("X-Frame-Options", "SAMEORIGIN")
		w.Header().Set("X-Content-Type-Options", "nosniff")
		w.Header().Set("X-XSS-Protection", "1; mode=block")
		w.Header().Set("Referrer-Policy", "strict-origin-when-cross-origin")
		w.Header().Set("Permissions-Policy", "camera=(), microphone=(), geolocation=()")

		assetPath := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
		if assetPath == "." || assetPath == "" {
			serveIndex(w, r, dist)
			return
		}

		if asset, err := fs.Stat(dist, assetPath); err == nil && !asset.IsDir() {
			setCacheHeaders(w, assetPath)
			files.ServeHTTP(w, r)
			return
		}

		if looksLikeAsset(assetPath) {
			http.NotFound(w, r)
			return
		}

		serveIndex(w, r, dist)
	})
}

func serveIndex(w http.ResponseWriter, r *http.Request, dist fs.FS) {
	w.Header().Set("Cache-Control", "no-cache")

	index, err := fs.ReadFile(dist, "index.html")
	if err != nil {
		http.NotFound(w, r)
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	if r.Method == http.MethodHead {
		return
	}
	_, _ = w.Write(index)
}

func setCacheHeaders(w http.ResponseWriter, assetPath string) {
	if strings.HasPrefix(assetPath, "assets/") {
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
		return
	}

	switch mime.TypeByExtension(path.Ext(assetPath)) {
	case "text/css; charset=utf-8", "application/javascript", "image/svg+xml", "image/x-icon":
		w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
	default:
		w.Header().Set("Cache-Control", "public, max-age=3600")
	}
}

func looksLikeAsset(assetPath string) bool {
	base := path.Base(assetPath)
	return strings.Contains(base, ".")
}
