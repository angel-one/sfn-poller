// Package pollableiface contains an interface for a PollableTask
package pollableiface

import "github.com/angel-one/sfn-poller/sfnpoller/cancellablecontext/cancellablecontextiface"

// PollableTask represents a thing that can poll.
type PollableTask interface {
	Start(cancellablecontextiface.Context)
	Started() <-chan struct{}
	Done() <-chan struct{}
}
