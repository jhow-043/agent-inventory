.PHONY: help build-server build-agent run test lint docker-up docker-down docker-logs create-user

help: ## Show available targets
	@grep -E '^[a-zA-Z_-]+:.*?## .*$$' $(MAKEFILE_LIST) | sort | awk 'BEGIN {FS = ":.*?## "}; {printf "\033[36m%-20s\033[0m %s\n", $$1, $$2}'

build-server: ## Build the API server binary
	cd server && go build -o ../bin/server ./cmd/api

build-agent: ## Build the Windows agent binary
	cd agent && GOOS=windows GOARCH=amd64 go build -o ../bin/agent.exe ./cmd/agent

run: ## Run the API server locally
	cd server && go run ./cmd/api

test: ## Run all tests with race detection and coverage
	go test ./... -race -cover -count=1

lint: ## Run linters
	golangci-lint run ./...

docker-up: ## Start all Docker services
	docker compose up -d --build

docker-down: ## Stop all Docker services
	docker compose down

docker-logs: ## Tail Docker logs
	docker compose logs -f

create-user: ## Create a dashboard user (usage: make create-user USERNAME=admin PASSWORD=secret)
	docker compose exec api ./server create-user --username=$(USERNAME) --password=$(PASSWORD)

tidy: ## Run go mod tidy on all modules
	cd shared && go mod tidy
	cd server && go mod tidy
	cd agent && go mod tidy
