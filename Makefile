deps:
	go get -u golang.org/x/lint/golint

docs:
	@echo "pass"

lint:
	~/go/bin/golint # no, I haven't bothered with my GOPATH/PATH yet.

test:
	go test