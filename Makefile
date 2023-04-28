.PHONY: install, run

ARGS :=

install:
	go install ./...

run:
	go run ./cmd/kat $(ARGS)
