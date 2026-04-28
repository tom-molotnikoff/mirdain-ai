// Package registry holds the in-memory run state for all active and recently
// completed agent runs. Entries survive only for the lifetime of the
// orchestrator process; they are reconciled from the tracker on restart.
package registry

import "sync"

const maxEventBuffer = 1000

// RunState holds the in-memory state for a single agent run.
type RunState struct {
	// Immutable after creation.
	RunID   string
	Secret  string
	IssueID string

	done chan struct{} // closed exactly once when the run terminates

	mu     sync.Mutex
	events [][]byte    // buffered raw JSON events, capped at maxEventBuffer
	subs   []chan []byte // active UI subscriber channels
}

// Publish appends data to the event buffer and fans it out to all active
// UI subscriber channels. Slow subscribers are skipped (non-blocking send)
// so a stalled UI connection never blocks agent ingest.
func (rs *RunState) Publish(data []byte) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	if len(rs.events) < maxEventBuffer {
		rs.events = append(rs.events, data)
	}
	for _, ch := range rs.subs {
		select {
		case ch <- data:
		default:
		}
	}
}

// Subscribe returns a snapshot of all buffered events and a live channel for
// new events. The caller must call Unsubscribe when it is done reading.
func (rs *RunState) Subscribe() ([][]byte, chan []byte) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	ch := make(chan []byte, 64)
	snapshot := make([][]byte, len(rs.events))
	copy(snapshot, rs.events)
	rs.subs = append(rs.subs, ch)
	return snapshot, ch
}

// Unsubscribe removes the subscriber channel from the run's fan-out list.
// The channel is not closed here; callers must handle the case where the
// done channel fires instead.
func (rs *RunState) Unsubscribe(ch chan []byte) {
	rs.mu.Lock()
	defer rs.mu.Unlock()
	for i, s := range rs.subs {
		if s == ch {
			rs.subs = append(rs.subs[:i], rs.subs[i+1:]...)
			return
		}
	}
}

// Done returns a channel that is closed when the run terminates. Callers can
// select on this channel alongside their subscriber channel to detect end-of-stream.
func (rs *RunState) Done() <-chan struct{} {
	return rs.done
}

// IsTerminated reports whether the run has terminated.
func (rs *RunState) IsTerminated() bool {
	select {
	case <-rs.done:
		return true
	default:
		return false
	}
}

// StatusString returns "terminated" or "running".
func (rs *RunState) StatusString() string {
	if rs.IsTerminated() {
		return "terminated"
	}
	return "running"
}

// Registry is the in-memory store of all runs known to the orchestrator.
type Registry struct {
	mu   sync.RWMutex
	runs map[string]*RunState
}

// New returns an empty Registry.
func New() *Registry {
	return &Registry{runs: make(map[string]*RunState)}
}

// Create inserts a new run with status "running" and returns it.
func (r *Registry) Create(runID, secret, issueID string) *RunState {
	rs := &RunState{
		RunID:   runID,
		Secret:  secret,
		IssueID: issueID,
		done:    make(chan struct{}),
	}
	r.mu.Lock()
	r.runs[runID] = rs
	r.mu.Unlock()
	return rs
}

// Get retrieves a run by ID. Returns (nil, false) if not found.
func (r *Registry) Get(runID string) (*RunState, bool) {
	r.mu.RLock()
	rs, ok := r.runs[runID]
	r.mu.RUnlock()
	return rs, ok
}

// Terminate marks a run as terminated by closing its done channel.
// Safe to call multiple times; only the first call closes the channel.
func (r *Registry) Terminate(runID string) {
	r.mu.RLock()
	rs, ok := r.runs[runID]
	r.mu.RUnlock()
	if !ok {
		return
	}
	rs.mu.Lock()
	defer rs.mu.Unlock()
	select {
	case <-rs.done:
		// already terminated
	default:
		close(rs.done)
	}
}

// Remove deletes a run from the registry. Used to roll back a failed launch.
func (r *Registry) Remove(runID string) {
	r.mu.Lock()
	delete(r.runs, runID)
	r.mu.Unlock()
}
