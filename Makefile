deps:
	go get -u golang.org/x/lint/golint

docs:
	@echo "pass"

lint:
	golint

test:
	go test