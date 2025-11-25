package main

import (
	"context"
	"fmt"
	"time"

	supervisor "go.fergus.london/go-supervise"
)

// printerActor demonstrates a simple actor that prints messages until it
// receives a stop control message or its context is cancelled.
type printerActor struct {
	mailbox chan supervisor.Envelope
}

func newPrinterActor() *printerActor {
	return &printerActor{mailbox: make(chan supervisor.Envelope, 4)}
}

func (a *printerActor) Mailbox() <-chan supervisor.Envelope {
	return a.mailbox
}

func (a *printerActor) Handle(ctx context.Context, msg interface{}) {
	fmt.Println("received:", msg)
}

func (a *printerActor) Terminate(ctx context.Context) {
	fmt.Println("actor terminating")
}

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	actor := newPrinterActor()
	worker := supervisor.ActorWorker(actor)
	s := supervisor.NewSimpleSupervisor(ctx, worker)

	s.Run()

	actor.mailbox <- supervisor.Envelope{Payload: "hello"}
	actor.mailbox <- supervisor.Envelope{Payload: "world"}

	time.Sleep(100 * time.Millisecond)

	// Stop the actor via a control message; Supervisor will finish once the
	// worker exits.
	actor.mailbox <- supervisor.Envelope{Control: supervisor.MessageStop}
	time.Sleep(100 * time.Millisecond)
}
