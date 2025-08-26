docker-up:
	docker compose -f docker-compose.yaml up -d

docker-down:
	docker compose -f docker-compose.yaml down || true

docker-logs:
	docker compose -f docker-compose.yaml logs -f

docker-build:
	docker compose -f docker-compose.yaml build

docker-restart:
	docker compose -f docker-compose.yaml down
	docker compose -f docker-compose.yaml up -d --build

mock:
	mockgen -destination internal/service/mock/repository_mock.go noteApp/internal/service RepositoryI
	mockgen -destination internal/service/mock/hasher_mock.go noteApp/internal/service HasherI
	mockgen -destination internal/handler/mock/service_mock.go noteApp/internal/handler ServiceI

test-start: docker-up test docker-down

test:
	go test -v -cover ./...

run:
	go run ./cmd


.PHONY: docker-up docker-down docker-logs docker-build docker-restart test test-start mock run