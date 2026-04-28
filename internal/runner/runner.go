package runner

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
	"sync"
)

// RunConfig holds the parameters for a single agent run.
type RunConfig struct {
	RunID           string
	RunSecret       string
	SkillPath       string
	IssueID         string
	RepoID          string
	OrchestratorURL string
}

// AgentRunner launches and stops containerised agent runs.
// Kubernetes, Nomad, and SSH+Docker are deferred to later versions.
type AgentRunner interface {
	Start(ctx context.Context, cfg RunConfig) error
	Stop(ctx context.Context, runID string) error
}

// LocalDocker implements AgentRunner using the local Docker daemon via the
// docker CLI. It avoids the Docker SDK to keep the dependency footprint small.
//
// Containers use --network=host so they can reach the orchestrator on
// 127.0.0.1. This works on Linux (the primary deployment target and CI).
// On macOS Docker Desktop, set OrchestratorURL to ws://host.docker.internal:{port}.
type LocalDocker struct {
	mu           sync.Mutex
	containerIDs map[string]string // runID → containerID
}

// NewLocalDocker returns a LocalDocker runner.
func NewLocalDocker() *LocalDocker {
	return &LocalDocker{containerIDs: make(map[string]string)}
}

// Start launches a mirdain-base container for the given run. The container
// receives MIRDAIN_RUN_ID, MIRDAIN_RUN_SECRET, and MIRDAIN_ORCHESTRATOR_URL
// as environment variables. No GitHub write credentials are injected.
func (d *LocalDocker) Start(ctx context.Context, cfg RunConfig) error {
	cmd := exec.CommandContext(ctx, "docker", "run", "-d",
		"--network=host",
		"-e", "MIRDAIN_RUN_ID="+cfg.RunID,
		"-e", "MIRDAIN_RUN_SECRET="+cfg.RunSecret,
		"-e", "MIRDAIN_ORCHESTRATOR_URL="+cfg.OrchestratorURL,
		"mirdain-base",
	)
	out, err := cmd.Output()
	if err != nil {
		var stderr string
		if exitErr, ok := err.(*exec.ExitError); ok {
			stderr = strings.TrimSpace(string(exitErr.Stderr))
		}
		if stderr != "" {
			return fmt.Errorf("docker run: %s", stderr)
		}
		return fmt.Errorf("docker unavailable: %w", err)
	}
	containerID := strings.TrimSpace(string(out))
	d.mu.Lock()
	d.containerIDs[cfg.RunID] = containerID
	d.mu.Unlock()
	return nil
}

// Stop terminates the container associated with the given run.
func (d *LocalDocker) Stop(ctx context.Context, runID string) error {
	d.mu.Lock()
	containerID, ok := d.containerIDs[runID]
	if ok {
		delete(d.containerIDs, runID)
	}
	d.mu.Unlock()
	if !ok {
		return nil
	}
	cmd := exec.CommandContext(ctx, "docker", "stop", containerID)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("docker stop %s: %w", containerID, err)
	}
	return nil
}
