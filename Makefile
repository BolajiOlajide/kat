.PHONY: install, run

ARGS :=
VERSION ?= dev

install:
	go install ./...

run:
	go run ./cmd/kat $(ARGS)

start_doc:
	bundle exec jekyll serve

build:
	go build -ldflags "-X github.com/BolajiOlajide/kat/internal/version.version=$(VERSION)" ./cmd/kat
