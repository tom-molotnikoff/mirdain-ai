package runner

import "context"

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
// LocalDocker is the v1 implementation (TODO(#19)).
// Kubernetes, Nomad, and SSH+Docker are deferred to later versions.
type AgentRunner interface {
	Start(ctx context.Context, cfg RunConfig) error
	Stop(ctx context.Context, runID string) error
}
