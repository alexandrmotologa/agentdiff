package server

import (
	"embed"
	"io/fs"
	"net/http"
	"strings"
)

//go:embed all:dist
var embeddedFS embed.FS

// FileServerHandler returns an http.Handler that serves the embedded React 19 SPA.
func (s *Server) FileServerHandler() http.Handler {
	distFS, err := fs.Sub(embeddedFS, "dist")
	if err != nil {
		// Fallback to empty handler if sub fails
		return http.NotFoundHandler()
	}

	fileServer := http.FileServer(http.FS(distFS))

	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// API requests should already be handled by mux
		if strings.HasPrefix(r.URL.Path, "/api/") {
			http.NotFound(w, r)
			return
		}

		path := strings.TrimPrefix(r.URL.Path, "/")
		if path == "" {
			path = "index.html"
		}

		// Check if file exists in embedded dist
		f, err := distFS.Open(path)
		if err != nil {
			// SPA fallback to index.html for client-side routing
			indexData, err := fs.ReadFile(distFS, "index.html")
			if err != nil {
				http.NotFound(w, r)
				return
			}
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.Write(indexData)
			return
		}
		_ = f.Close()

		fileServer.ServeHTTP(w, r)
	})
}
