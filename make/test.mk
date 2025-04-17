.PHONY: test
test: build
	@echo "running unit tests..."
	go test -v -failfast ./...