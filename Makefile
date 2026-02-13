.PHONY: dev build build-api build-agent test lint docker-up docker-down migrate clean

run:
	@go run ./cmd/api

dev:
	@if command -v air > /dev/null; then \
		air -c .air.toml; \
	else \
		echo "air not installed. Install: https://github.com/air-verse/air?tab=readme-ov-file#installation"; \
	fi

build: build-agent build-api 

build-api:
	@go mod download
	@go build -o bin/signet ./cmd/api

build-agent:
	@cd agent && npm install && npm run build

test: 
	@go test -v -race -coverprofile=coverage.out ./...
	@go tool cover -html=coverage.out -o coverage.html

fmt:
	@go fmt ./...

tidy:
	@go mod tidy
	@go mod vendor

lint:
	@if command -v golangci-lint > /dev/null; then \
		golangci-lint run ./...; \
	else \
		echo "golangci-lint not installed. Install: https://golangci-lint.run/usage/install/"; \
	fi

docker-up:
	@docker-compose up -d

docker-up-infra:
	@docker-compose up -d signet-db signet-cache

docker-down:
	@docker-compose down

docker-build:
	@docker-compose build

migrate:
	@docker-compose exec signet-db psql -U signet -d signet_db -f /docker-entrypoint-initdb.d/001_create_initial_schema.up.sql

migrate-down: 
	@docker-compose exec signet-db psql -U signet -d signet_db -f /docker-entrypoint-initdb.d/001_create_initial_schema.down.sql

logs:
	@docker-compose logs -f signet-api

logs-db:
	@docker-compose logs -f signet-db

logs-redis:
	@docker-compose logs -f signet-cache

clean:
	@rm -rf bin/
	@rm -rf agent/dist/
	@rm -rf agent/node_modules/
	@rm -f coverage.out coverage.html

psql: 
	@docker-compose exec signet-db psql -U signet -d signet_db

redis-cli:
	@docker-compose exec signet-cache redis-cli

install-deps:
	@go mod download
	@cd agent && npm install
