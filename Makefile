DB_URL=postgresql://admin:secret@localhost:3000/simple_bank?sslmode=disable

postgres:
	docker run --name postgres -p 3000:5432 --network simplebank-network -e POSTGRES_USER=admin -e POSTGRES_PASSWORD=secret -d postgres:14-alpine

createdb:
	docker exec -it postgres createdb -U admin --username=admin --owner=admin simple_bank

dropdb:
	docker exec -it postgres dropdb -U admin simple_bank

mock:
	mockgen -build_flags=--mod=mod -package mockdb -destination db/mock/store.go github.com/nokkvi/simplebank/db/sqlc Store

migrateup:
	migrate -path db/migration -database "$(DB_URL)" -verbose up

migrateup1:
	migrate -path db/migration -database "$(DB_URL)" -verbose up 1

migratedown:
	migrate -path db/migration -database "$(DB_URL)" -verbose down

migratedown1:
	migrate -path db/migration -database "$(DB_URL)" -verbose down 1

# may need to be run manually
sqlcinit:
	docker run --rm -v %cd%:/src" -w /src kjconroy/sqlc init

# may need to be run manually
sqlcgenerate:
	docker run --rm -v "%cd%:/src" -w /src kjconroy/sqlc generate

db_docs:
	dbdocs build doc/db.dbml

db_schema:
	dbml2sql --postgres -o doc/schema.sql doc/db.dbml

test:
	go test -v -cover ./...

server:
	go run main.go

.PHONY: postgres createdb dropdb mock migrateup migrateup1 migratedown migratedown1 sqlcinit sqlcgenerate db_docs db_schema server test