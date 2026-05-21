.PHONY: up down build logs

up: .env
	docker compose up --build

down:
	docker compose down

build: .env
	docker compose build

logs:
	docker compose logs -f

# Generate .env from template and fill JWT secrets with random values.
# Make treats .env as a file target — this recipe runs only if .env is missing.
.env:
	cp .env.example .env
	@sed -i "s|^JWT_ACCESS_SECRET=.*|JWT_ACCESS_SECRET=$$(openssl rand -hex 64)|" .env
	@sed -i "s|^JWT_REFRESH_SECRET=.*|JWT_REFRESH_SECRET=$$(openssl rand -hex 64)|" .env
	@echo ".env created with random JWT secrets"
