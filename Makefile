.PHONY: build
build:
	go build -o kubemage

.PHONY: test
test:
	go test ./...

.PHONY: run
run: build
	./kubemage
