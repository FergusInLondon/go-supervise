// Package supervisor is a very simple implementation of the Supervisor pattern
// used in Erlang/OTP. It provides a mechanism for controlling/coordinating
// go-routines, and encourages the principle of failing early by ensuring the
// timely restart after any failures.
package supervisor

import (
	"context"
	"sync"
)

// Supervisable specifies the required signature of a Worker function. To
// correctly manage a Supervisable there are three requirements:
//
// 1. The Supervisable **must** handle context cancellation correctly;
//
// 2. The Supervisable **must** ensure that `recover()` is called.
type Supervisable func(context.Context)

type SupervisableWorker struct {
	Func  Supervisable
	Count int
}

// Supervisor is the basic Supervision Tree supervisor node. It's capable
// of monitoring a given goroutine and restarting it upon failure, as well
// as terminating or restarting it upon request.
type Supervisor struct {
	mtx            *sync.RWMutex
	workers        []SupervisableWorker
	ctx            context.Context
	stop           context.CancelFunc
	wg             *sync.WaitGroup
	runningWorkers int
}

func (s *Supervisor) incWorkerCount() {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	s.runningWorkers++
}

func (s *Supervisor) decWorkerCount() {
	s.mtx.Lock()
	defer s.mtx.Unlock()

	s.runningWorkers--
}

type Option func(*Supervisor) error

func WithWorkers(supervisables ...SupervisableWorker) Option {
	return func(s *Supervisor) error {
		s.workers = append(s.workers, supervisables...)
		return nil
	}
}

// NewSupervisorWithOptions configures a new Supervisor using any options
// specified by the Options struct.
func NewSupervisorWithOptions(ctx context.Context, opts ...Option) (*Supervisor, error) {
	supervisor := &Supervisor{
		mtx:     &sync.RWMutex{},
		workers: make([]SupervisableWorker, 0),
		wg:      &sync.WaitGroup{},
	}

	for _, opt := range opts {
		if err := opt(supervisor); err != nil {
			return nil, err
		}
	}

	supervisor.ctx, supervisor.stop = context.WithCancel(ctx)
	return supervisor, nil
}

// Run is the entrypoint for the supervisor; calling run will configure
// all the supplied Supervisables at the specified number of instances.
func (s *Supervisor) Run() {
	for _, worker := range s.workers {
		for i := 0; i < worker.Count; i++ {
			s.wg.Go(func() {
				s.incWorkerCount()
				defer s.decWorkerCount()

				for {
					worker.Func(s.ctx)

					if s.ctx.Err() != nil {
						break
					}
				}
			})
		}
	}
}

// Restart terminates the current worker goroutines, and then executes
// them again. This is a convenience wrapper around calling `Stop` and
// `Run` consecutively.
func (s *Supervisor) Restart() {
	s.Stop()
	defer s.Run()

	s.wg.Wait()
}

// Stop terminates any current goroutines by simply invoking the context
// cancellation function.
func (s *Supervisor) Stop() {
	s.stop()
}

// CurrentWorkerCount returns the number of current workers that are executing.
func (s *Supervisor) CurrentWorkerCount() int {
	s.mtx.RLock()
	defer s.mtx.RUnlock()

	return s.runningWorkers
}

// Wait blocks until all workers have completed running.
func (s *Supervisor) Wait() {
	s.wg.Wait()
}
