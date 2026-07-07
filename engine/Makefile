.PHONY: all build test lint clean deps

all: clean deps build test lint

deps:
	go mod tidy
	go mod download

build:
	mkdir -p bin
	go build -o bin/substrate ./cmd/substrate

test:
	go test -v ./...

lint:
	go vet ./...

clean:
	rm -rf bin/
