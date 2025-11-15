.PHONY: build up down image run fmt test
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
fmt:
	gofmt -w .
run: build
	./bin/$(APP_NAME)
test:
	go test ./...