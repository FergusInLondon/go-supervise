/*
This example demonstrates a Supervisor that's configured to monitor a
pipeline composed of multiple go-routines.

The pipeline is composed of 5 go-routines that are chained via channels.
Two of those go-routines are configured to panic upon every third input.
When executed you can see that despite these panics the functionality
looks interrupted.

	$ ./examples/bin/pipeline
	[Example] Dispatching counter 0
	Got new input 0 with ID 0
	Got new input 0 with ID 1
	Got new input 0 with ID 2
	Got new input 0 with ID 3
	panicked! 3
	Got new input 0 with ID 4
	panicked! 4
	[Example] Receiving counter 0
	[Example] Dispatching counter 1
	Got new input 1 with ID 0
	Got new input 1 with ID 1
	Got new input 1 with ID 2
	Got new input 1 with ID 3
	Got new input 1 with ID 4
	[Example] Receiving counter 1
		[...]
	[Example] Dispatching counter 13
	Got new input 13 with ID 0
	Got new input 13 with ID 1
	Got new input 13 with ID 2
	Got new input 13 with ID 3
	Got new input 13 with ID 4
	[Example] Receiving counter 13
	stopped supervisor
*/
package main

import (
	"context"
	"fmt"
	"time"

	"go.fergus.london/go-supervise/supervisor"
)

func generateSupervisable(shouldPanic bool, id int, rx, tx chan int) supervisor.Supervisable {
	return func(ctx context.Context) {
		defer func() {
			if recover() != nil {
				fmt.Println("panicked!", id)
			}
		}()

		for {
			select {
			case <-ctx.Done():
				return

			case v := <-rx:
				fmt.Println("Got new input", v, "with ID", id)
				tx <- v

				if v%3 == 0 && shouldPanic {
					panic("panic!")
				}
			}
		}
	}
}

func main() {
	ioChans := make([]chan int, 6)
	for i := 0; i < 6; i++ {
		ioChans[i] = make(chan int)
	}

	supervisorWorkers := make([]supervisor.SupervisableWorker, 5)
	for i := 0; i < 5; i++ {
		supervisorWorkers[i] = supervisor.SupervisableWorker{
			Func:  generateSupervisable((i >= 3), i, ioChans[i], ioChans[i+1]),
			Count: 1,
		}
	}

	ctx, _ := context.WithCancel(context.Background())
	s, _ := supervisor.NewSupervisorWithOptions(
		ctx, supervisor.WithWorkers(supervisorWorkers...),
	)
	s.Run()

	go func() {
		counter := 0
		for {
			select {
			case <-time.After(time.Millisecond * 100):
				fmt.Println("[Example] Dispatching counter", counter)
				ioChans[0] <- counter
				counter++
			case v := <-ioChans[5]:
				fmt.Println("[Example] Receiving counter", v)
			}
		}
	}()

	<-time.After(time.Millisecond * 1500)
	s.Stop()
	s.Wait()

	fmt.Println("stopped supervisor")
}
