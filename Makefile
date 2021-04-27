deps:
	go get -u golang.org/x/lint/golint

lint:
	golint

test:
	go test