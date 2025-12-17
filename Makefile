.PHONY: generate lint test

generate:
	go generate ./...

lint:
	golangci-lint run ./...

test:
	go test -v ./...