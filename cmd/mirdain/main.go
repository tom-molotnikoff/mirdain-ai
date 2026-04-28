package main

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/tom-molotnikoff/mirdain-ai/internal/config"
	"github.com/tom-molotnikoff/mirdain-ai/internal/tracker"
	"github.com/tom-molotnikoff/mirdain-ai/ui"
)

func main() {
	cfg, err := config.Load("mirdain.yaml")
	if err != nil {
		log.Fatalf("config: %v", err)
	}

	if len(cfg.Repos) == 0 {
		log.Fatalf("config: repos list is empty — add at least one repo to mirdain.yaml")
	}
	if !cfg.Repos[0].Enabled {
		log.Fatalf("config: repos[0] is disabled — set enabled: true in mirdain.yaml")
	}

	gh, err := tracker.NewGitHub(cfg.GitHubApp.PAT, cfg.Repos[0].ID)
	if err != nil {
		log.Fatalf("tracker: %v", err)
	}

	distFS, err := fs.Sub(ui.Files, "dist")
	if err != nil {
		log.Fatalf("failed to sub embedded UI fs: %v", err)
	}

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/issues", issuesHandler(gh))
	// TODO(#20): mount remaining /api/ routes once AgentRunner and WS are wired.
	mux.Handle("/", spaHandler(distFS))

	addr := fmt.Sprintf("%s:%d", cfg.Server.Bind, cfg.Server.Port)
	log.Printf("mirdain listening on http://%s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

// issueResponse is the JSON shape for a single issue in GET /api/issues.
type issueResponse struct {
	Number        int    `json:"number"`
	Title         string `json:"title"`
	WorkflowLabel string `json:"workflow_label"`
}

// issuesHandler returns a handler for GET /api/issues.
func issuesHandler(t tracker.Tracker) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		issues, err := t.ListIssues(r.Context(), "mirdain")
		if err != nil {
			if errors.Is(err, tracker.ErrUnauthorized) {
				w.Header().Set("Content-Type", "application/json")
				w.WriteHeader(http.StatusUnauthorized)
				json.NewEncoder(w).Encode(map[string]string{"error": "unauthorized: check MIRDAIN_GITHUB_TOKEN"})
				return
			}
			log.Printf("GET /api/issues: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": "internal server error"})
			return
		}

		resp := make([]issueResponse, len(issues))
		for i, iss := range issues {
			resp[i] = issueResponse{
				Number:        iss.Number,
				Title:         iss.Title,
				WorkflowLabel: workflowLabel(iss.Labels),
			}
		}

		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}
}

// workflowLabel returns the first workflow:* label in labels, or "unlabelled".
// Logs a warning if multiple workflow:* labels are found (invalid state).
func workflowLabel(labels []string) string {
	var found []string
	for _, l := range labels {
		if strings.HasPrefix(l, "workflow:") {
			found = append(found, l)
		}
	}
	if len(found) == 0 {
		return "unlabelled"
	}
	if len(found) > 1 {
		log.Printf("warning: issue has multiple workflow:* labels %v; using first", found)
	}
	return found[0]
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
