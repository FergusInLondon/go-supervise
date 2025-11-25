/*
This example demonstrates a very simple Supervisor; there's no tree and only
one go-routine.

The go-routine is configured to increment a counter every 250ms, and then
panic every third iteration. It prints it's status to stdout, but as the
Supervisor keeps the go-routine executing, it looks uninterrupted.

	$ ./examples/bin/simple
	Got new count 2
	Got new count 3
	panicked!
	Got new count 4
	Got new count 5
	Got new count 6
	panicked!
	Got new count 7
	Got new count 8
	Got new count 9
	panicked!
	Got new count 10
	Got new count 11
	Got new count 12
	panicked!
	stopped supervisor
*/
package main

import (
	"context"
	"fmt"
	"time"

	supervisor "go.fergus.london/go-supervise/supervisor"
)

func generateSupervisable(shouldPanic bool) supervisor.Supervisable {
	counter := 1

	return func(ctx context.Context) {
		defer func() {
			if recover() != nil {
				fmt.Println("panicked!")
			}
		}()

		for {
			select {
			case <-ctx.Done():
				return

			case <-time.After(time.Millisecond * 250):
				counter++
				fmt.Println("Got new count", counter)

				if counter%3 == 0 && shouldPanic {
					panic("hit 3!")
				}
			}
		}
	}
}

func main() {
	ctx, _ := context.WithCancel(context.Background())
	s, _ := supervisor.NewSupervisorWithOptions(ctx, supervisor.WithWorkers(
		supervisor.SupervisableWorker{
			Func:  generateSupervisable(true),
			Count: 1,
		},
	))

	s.Run()
	<-time.After(time.Millisecond * 3000)

	// Stop the supervisor, and await for 1 second to demonstrate that the
	// routine has actually stopped.
	s.Stop()
	fmt.Println("stopped supervisor")
	<-time.After(time.Millisecond * 1000)
}
