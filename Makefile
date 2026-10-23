# Makefile for JioSaavn Scrape API

BINARY := bin/jiosaavn-scraper

.PHONY: build test run fmt vet tidy clean

build:
	go build -trimpath -o $(BINARY) .

test:
	go test ./...

run:
	go run .

fmt:
	gofmt -w .

vet:
	go vet ./...

tidy:
	go mod tidy

clean:
	rm -rf bin
