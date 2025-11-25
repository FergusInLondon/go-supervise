package main

import (
	"context"
	"fmt"
	"time"

	"go.fergus.london/go-supervise/actor"
	"go.fergus.london/go-supervise/supervisor"
)

// printerActor demonstrates a simple actor that prints messages until it
// receives a stop control message or its context is cancelled.
type printerActor struct {
	mailbox chan actor.Envelope
}

func newPrinterActor() *printerActor {
	return &printerActor{mailbox: make(chan actor.Envelope, 4)}
}

func (a *printerActor) Mailbox() <-chan actor.Envelope {
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

	a := newPrinterActor()
	worker := actor.ActorWorker(a)
	s, _ := supervisor.NewSupervisorWithOptions(ctx, supervisor.WithWorkers(supervisor.SupervisableWorker{
		Func:  worker,
		Count: 2,
	}))

	s.Run()

	a.mailbox <- actor.Envelope{Payload: "hello"}
	a.mailbox <- actor.Envelope{Payload: "world"}

	time.Sleep(100 * time.Millisecond)

	// Stop the actor via a control message; Supervisor will finish once the
	// worker exits.
	a.mailbox <- actor.Envelope{Control: actor.MessageStop}
	time.Sleep(100 * time.Millisecond)
}
