package e2e_test

import "testing"

// TestTracerE2E validates the full orchestrator↔agent↔UI pipeline.
//
// TODO(#2): implement once the tracer skill, AgentRunner, and Tracker are wired.
// The test should:
//  1. Start the orchestrator against a fixture GitHub repo.
//  2. Verify GET /api/issues returns mirdain-labelled issues.
//  3. POST /api/runs to start a tracer run.
//  4. Connect to WS /ws/{run_id} and assert run.started, text.delta,
//     and run.completed arrive in order within 500ms of each event's ts.
//  5. Verify mirdain.add_comment posted a comment on the fixture issue
//     via `gh issue view --comments`.
//  6. Restart the orchestrator; assert GET /api/runs/{id} returns
//     {"status":"terminated"} within 30 seconds.
func TestTracerE2E(t *testing.T) {
	t.Skip("stub — TODO(#2)")
}
