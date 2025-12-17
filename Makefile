.PHONY: docs generate lint test

docs:
	pkgsite -open .

generate:
	go generate ./...

lint:
	golangci-lint run ./...

test:
	go test -v ./...