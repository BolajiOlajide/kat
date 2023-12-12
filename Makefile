.PHONY: install, run

ARGS :=

install:
	go install ./...

run:
	go run ./cmd/kat $(ARGS)

start_doc:
	bundle exec jekyll serve
