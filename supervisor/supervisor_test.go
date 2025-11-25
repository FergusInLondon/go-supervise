package supervisor

import (
	"context"
	"testing"
	"time"

	"go.uber.org/goleak"
)

type mockSupervisable struct {
	nCalls      int
	nPanics     int
	shouldPanic bool
	ctxStopped  bool
	isRunning   bool
}

func generateSupervisable(ms *mockSupervisable) Supervisable {
	ms.nCalls = 0
	ms.nPanics = 0
	return func(ctx context.Context) {
		defer func() {
			if recover() != nil {
				// test == nothing to do
			}
			ms.isRunning = false
		}()

		ms.isRunning = true
		ms.nCalls++

		for {
			select {
			case <-ctx.Done():
				ms.ctxStopped = true
				return
			case <-time.After(50 * time.Millisecond):
				if ms.shouldPanic {
					ms.nPanics++
					panic("testing")
				}
			}
		}
	}
}

//
// These tests monitor the basic functionality, but there's also a little
// bit of magic behind the scenes in that we're also testing for leaking
// goroutines.
//

func Test_SupervisorMustTerminateWhenStopped(t *testing.T) {
	defer goleak.VerifyNone(t)

	ms := &mockSupervisable{}
	s, _ := NewSupervisorWithOptions(
		context.Background(),
		WithWorkers(SupervisableWorker{
			Func:  generateSupervisable(ms),
			Count: 1,
		}),
	)

	isUnblocked := false
	go func() {
		s.Run()
		isUnblocked = true
	}()

	<-time.After(time.Millisecond * 100)
	s.Stop()
	<-time.After(time.Millisecond * 100)

	if !isUnblocked {
		t.Error("call to Stop should prevent Run from blocking")
	}

	if !ms.ctxStopped {
		t.Error("call to Stop should result in context cancellation")
	}

	if ms.isRunning {
		t.Error("call to Stop should ensure goroutine has terminated")
	}

	if !(ms.nCalls >= 1) {
		t.Error("supervisable not called")
	}

	if s.CurrentWorkerCount() != 0 {
		t.Error("supervisor indicates it's still running")
	}
}

func Test_SupervisorMustRestartWorkerFollowingPanic(t *testing.T) {
	defer goleak.VerifyNone(t)

	ms := &mockSupervisable{
		shouldPanic: true,
	}
	s, _ := NewSupervisorWithOptions(
		context.Background(),
		WithWorkers(SupervisableWorker{
			Func:  generateSupervisable(ms),
			Count: 1,
		}),
	)

	s.Run()

	<-time.After(time.Millisecond * 100)
	s.Stop()
	<-time.After(time.Millisecond * 100)

	if !(ms.nCalls >= 1) {
		t.Error("supervisable not called")
	}

	// ms.nCalls = ms.nPanics + initial call
	if !((ms.nCalls - ms.nPanics) < 2) {
		t.Error("supervisable did not restart after each panic", ms.nCalls, ms.nPanics)
	}
}

func Test_SupervisorMustNotifyCallerWithWaitGroup(t *testing.T) {
	defer goleak.VerifyNone(t)

	ms := &mockSupervisable{}

	s, _ := NewSupervisorWithOptions(
		context.Background(),
		WithWorkers(SupervisableWorker{
			Func:  generateSupervisable(ms),
			Count: 1,
		}),
	)
	s.Run()

	wgComplete := false
	go func() {
		s.Wait()
		wgComplete = true
	}()

	<-time.After(time.Millisecond * 100)
	s.Stop()
	<-time.After(time.Millisecond * 100)

	if !(ms.nCalls >= 1) {
		t.Error("supervisable not called")
	}

	if !wgComplete {
		t.Error("waitgroup never completed")
	}
}

func Test_SupervisorShouldRestartWhenRequested(t *testing.T) {
	defer goleak.VerifyNone(t)

	ms := &mockSupervisable{}
	s, _ := NewSupervisorWithOptions(
		context.Background(),
		WithWorkers(SupervisableWorker{
			Func:  generateSupervisable(ms),
			Count: 1,
		}),
	)

	s.Run()

	<-time.After(time.Millisecond * 100)
	s.Restart()
	<-time.After(time.Millisecond * 100)
	s.Stop()
	<-time.After(time.Millisecond * 100)

	if !(ms.nCalls == 2) {
		t.Error("supervisable not restarted", ms.nCalls)
	}
}
