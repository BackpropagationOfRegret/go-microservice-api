.PHONY: up down build logs tidy test demo docs frontend-dev

up:
	docker compose up --build -d

down:
	docker compose down -v

build:
	docker compose build

logs:
	docker compose logs -f

tidy:
	@for dir in pkg/events pkg/kafka pkg/auth services/user services/restaurant services/order services/payment services/delivery services/notification services/gateway; do \
		(cd $$dir && go mod tidy); \
	done

test:
	@for dir in pkg/events pkg/kafka pkg/auth services/user services/restaurant services/order services/payment services/delivery services/notification services/gateway; do \
		echo "==> $$dir" && (cd $$dir && go test ./...); \
	done

docs:
	@echo "Frontend:             http://localhost:3000"
	@echo "OpenAPI:              http://localhost:8090/openapi.yaml"
	@echo "Scalar:               http://localhost:8090/docs"
	@echo "Scalar (via Traefik): http://localhost/docs"
	@echo "Traefik dashboard:    http://localhost:8088"

frontend-dev:
	cd frontend && npm run dev

demo:
	@echo "=== Register user ==="
	@curl -s -X POST http://localhost:8090/api/auth/register \
		-H 'Content-Type: application/json' \
		-d '{"email":"demo@food.local","password":"secret123","name":"Demo User","phone":"+79001234567"}' | jq .
	@echo "\n=== List restaurants ==="
	@curl -s http://localhost:8090/api/restaurants | jq .
	@echo "\nSave token and restaurant_id from above, then run order demo manually."
