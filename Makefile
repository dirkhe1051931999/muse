.PHONY: help fmt test check run

help:
	@echo "Targets:"
	@echo "  fmt    gofmt all Go files"
	@echo "  test   run go test ./..."
	@echo "  check  run fmt + test"
	@echo "  run    run git-muse help"

fmt:
	go fmt ./...

test:
	go test ./...

check: fmt test

run:
	go run . help
