.PHONY: clean cover docs generate lint test

clean:
	rm -f coverage.out

cover:
	go test -coverprofile=coverage.out
	go tool cover -html=coverage.out

docs:
	pkgsite -open .

generate:
	go generate ./...

lint:
	golangci-lint run ./...

test:
	go build -o patrolbot
	go test ./... -v

testupdate:
	go build -o patrolbot
	go test main_test.go -v -update