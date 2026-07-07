// Package testutil provides small helpers for use in tests.
package testutil

import (
	"context"
	"testing"
	"time"
)

const (
	minCleanupSafety = time.Second * 30
	maxCleanupSafety = time.Minute * 5
)

// Context returns a context configured for the test deadline, reserving a
// fraction of the remaining time (bounded by minCleanupSafety and
// maxCleanupSafety) so cleanup tasks can finish before the test times out.
func Context(t *testing.T) context.Context {
	ctx := t.Context()
	if deadline, ok := t.Deadline(); ok {
		timeoutDuration := time.Until(deadline)
		cleanupSafety := min(max(timeoutDuration/4, minCleanupSafety), maxCleanupSafety)
		var cancel func()
		ctx, cancel = context.WithDeadline(ctx, deadline.Add(-1*cleanupSafety))
		t.Cleanup(cancel)
	}
	return ctx
}
