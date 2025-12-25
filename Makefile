NETWORK_NAME = simplebank-network
CONTAINER_NAME = simplebank
DB_CONTAINER = simplebank-db
IMAGE_NAME = simplebank:latest

include app.env
export

initdb:
	docker network create $(NETWORK_NAME)
	docker run --name $(DB_CONTAINER) --network $(NETWORK_NAME) -p 5432:5432 -e POSTGRES_USER=root -e POSTGRES_PASSWORD=secret -d postgres:14-alpine
	@echo "Waiting for PostgreSQL to start..."
	@sleep 3
	docker exec -it $(DB_CONTAINER) createdb --username=root --owner=root simple_bank
	migrate -path db/migrations -database $(DB_SOURCE) --verbose up
	@echo "Database initialized"

migrateup:
	migrate -path db/migrations -database $(DB_SOURCE) --verbose up 1

migratedown:
	migrate -path db/migrations -database $(DB_SOURCE) --verbose down 1

generate:
	sqlc generate
	mockgen -destination db/mock/store.go -package mockdb github.com/WilliamOdinson/simplebank/db/sqlc Store
	@echo "Code generated"

test:
	go test -v -cover ./... -count=1

coverage:
	go test -coverprofile=coverage.out ./... -count=1
	go tool cover -html=coverage.out -o coverage.html
	rm -f coverage.out

server:
	docker build -t $(IMAGE_NAME) .
	docker rm -f $(CONTAINER_NAME) || true
	docker run --name $(CONTAINER_NAME) --network $(NETWORK_NAME) -p 8080:8080 -e DB_SOURCE="postgresql://root:secret@$(DB_CONTAINER):5432/simple_bank?sslmode=disable" -d $(IMAGE_NAME)
	@echo "Server running on port 8080"

clean:
	docker rm -f $(CONTAINER_NAME) $(DB_CONTAINER) || true
	docker network rm $(NETWORK_NAME) || true
	rm -f db/mock/store.go
	rm -f db/sqlc/*.sql.go db/sqlc/db.go db/sqlc/models.go db/sqlc/querier.go
	rm -f coverage.html
	@echo "Cleaned up"

.PHONY: initdb migrateup migratedown generate test coverage server clean
