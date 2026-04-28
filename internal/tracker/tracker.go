package tracker

import "context"

// Issue represents a single issue in the tracker backend.
type Issue struct {
	ID     string
	Number int
	Title  string
	Labels []string
	Body   string
}

// Tracker is the pluggable interface for the issue tracker backend.
// GitHub Issues is the v1 implementation (TODO(#20)).
// Jira, Linear, and Gerrit are deferred to later versions.
//
// The full verb set is specified in #1 (constitution). Only the verbs
// required by the tracer slice (#2) are listed here; later issues extend
// this interface.
type Tracker interface {
	ListIssues(ctx context.Context, label string) ([]Issue, error)
	GetIssue(ctx context.Context, id string) (Issue, error)
	AddComment(ctx context.Context, issueID, body string) (commentID string, err error)
}
