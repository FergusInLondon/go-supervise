package supervisor

import (
	"context"
	"fmt"
)

// ControlMessage denotes the control instruction associated with an Envelope.
// Control messages can be used to request shutdown or restart behaviour without
// conflating those signals with user payloads.
type ControlMessage int

const (
	// MessageData is the default control message indicating a user payload
	// should be processed by the Actor.
	MessageData ControlMessage = iota
	// MessageStop requests that the Actor stops gracefully.
	MessageStop
	// MessageRestart requests that the Actor restarts. Supervisors will
	// re-run the worker once it returns.
	MessageRestart
)

// Envelope wraps Actor messages, allowing control messages to be sent alongside
// user-defined payloads.
type Envelope struct {
	Control ControlMessage
	Payload interface{}
}

// Actor represents a message-driven worker that can be supervised. Actors
// expose a mailbox channel, a Handle function for processing messages, and may
// optionally implement Init and Terminate hooks.
type Actor interface {
	Mailbox() <-chan Envelope
	Handle(ctx context.Context, msg interface{})
}

// Initialiser allows Actors to run setup logic before processing begins.
type Initialiser interface {
	Init(ctx context.Context) error
}

// Terminator allows Actors to perform cleanup when the worker terminates.
type Terminator interface {
	Terminate(ctx context.Context)
}

// ActorWorker adapts an Actor to the Supervisable function signature, enabling
// actors to be supervised without altering the Supervisor core.
func ActorWorker(actor Actor) Supervisable {
	return func(ctx context.Context, done chan struct{}) {
		defer func() {
			if r := recover(); r != nil {
				log(fmt.Sprintf("recovered panic in actor: %v", r))
			}
			if terminator, ok := actor.(Terminator); ok {
				safeTerminate(ctx, terminator)
			}
			close(done)
		}()

		if initialiser, ok := actor.(Initialiser); ok {
			if err := initialiser.Init(ctx); err != nil {
				log(fmt.Sprintf("actor init failed: %v", err))
				return
			}
		}

                for {
                        select {
                        case <-ctx.Done():
                                return
                        case envelope, ok := <-actor.Mailbox():
                                if !ok {
					return
				}

                                switch envelope.Control {
                                case MessageRestart, MessageStop:
                                        // Returning here ends the current worker loop. When running
                                        // under a Supervisor the loop will be restarted unless the
                                        // supervisor context has been cancelled.
                                        return
                                default:
                                        actor.Handle(ctx, envelope.Payload)
                                }
                        }
                }
        }
}

func safeTerminate(ctx context.Context, terminator Terminator) {
	defer func() {
		if r := recover(); r != nil {
			log(fmt.Sprintf("recovered panic in actor termination: %v", r))
		}
	}()
	terminator.Terminate(ctx)
}
