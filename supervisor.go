// Package Supervisor is a very simple implementation of the Supervisor pattern
// used in Erlang/OTP. It provides a mechanism for controlling/coordinating
// go-routines, and encourages the principle of failing early by ensuring the
// timely restart after any failures.
package supervisor

import (
	"context"
	"sync"
	"time"
)

// Supervisable specifies the required signature of a Worker function. To
// correctly manage a Supervisable there are three requirements:
//
// 1. The Supervisable **must** handle context cancellation correctly;
//
// 2. The Supervisable **must** defer the close of `chan struct{}`;
//
// 3. The Supervisable **must** ensure that `recover()` is called.
type Supervisable func(context.Context, chan struct{})

// Supervisor is the basic Supervision Tree supervisor node. It's capable
// of monitoring a given goroutine and restarting it upon failure, as well
// as terminating or restarting it upon request.
type Supervisor struct {
	isSimple       bool
	workers        []Supervisable
	ctx            context.Context
	stop           context.CancelFunc
	wg             *sync.WaitGroup
	workerCount    int
	hasStopped     bool
	runningWorkers int
}

// NewSimpleSupervisor returns a supervisor which can only run a single
// instance of a single worker goroutine. For a lot of uses this will be
// enough.
func NewSimpleSupervisor(ctx context.Context, worker Supervisable) *Supervisor {
	supervisorCtx, cancel := context.WithCancel(ctx)
	return &Supervisor{
		isSimple: true,
		workers:  []Supervisable{worker},
		ctx:      supervisorCtx,
		stop:     cancel,
	}
}

// Options holds basic configuration information for the Supervisor.
type Options struct {
	// WorkerCount determines how many instances *of each* worker should
	// be executed.
	WorkerCount int
	// Workers is a slice of different Supervisable workers, these will
	// all be executed with WorkerCount instances
	Workers []Supervisable
	// Context allows a parent context.Context object to be used, useful
	// where there are external timeouts or cancellations that may occur
	// further up the call chain.
	Context context.Context
	// Waiter allows the caller to block until the Supervisor has completed.
	Waiter *sync.WaitGroup
}

// NewSupervisorWithOptions configures a new Supervisor using any options
// specified by the Options struct.
func NewSupervisorWithOptions(opts *Options) *Supervisor {
	ctx := opts.Context
	if ctx == nil {
		ctx = context.Background()
	}
	supervisorCtx, cancel := context.WithCancel(ctx)

	return &Supervisor{
		workers:     opts.Workers,
		workerCount: opts.WorkerCount,
		ctx:         supervisorCtx,
		stop:        cancel,
	}
}

// Run is the entrypoint for the supervisor; calling run will configure
// all the supplied Supervisables at the specified number of instances.
//
// A call to run **is blocking**.
func (s *Supervisor) Run() {
	for _, worker := range s.workers {
		go s.runLoop(worker)
	}
}

func (s *Supervisor) runLoop(worker Supervisable) {
	if s.wg != nil {
		s.wg.Add(1)
		defer s.wg.Done()
	}

	// BUG(): This is a quick hack, and should be handled via the WaitGroup
	// Just need to work out how to handle `.WithWaitGroup(sync.WaitGroup)`
	// calls that happen in conjunction with an internal pre-existing one.
	s.runningWorkers++
	defer func() {
		s.runningWorkers--
	}()

	for {
		isDone := make(chan struct{})
		go worker(s.ctx, isDone)

		<-isDone
		if s.ctx.Err() != nil {
			break
		}
	}
}

// Restart terminates the current worker goroutines, and then executes
// them again. This is a convenience wrapper around calling `Stop` and
// `Run` consecutively.
func (s *Supervisor) Restart() {
	s.Stop()
	defer s.Run()

	for {
		// @todo - come on, man. This isn't the way.
		<-time.After(time.Millisecond * 250)
		if s.hasStopped {
			return
		}
	}
}

// Stop terminates any current goroutines by simply invoking the context
// cancellation function.
func (s *Supervisor) Stop() {
	s.stop()
}

// HasStopped returns a boolean stating wheter the Supervisor is running.
func (s *Supervisor) HasStopped() bool {
	return s.hasStopped
}

// WithWaitGroup allows a WaitGroup to be specified and incremented
// for each Supervisable supplied; when the WaitGroup is Done this
// means that all Supervisables have completed for good, and there
// will be no attempt at restarting them.
func (s *Supervisor) WithWaitGroup(wg *sync.WaitGroup) {
	s.wg = wg
}
