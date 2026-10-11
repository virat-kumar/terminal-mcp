package terminal

import (
	_ "embed"
	"net/http"
	"strconv"
)

//go:embed favicon.png
var faviconPNG []byte

// FaviconHandler serves the embedded site icon for browser tabs.
func FaviconHandler() http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodGet && r.Method != http.MethodHead {
			w.Header().Set("Allow", "GET, HEAD")
			http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
			return
		}
		w.Header().Set("Content-Type", "image/png")
		w.Header().Set("Cache-Control", "public, max-age=86400")
		w.Header().Set("Content-Length", strconv.Itoa(len(faviconPNG)))
		if r.Method == http.MethodHead {
			return
		}
		_, _ = w.Write(faviconPNG)
	})
}
