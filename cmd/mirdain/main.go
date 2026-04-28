package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"path/filepath"
	"strings"

	"github.com/gorilla/websocket"
	"github.com/tom-molotnikoff/mirdain-ai/internal/config"
	"github.com/tom-molotnikoff/mirdain-ai/internal/registry"
	"github.com/tom-molotnikoff/mirdain-ai/internal/runner"
	"github.com/tom-molotnikoff/mirdain-ai/internal/tracker"
	"github.com/tom-molotnikoff/mirdain-ai/ui"
)

// wsUpgrader upgrades HTTP connections to WebSocket. CheckOrigin is permissive
// because the server binds to 127.0.0.1 only, so ambient network access is
// already constrained.
var wsUpgrader = websocket.Upgrader{
	ReadBufferSize:  4096,
	WriteBufferSize: 4096,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

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

	reg := registry.New()
	agentRunner := runner.NewLocalDocker()

	mux := http.NewServeMux()
	mux.HandleFunc("GET /api/issues", issuesHandler(gh))
	mux.HandleFunc("POST /api/runs", createRunHandler(reg, agentRunner, cfg.Server.Port))
	mux.HandleFunc("GET /api/runs/{id}", getRunHandler(reg))
	mux.HandleFunc("GET /ws/{run_id}", uiWSHandler(reg))
	mux.HandleFunc("GET /internal/agent/{run_id}", agentWSHandler(reg))
	mux.Handle("/", spaHandler(distFS))

	addr := fmt.Sprintf("%s:%d", cfg.Server.Bind, cfg.Server.Port)
	log.Printf("mirdain listening on http://%s", addr)
	log.Fatal(http.ListenAndServe(addr, mux))
}

// newID returns a random 128-bit hex string suitable for use as a run ID or run secret.
func newID() string {
	b := make([]byte, 16)
	if _, err := rand.Read(b); err != nil {
		panic("crypto/rand: " + err.Error())
	}
	return hex.EncodeToString(b)
}

// createRunHandler handles POST /api/runs.
func createRunHandler(reg *registry.Registry, ar runner.AgentRunner, port int) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var req struct {
			IssueID string `json:"issue_id"`
		}
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil || req.IssueID == "" {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusBadRequest)
			json.NewEncoder(w).Encode(map[string]string{"error": "issue_id is required"})
			return
		}

		runID := newID()
		runSecret := newID()

		// Create the registry entry before starting the container — the agent
		// needs it to authenticate its WS connection immediately on startup.
		reg.Create(runID, runSecret, req.IssueID)

		orchestratorURL := fmt.Sprintf("ws://127.0.0.1:%d", port)
		if err := ar.Start(r.Context(), runner.RunConfig{
			RunID:           runID,
			RunSecret:       runSecret,
			IssueID:         req.IssueID,
			OrchestratorURL: orchestratorURL,
		}); err != nil {
			reg.Remove(runID)
			log.Printf("POST /api/runs: %v", err)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusInternalServerError)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		json.NewEncoder(w).Encode(map[string]string{
			"run_id": runID,
			"status": "running",
		})
	}
}

// getRunHandler handles GET /api/runs/{id}.
func getRunHandler(reg *registry.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		runID := r.PathValue("id")
		run, ok := reg.Get(runID)
		if !ok {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]string{"error": "run not found"})
			return
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{
			"run_id": run.RunID,
			"status": run.StatusString(),
		})
	}
}

// agentWSHandler handles GET /internal/agent/{run_id}.
// It authenticates the agent via ?secret=, then reads events and publishes
// them to the run's event bus. On run.completed, it terminates the run.
func agentWSHandler(reg *registry.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		runID := r.PathValue("run_id")
		secret := r.URL.Query().Get("secret")

		run, ok := reg.Get(runID)
		if !ok {
			http.Error(w, "run not found", http.StatusNotFound)
			return
		}
		if secret != run.Secret {
			http.Error(w, "invalid or missing run secret", http.StatusUnauthorized)
			return
		}

		conn, err := wsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("agent WS upgrade %s: %v", runID, err)
			return
		}
		defer conn.Close()

		for {
			_, data, err := conn.ReadMessage()
			if err != nil {
				break
			}

			// Parse just the type field to detect run.completed.
			var envelope struct {
				Type string `json:"type"`
			}
			if jsonErr := json.Unmarshal(data, &envelope); jsonErr != nil {
				log.Printf("agent WS %s: invalid JSON: %v", runID, jsonErr)
				continue
			}

			run.Publish(data)

			if envelope.Type == "run.completed" {
				reg.Terminate(runID)
				return
			}
		}

		// Connection dropped without run.completed — terminate the run.
		reg.Terminate(runID)
	}
}

// uiWSHandler handles GET /ws/{run_id}.
// It streams all run events (buffered replay + live) to the connected UI client.
func uiWSHandler(reg *registry.Registry) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		runID := r.PathValue("run_id")

		run, ok := reg.Get(runID)
		if !ok {
			http.Error(w, "run not found", http.StatusNotFound)
			return
		}

		conn, err := wsUpgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Printf("UI WS upgrade %s: %v", runID, err)
			return
		}
		defer conn.Close()

		snapshot, ch := run.Subscribe()
		defer run.Unsubscribe(ch)

		// Replay buffered events so late-connecting UIs don't miss anything.
		for _, data := range snapshot {
			if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
				return
			}
		}

		// If the run is already terminated, nothing more to stream.
		select {
		case <-run.Done():
			return
		default:
		}

		// Stream live events until the run terminates or the client disconnects.
		for {
			select {
			case data := <-ch:
				if err := conn.WriteMessage(websocket.TextMessage, data); err != nil {
					return
				}
			case <-run.Done():
				// Drain any events published just before termination.
				for {
					select {
					case data := <-ch:
						conn.WriteMessage(websocket.TextMessage, data) //nolint:errcheck
					default:
						return
					}
				}
			case <-r.Context().Done():
				return
			}
		}
	}
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
