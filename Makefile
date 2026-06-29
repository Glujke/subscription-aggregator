.PHONY: up down logs build test-health swagger

swagger:
	swag init -g cmd/api/main.go -o docs --parseDependency --parseInternal

up:
	docker compose up --build -d

down:
	docker compose down

logs:
	docker compose logs -f api

build:
	docker compose build

test-health:
	curl -s http://localhost:8080/health
