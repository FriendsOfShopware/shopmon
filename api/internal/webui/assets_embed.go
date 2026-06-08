package webui

import (
	"embed"
	"io/fs"
)

//go:embed all:dist
var distFS embed.FS

func embeddedFS() (fs.FS, bool) {
	return distFS, true
}
