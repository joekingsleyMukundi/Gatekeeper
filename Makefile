getpostgres:
	docker pull postgres:17-alpine

postgres:
	docker run --name GKpostgres -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:17-alpine

startpostgres:
	docker start GKpostgres

stoppostgres:
	docker stop GKpostgres

createdb:
	docker exec -it GKpostgres createdb --username=root --owner=root gate_keeper

dropdb:
	docker exec -it GKpostgres dropdb gate_keeper

migrate:
	@read -p "Enter migration name: " name; \
	migrate create -ext sql -dir db/migrations -seq $$name

migrateup:
	migrate -path db/migrations -database "postgresql://root:secret@localhost:5432/gate_keeper?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migrations -database "postgresql://root:secret@localhost:5432/gate_keeper?sslmode=disable" -verbose down

sqlc:
	sqlc generate

mock:
	mockgen -package mockdb -destination db/mock/store.go github.com/joekingsleyMukundi/Gatekeeper/db/sqlc Store

redis:
	docker run --name redis -p 6379:6379 -d redis:7-alpine

startredis:
	docker start redis

start: startredis startpostgres
	@echo "All containers started"

status:
	@echo "Docker containers status:"
	@docker ps -a | grep -E 'redis|GKpostgres'

test:
	go test -v -cover ./...

run:
	@echo "Checking if containers are running..."
	@if ! docker ps | grep -q "redis"; then \
		echo "Starting Redis..."; \
		docker start redis 2>/dev/null || docker run --name redis -p 6379:6379 -d redis:7-alpine; \
	fi
	@if ! docker ps | grep -q "GKpostgres"; then \
		echo "Starting PostgreSQL..."; \
		docker start GKpostgres 2>/dev/null || make postgres; \
	fi
	@echo "Running Go application..."
	go run main.go

# Default target to display help
help:
	@echo "╔════════════════════════════════════════════════════════════════╗"
	@echo "║                   Gatekeeper Project Makefile                  ║"
	@echo "╚════════════════════════════════════════════════════════════════╝"
	@echo ""
	@echo "🐳 Docker Container Commands:"
	@echo "  make getpostgres      - Pull PostgreSQL 17 Alpine image"
	@echo "  make postgres         - Create and start a new PostgreSQL container"
	@echo "  make startpostgres    - Start existing PostgreSQL container"
	@echo "  make stoppostgres     - Stop PostgreSQL container"
	@echo "  make redis            - Create and start a new Redis container"
	@echo "  make startredis       - Start existing Redis container"
	@echo "  make start            - Start all containers (Redis and PostgreSQL)"
	@echo "  make status           - Check status of all containers"
	@echo ""
	@echo "💾 Database Management:"
	@echo "  make createdb         - Create the gate_keeper database"
	@echo "  make dropdb           - Drop the gate_keeper database"
	@echo "  make migrate          - Create a new migration file (will prompt for name)"
	@echo "  make migrateup        - Run all database migrations"
	@echo "  make migratedown      - Revert all database migrations"
	@echo ""
	@echo "🔧 Development Commands:"
	@echo "  make sqlc             - Generate Go code from SQL"
	@echo "  make mock             - Generate mock Store for testing"
	@echo "  make test             - Run all tests with coverage"
	@echo "  make run              - Run the Go application (auto-starts containers)"
	@echo ""
	@echo "For more details, see the project documentation."

# Default target
.DEFAULT_GOAL := help

.PHONY: getpostgres postgres startpostgres stoppostgres createdb dropdb migrate migrateup migratedown sqlc test run mock redis startredis start status help