DB_MIGRATE_URL = postgres://login:pass@localhost:5432/db-name?sslmode=disable
MIGRATE_PATH = ./migration/postgres/apple

.PHONY: up down run mod mod-update lint vuln test test-race test-integration test-coverage \
	build docker-build migrate-install migrate-create migrate-up migrate-down

up:
	docker compose up --build -d --force-recreate
	docker compose logs -f

down:
	docker compose down

run: mod
	go run ./cmd/app

build:
	CGO_ENABLED=0 go build -trimpath -ldflags="-s -w" -o ./my-app ./cmd/app

docker-build:
	docker build -t my-app .

mod:
	go mod tidy

mod-update:
	go get -u ./...
	go mod tidy

lint:
	golangci-lint run

vuln:
	go run golang.org/x/vuln/cmd/govulncheck@latest ./...

test:
	go test -race -cover ./...

# Needs Docker: starts Postgres and Redis via testcontainers.
test-integration:
	go test -tags integration -race -count=1 ./...

test-coverage:
	go test -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out

migrate-install:
	go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@v4.18.1

# usage: make migrate-create name=add_something
migrate-create:
	migrate create -ext sql -dir "$(MIGRATE_PATH)" $(name)

migrate-up:
	migrate -database "$(DB_MIGRATE_URL)" -path "$(MIGRATE_PATH)" up

migrate-down:
	migrate -database "$(DB_MIGRATE_URL)" -path "$(MIGRATE_PATH)" down -all
