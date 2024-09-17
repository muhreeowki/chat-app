build:
	@go build -o bin/mchat

run: build
	@./bin/mchat

test:
	@go test ./...
