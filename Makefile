deps:
	go get -u golang.org/x/lint/golint

lint:
	golint

test:
	go test ./...

.PHONY: examples

examples:
	go build -o ./examples/bin/simple ./examples/simple/main.go
	go build -o ./examples/bin/pipeline ./examples/pipeline/main.go
	go build -o ./examples/bin/actor ./examples/actor/main.go
