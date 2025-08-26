docker-up:
	docker compose -f docker-compose.yaml up -d

docker-down:
	docker compose -f docker-compose.yaml down || true

mock:
	mockgen -destination internal/service/mock/repository_mock.go noteApp/internal/service RepositoryI
	mockgen -destination internal/service/mock/hasher_mock.go noteApp/internal/service HasherI
	mockgen -destination internal/handler/mock/service_mock.go noteApp/internal/handler ServiceI

test-start: docker-up test docker-down

test:
	go test -v -cover ./...

start: docker-up run

run:
	go run ./cmd


.PHONY: docker-up docker-down test test-start mock start run