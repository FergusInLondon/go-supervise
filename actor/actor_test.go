package actor

import (
	"context"
	"testing"
	"time"

	"go.uber.org/goleak"
)

type testActor struct {
	mailbox       chan Envelope
	handled       []any
	initCalled    int
	terminateCall int
	panicOnHandle bool
}

func (a *testActor) Mailbox() <-chan Envelope {
	return a.mailbox
}

func (a *testActor) Handle(ctx context.Context, msg any) {
	if a.panicOnHandle {
		panic("handle panic")
	}
	a.handled = append(a.handled, msg)
}

func (a *testActor) Init(ctx context.Context) error {
	a.initCalled++
	return nil
}

func (a *testActor) Terminate(ctx context.Context) {
	a.terminateCall++
	close(a.mailbox)
}

func TestActorWorkerProcessesMessagesAndStops(t *testing.T) {
	defer goleak.VerifyNone(t)

	actor := &testActor{mailbox: make(chan Envelope, 2)}
	worker := ActorWorker(actor)

	ctx, cancel := context.WithCancel(context.Background())
	go worker(ctx)

	actor.mailbox <- Envelope{Payload: "hello"}
	actor.mailbox <- Envelope{Control: MessageStop}

	<-time.After(time.Second)
	cancel()

	if len(actor.handled) != 1 {
		t.Fatalf("expected 1 message handled, got %d", len(actor.handled))
	}

	if actor.terminateCall != 1 {
		t.Fatalf("terminate should be called once, got %d", actor.terminateCall)
	}
}

func TestActorWorkerHandlesContextCancellation(t *testing.T) {
	defer goleak.VerifyNone(t)

	actor := &testActor{mailbox: make(chan Envelope)}
	worker := ActorWorker(actor)

	ctx, cancel := context.WithCancel(context.Background())
	go worker(ctx)

	<-time.After(500 * time.Millisecond)
	cancel()

	<-time.After(500 * time.Millisecond)
	if actor.terminateCall != 1 {
		t.Fatalf("terminate should be called after context cancellation, got %d", actor.terminateCall)
	}
}

func TestActorWorkerRecoversPanics(t *testing.T) {
	defer goleak.VerifyNone(t)

	actor := &testActor{mailbox: make(chan Envelope, 1), panicOnHandle: true}
	worker := ActorWorker(actor)

	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	go worker(ctx)

	actor.mailbox <- Envelope{Payload: "boom"}
	<-time.After(500 * time.Millisecond)
	cancel()

	<-time.After(500 * time.Millisecond)
	if actor.initCalled != 1 {
		t.Fatalf("init should be called before handling messages, got %d", actor.initCalled)
	}
}

func TestActorWorkerHandlesRestartMessage(t *testing.T) {
	defer goleak.VerifyNone(t)

	actor := &testActor{mailbox: make(chan Envelope, 1)}
	worker := ActorWorker(actor)

	ctx, cancel := context.WithCancel(context.Background())
	go worker(ctx)

	actor.mailbox <- Envelope{Control: MessageRestart}
	<-time.After(time.Second)
	cancel()

	if len(actor.handled) != 0 {
		t.Fatalf("restart message should not be passed to Handle, got %d", len(actor.handled))
	}
}
