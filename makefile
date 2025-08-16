docker-up:
	docker-compose -f docker-compose.yaml up -d
	sleep 10  # ждём healthcheck

docker-down:
	docker-compose -f docker-compose.yaml down || true

mock:
	mockgen -destination internal/service/mock/repository_mock.go noteApp/internal/service RepositoryI
	mockgen -destination internal/service/mock/hasher_mock.go noteApp/internal/service HasherI
	mockgen -destination internal/handler/mock/service_mock.go noteApp/internal/handler ServiceI

test:
	go test -v -cover ./...

.PHONY: docker-up docker-down test mock