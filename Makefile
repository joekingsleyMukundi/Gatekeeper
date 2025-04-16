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
	migrate create -ext sql -dir db/migrations -seq init_schema
migrateup:
	migrate -path db/migrations -database "postgresql://root:secret@localhost:5432/gate_keeper?sslmode=disable" -verbose up
migratedown:
	migrate -path db/migrations -database "postgresql://root:secret@localhost:5432/gate_keeper?sslmode=disable" -verbose down
sqlc:
	sqlc generate
.PHONY: getpostgres postgres startpostgres stoppostgres createdb dropdb sqlc