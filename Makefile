.PHONY: build up down image run fmt test lint lint-fix
APP_NAME=server
DOCKER_IMAGE=reviewer-service:local
build:
	go build -o bin/$(APP_NAME) ./cmd/server
image: build
	docker build -t $(DOCKER_IMAGE) .
up: image
	docker-compose up --build
down:
	docker-compose down -v
run: build
	./bin/$(APP_NAME)
fmt:
	gofmt -w .
logs:
	docker-compose logs -f
lint:
	golangci-lint run
lint-fix:
	golangci-lint run --fix