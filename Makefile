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
test:
	go test -v -cover ./...
run:
	go run main.go
.PHONY: getpostgres postgres startpostgres stoppostgres createdb dropdb sqlc test run mock
