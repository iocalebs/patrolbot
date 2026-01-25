export GOCOVERDIR := coverage

.PHONY: build
.PHONY: clean
.PHONY: cover
.PHONY: docs 
.PHONY: e2e
.PHONY: generate 
.PHONY: lint 
.PHONY: short
.PHONY: test 
.PHONY: update
.PHONY: updateall

build:
	go build -o patrolbot

clean:
	rm -f patrolbot
	rm -rf $(GOCOVERDIR)

cover:
	rm -rf $(GOCOVERDIR)
	mkdir -p $(GOCOVERDIR)/integration
	mkdir -p $(GOCOVERDIR)/unit
	go build -cover -o patrolbot
	-GOCOVERDIR=$(CURDIR)/$(GOCOVERDIR)/integration go test -count=1 -parallel=1 main_test.go
	-go test -cover -parallel=1 ./internal/... -args -test.gocoverdir=$(CURDIR)/$(GOCOVERDIR)/unit
	go tool covdata textfmt -i=./$(GOCOVERDIR)/integration -o=./$(GOCOVERDIR)/profile-integation.txt
	go tool covdata textfmt -i=./$(GOCOVERDIR)/unit -o=./$(GOCOVERDIR)/profile-unit.txt
	go tool covdata textfmt -i=./$(GOCOVERDIR)/integration,./$(GOCOVERDIR)/unit -o=./$(GOCOVERDIR)/profile-merged.txt
# It seems the 2nd-generation coverage output omits files with 0% coverage.
# Workaround: Merge profiles with a "zero" profile that lists all files at 0% coverage, made by running the old coverage 
# generator on all packages but with a -run regex that matches no tests.
	go test -coverprofile=$(GOCOVERDIR)/profile-zero.txt -parallel=1 -run "a^" ./...
# go internal tools and test utils are removed from the report.
	sed '/^github\.com\/iocalebs\/patrolbot\/internal\/tools/d' $(GOCOVERDIR)/profile-zero.txt > $(GOCOVERDIR)/profile-zero.tmp
	sed '/^github\.com\/iocalebs\/patrolbot\/internal\/httpstub/d' $(GOCOVERDIR)/profile-zero.txt > $(GOCOVERDIR)/profile-zero.tmp
	mv $(GOCOVERDIR)/profile-zero.tmp $(GOCOVERDIR)/profile-zero.txt
	go tool gocovmerge $(GOCOVERDIR)/profile-merged.txt $(GOCOVERDIR)/profile-zero.txt > $(GOCOVERDIR)/tmp.txt
	mv $(GOCOVERDIR)/tmp.txt $(GOCOVERDIR)/profile-merged.txt
	go tool cover -html=$(GOCOVERDIR)/profile-merged.txt

docs:
	pkgsite -open .

e2e: build
	./patrolbot report weekly --wiki testwiki --yes

generate:
	go generate ./...

lint:
	golangci-lint run --fix ./...

short: build
	go test ./... -short

test: build
	go test ./...

update: build
	go test main_test.go -update -short

updateall: build
	go test main_test.go -update
