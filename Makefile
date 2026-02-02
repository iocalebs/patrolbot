export GOCOVERDIR := coverage

.PHONY: apply
.PHONY: build
.PHONY: clean
.PHONY: cover
.PHONY: decrypt
.PHONY: docker
.PHONY: docs 
.PHONY: e2e
.PHONY: edit
.PHONY: encrypt
.PHONY: generate 
.PHONY: lint
.PHONY: local
.PHONY: login
.PHONY: ngrok
.PHONY: plan
.PHONY: register
.PHONY: serve
.PHONY: short
.PHONY: test 
.PHONY: update
.PHONY: updateall

apply:
	cd terraform && terraform apply

build:
	go build -trimpath -o patrolbot

clean:
	rm -f patrolbot
	rm -rf $(GOCOVERDIR)

cover:
	rm -rf $(GOCOVERDIR)
	mkdir -p $(GOCOVERDIR)/integration
	mkdir -p $(GOCOVERDIR)/unit
	go build -trimpath -cover -o patrolbot
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
	sed -e '/^github\.com\/iocalebs\/patrolbot\/internal\/tools/d' \
	    -e '/^github\.com\/iocalebs\/patrolbot\/internal\/testscriptutil/d' \
	    $(GOCOVERDIR)/profile-zero.txt > $(GOCOVERDIR)/profile-zero.tmp
	mv $(GOCOVERDIR)/profile-zero.tmp $(GOCOVERDIR)/profile-zero.txt
	go tool gocovmerge $(GOCOVERDIR)/profile-merged.txt $(GOCOVERDIR)/profile-zero.txt > $(GOCOVERDIR)/tmp.txt
	mv $(GOCOVERDIR)/tmp.txt $(GOCOVERDIR)/profile-merged.txt
	go tool cover -html=$(GOCOVERDIR)/profile-merged.txt

decrypt:
	sops decrypt --input-type dotenv --output-type dotenv .env.enc > .env
	direnv allow

docker:
	docker build -t patrolbot:local .
	docker run --env-file .env --rm -p 8080:8080 patrolbot:local

docs:
	pkgsite -open .

e2e: build
	./patrolbot report expiring --wiki testwiki --yes

edit:
	EDITOR="code --wait" sops edit --input-type dotenv --output-type dotenv .env.enc

encrypt:
	sops encrypt .env > .env.enc

generate:
	go generate ./...

lint:
	golangci-lint run --fix ./...

local:
	cp .env.local .env
	direnv allow

login:
	gcloud auth application-default login

ngrok:
	ngrok http 8080

plan:
	cd terraform && terraform plan

register: build
	./patrolbot register

serve:
	air serve

short: build
	go test ./... -short

test: build
	go test ./...

update: build
	go test main_test.go -update -short

updateall: build
	go test main_test.go -update
