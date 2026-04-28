// TODO(#21): mirdain-bridge Pi extension.
//
// This module connects to the orchestrator WebSocket and:
//   - Emits structured run events:
//       run.started, text.delta, tool.call, artifact.produced, run.completed
//   - Receives orchestrator messages:
//       tool.result (response to a brokered tool call)
//       user.message (HITL input, handled in #7)
//   - Executes brokered writes by sending tool.call messages and awaiting
//     tool.result — the agent has NO direct write credentials.
//
// Environment variables injected by the orchestrator at container start:
//   MIRDAIN_RUN_SECRET       — auth token for /internal/agent/{run_id}?secret=…
//   MIRDAIN_ORCHESTRATOR_URL — ws://host:port base URL

export {};
