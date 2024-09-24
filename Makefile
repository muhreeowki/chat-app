build-backend:
	@go build -o bin/mchat

backend: build-backend
	@./bin/mchat

build-client: client/
	@cd client && npm install && npm run build

client: build-client
	@cd client && npm run start

test:
	@go test ./...
