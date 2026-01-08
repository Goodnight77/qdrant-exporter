BINARY_NAME=qdrant-exporter
DOCKER_IMAGE=qdrant-exporter

build:
	go build -o $(BINARY_NAME) .

test:
	go test -v ./...

lint:
	golangci-lint run

docker-build:
	docker build -t $(DOCKER_IMAGE) .

run:
	go run main.go

clean:
	rm -f $(BINARY_NAME)

.PHONY: build test lint docker-build run clean
