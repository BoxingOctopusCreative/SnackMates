.PHONY: up down api web build test seed seed-remove bot

up:
	docker compose up -d

down:
	docker compose down

api:
	cd api && go run ./cmd/server

web:
	cd web && npm run dev

build:
	cd api && go build -o bin/server ./cmd/server
	cd web && npm run build

test:
	cd api && go test ./...

seed:
	npm run seed

seed-remove:
	npm run seed:remove

bot:
	cd api && go run ./cmd/bot run
