build-backend: backend/
	@cd backend && go build -o bin/mchat

backend: build-backend
	@cd backend && ./bin/mchat

build-client: client/
	@cd client && npm install && npm run build

client: build-client
	@cd client && npm run start

docker-build-backend:
	@cd backend && docker build -t chatserver .

docker-build-client:
	@cd client && docker build -t chatclient .

docker-up: docker-build-backend docker-build-client
	docker compose up

test:
	@go test ./...
