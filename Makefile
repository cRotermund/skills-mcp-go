BINARY := skills-server

.PHONY: build test race vet install dev

build:
	go build -o $(BINARY) ./cmd/server

test:
	go test ./...

race:
	go test -race ./...

vet:
	go vet ./...

install: build
	mkdir -p $(HOME)/.local/bin
	cp $(BINARY) $(HOME)/.local/bin/$(BINARY)

dev:
	go run ./cmd/server
