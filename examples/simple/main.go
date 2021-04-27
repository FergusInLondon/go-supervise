// This example demonstrates a very simple Supervisable; there is no tree
// - it's just one goroutine that should be resilient to panicking. Every
// 750ms the goroutine panics, but it's restarted repeatedly.
package main

import (
	"context"
	"fmt"
	"time"

	supervisor "go.fergus.london/go-supervise"
)

func generateSupervisable(shouldPanic bool) supervisor.Supervisable {
	counter := 1

	return func(ctx context.Context, completed chan struct{}) {
		defer func() {
			if recover() != nil {
				fmt.Println("panicked!")
			}

			close(completed)
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
	s := supervisor.NewSimpleSupervisor(context.Background(), generateSupervisable(true))

	go s.Run()
	<-time.After(time.Millisecond * 3000)

	// Stop the supervisor, and await for 1 second to demonstrate that the
	// routine has actually stopped.
	s.Stop()
	fmt.Println("stopped supervisor")
	<-time.After(time.Millisecond * 1000)
}
