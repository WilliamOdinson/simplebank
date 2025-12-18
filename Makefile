postgres:
	docker run --name simplebank -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:14-alpine

createdb:
	docker exec -it simplebank createdb --username=root --owner=root simple_bank

dropdb:
	docker exec -it simplebank dropdb simple_bank

migrateup:
	migrate -path db/migrations -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" --verbose up
	@echo "Migration completed"

migratedown:
	migrate -path db/migrations -database "postgresql://root:secret@localhost:5432/simple_bank?sslmode=disable" --verbose down
	@echo "Migration rolled back"

sqlc:
	sqlc generate

test:
	go test -v -cover ./...

coverage:
	go test -v -coverprofile=coverage.out ./...
	go tool cover -html=coverage.out -o coverage.html
	rm -f coverage.out

server:
	go run main.go

mock:
	mockgen -destination db/mock/store.go -package mockdb github.com/WilliamOdinson/simplebank/db/sqlc Store

.PHONY: createdb dropdb postgres migrateup migratedown sqlc test coverage server mock
