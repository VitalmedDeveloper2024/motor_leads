.PHONY: run build tidy up down migrate

run:
	go run ./cmd/server

build:
	CGO_ENABLED=0 go build -o bin/motor-leads ./cmd/server

tidy:
	go mod tidy

up:
	docker compose up --build

down:
	docker compose down

migrate:
	@for f in db/migrations/*.sql; do echo ">> $$f"; psql "$$DATABASE_URL" -f $$f; done
