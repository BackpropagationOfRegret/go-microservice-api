.PHONY: up down build logs tidy test demo docs frontend-dev env-init

ENV ?= dev
ENV_FILE := deploy/env/$(ENV).env
COMPOSE := docker compose --env-file $(ENV_FILE)

env-init:
	@test -f $(ENV_FILE) || cp deploy/env/$(ENV).env.example $(ENV_FILE) 2>/dev/null || cp deploy/env/dev.env $(ENV_FILE)
	@echo "Using $(ENV_FILE)"

up: env-init
	$(COMPOSE) up --build -d

down:
	$(COMPOSE) down -v

build: env-init
	$(COMPOSE) build

logs:
	$(COMPOSE) logs -f

tidy:
	@for dir in pkg/events pkg/kafka pkg/auth pkg/config services/user services/restaurant services/order services/payment services/delivery services/notification services/gateway; do \
		(cd $$dir && go mod tidy); \
	done

test:
	@for dir in pkg/events pkg/kafka pkg/auth pkg/config services/user services/restaurant services/order services/payment services/delivery services/notification services/gateway; do \
		echo "==> $$dir" && (cd $$dir && go test ./...); \
	done

docs:
	@echo "Frontend:             http://localhost:$${FRONTEND_PORT:-3000}"
	@echo "API Gateway:          http://localhost:$${GATEWAY_PORT:-8090}"
	@echo "OpenAPI:              http://localhost:$${GATEWAY_PORT:-8090}/openapi.yaml"
	@echo "Scalar:               http://localhost:$${GATEWAY_PORT:-8090}/docs"
	@echo "Traefik dashboard:    http://localhost:$${TRAEFIK_DASHBOARD_PORT:-8088}"
	@echo ""
	@echo "Environments: make up ENV=dev|stage|prod (uses deploy/env/\$$ENV.env)"

frontend-dev:
	cd frontend && npm run dev

demo:
	@echo "=== Register user ==="
	@curl -s -X POST http://localhost:$${GATEWAY_PORT:-8090}/api/auth/register \
		-H 'Content-Type: application/json' \
		-d '{"email":"demo@food.local","password":"secret123","name":"Demo User","phone":"+79001234567"}' | jq .
	@echo "\n=== List restaurants ==="
	@curl -s http://localhost:$${GATEWAY_PORT:-8090}/api/restaurants | jq .
