.PHONY: test load monitor swagger run

run:
	go run cmd/server/main.go

test:
	go test -v -race -cover ./...

test-integration:
	go test -tags=integration -v ./tests/...

load:
	k6 run tests/load/k6_script.js --env BASE_URL=http://localhost:8080

monitor:
	docker compose -f monitoring/docker-compose.monitoring.yml up -d

swagger:
	docker compose -f docs/swagger/docker-compose.yml up -d