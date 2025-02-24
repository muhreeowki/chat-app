build:
	@go build -o bin/mchat

clean-db:
	@docker stop postgres && docker rm postgres && docker run --name postgres -e POSTGRES_PASSWORD=mchat -p 5432:5432 -d postgres

run: build
	@./bin/mchat

docker-build-backend:
	@cd backend && docker build -t chatserver:1 .

docker-build-client:
	@cd client && docker build -t chatclient:1 .

docker-up: docker-build-backend docker-build-client
	@docker compose up -d

docker-down:
	@docker compose down

docker-rmi:
	@docker rmi chatclient:1 chatserver:1
	@echo "✅Successfully removed chat-app docker images."

test:
	@go test ./...
