APP_NAME=room-booking-service

up:
	docker compose up --build -d

down:
	docker compose down -v

logs:
	docker compose logs -f app

seed:
	docker compose exec app /seed

test:
	go test ./... -coverprofile=coverage.out

test-unit:
	go test ./internal/... ./tests/... -cover

test-db:
	RUN_DB_TESTS=1 go test ./tests -run TestPostgresRepositories -v

fmt:
	gofmt -w ./cmd ./internal ./tests

lint:
	golangci-lint run ./...

swagger:
	@echo "Install swag first: go install github.com/swaggo/swag/cmd/swag@latest"
	@echo "Then run: swag init -g cmd/server/main.go -o docs/swagger"
