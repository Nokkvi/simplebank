postgres:
	docker run --name postgres -p 3000:5432 -e POSTGRES_USER=admin -e POSTGRES_PASSWORD=secret -d postgres:14-alpine

createdb:
	docker exec -it postgres createdb -U admin --username=admin --owner=admin simple_bank

dropdb:
	docker exec -it postgres dropdb -U admin simple_bank

mock:
	mockgen -build_flags=--mod=mod -package mockdb -destination db/mock/store.go github.com/nokkvi/simplebank/db/sqlc Store

migrateup:
	migrate -path db/migration -database "postgresql://admin:secret@localhost:3000/simple_bank?sslmode=disable" -verbose up

migratedown:
	migrate -path db/migration -database "postgresql://admin:secret@localhost:3000/simple_bank?sslmode=disable" -verbose down

# may need to be run manually
sqlcinit:
	docker run --rm -v %cd%:/src" -w /src kjconroy/sqlc init

# may need to be run manually
sqlcgenerate:
	docker run --rm -v "%cd%:/src" -w /src kjconroy/sqlc generate

test:
	go test -v -cover ./...

server:
	go run main.go

.PHONY: postgres createdb dropdb mock migrateup migratedown sqlcinit sqlcgenerate server test