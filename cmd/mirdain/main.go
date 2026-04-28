package main

import (
	"io/fs"
	"log"
	"net/http"
	"path/filepath"

	"github.com/tom-molotnikoff/mirdain-ai/ui"
)

func main() {
	// TODO(#6): load bind address and port from mirdain.yaml config.
	const addr = "127.0.0.1:7777"

	distFS, err := fs.Sub(ui.Files, "dist")
	if err != nil {
		log.Fatalf("failed to sub embedded UI fs: %v", err)
	}

	mux := http.NewServeMux()

	// TODO(#19,#20): mount /api/ routes once AgentRunner and Tracker are wired.

	mux.Handle("/", spaHandler(distFS))

	log.Printf("mirdain listening on http://%s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

// spaHandler serves static assets directly and falls back to index.html for
// all extensionless paths so React Router can handle client-side navigation.
func spaHandler(distFS fs.FS) http.Handler {
	fileServer := http.FileServerFS(distFS)
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Requests with a file extension are static assets (JS, CSS, images…).
		// Serve them directly; the file server returns 404 if missing.
		if filepath.Ext(r.URL.Path) != "" {
			fileServer.ServeHTTP(w, r)
			return
		}
		// All other paths are SPA routes. Serve index.html and let React Router
		// render the correct view based on the URL.
		r.URL.Path = "/"
		fileServer.ServeHTTP(w, r)
	})
}
