// Package proto defines the versioned JSON event vocabulary exchanged over
// WebSocket between the orchestrator and agents, and between the orchestrator
// and the React UI. All messages carry a schema version field (V) so later
// issues can extend the vocabulary without a breaking rename.
//
// Agent → orchestrator (forwarded to UI): RunStarted, TextDelta, ToolCall,
//
//	ArtifactProduced, AwaitingInput, RunCompleted.
//
// Orchestrator → agent: ToolResult, UserMessage.
package proto

// Version is the WS event schema version included on every message.
const Version = 1

// EventType identifies the kind of a WS event.
type EventType string

const (
	EventRunStarted       EventType = "run.started"
	EventTextDelta        EventType = "text.delta"
	EventToolCall         EventType = "tool.call"
	EventToolResult       EventType = "tool.result"
	EventArtifactProduced EventType = "artifact.produced"
	EventAwaitingInput    EventType = "awaiting_input"
	EventRunCompleted     EventType = "run.completed"
	EventUserMessage      EventType = "user.message"
)

// BaseEvent is embedded in every WS event struct.
type BaseEvent struct {
	V     int       `json:"v"`
	Type  EventType `json:"type"`
	RunID string    `json:"run_id"`
	Ts    string    `json:"ts"`
}

// RunStartedEvent is emitted by the agent when a run begins.
type RunStartedEvent struct {
	BaseEvent
}

// TextDeltaEvent carries an incremental text chunk from the agent.
type TextDeltaEvent struct {
	BaseEvent
	Text string `json:"text"`
}

// ToolCallEvent is emitted by the agent to invoke a brokered orchestrator tool.
type ToolCallEvent struct {
	BaseEvent
	CallID string         `json:"call_id"`
	Tool   string         `json:"tool"`
	Args   map[string]any `json:"args"`
}

// ToolResultEvent is sent by the orchestrator in response to a ToolCallEvent.
type ToolResultEvent struct {
	BaseEvent
	CallID string         `json:"call_id"`
	OK     bool           `json:"ok"`
	Result map[string]any `json:"result,omitempty"`
	Error  string         `json:"error,omitempty"`
}

// ArtifactProducedEvent signals that the agent produced a named artifact.
// Implementation ships with the relevant phase issue; the type is defined here
// so all later issues extend the same schema.
type ArtifactProducedEvent struct {
	BaseEvent
	Name string `json:"name"`
	URI  string `json:"uri"`
}

// AwaitingInputEvent signals that the agent is paused waiting for user input.
// Implementation ships with the HITL phase (#7); defined here for schema
// completeness.
type AwaitingInputEvent struct {
	BaseEvent
}

// RunCompletedEvent is emitted by the agent when it exits.
type RunCompletedEvent struct {
	BaseEvent
	ExitCode int `json:"exit_code"`
}

// UserMessageEvent is sent by the orchestrator to deliver a user message to a
// waiting HITL agent. Implementation ships with #7.
type UserMessageEvent struct {
	BaseEvent
	Text string `json:"text"`
}
