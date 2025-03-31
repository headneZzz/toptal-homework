.PHONY: dc run test lint

dc:
	docker-compose up --remove-orphans --build

run:
	go build -o app cmd/main.go ./app

test:
	go test -race ./...

lint:
	golangci-lint run

swag:
	swag init -g cmd/server/main.go