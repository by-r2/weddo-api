.PHONY: build run test clean setup migrate-up migrate-down

BINARY=bin/api
MAIN=cmd/api/main.go

build:
	CGO_ENABLED=1 go build -o $(BINARY) $(MAIN)

run:
	CGO_ENABLED=1 go run $(MAIN)

test:
	CGO_ENABLED=1 go test ./... -v

clean:
	rm -rf bin/ data/

setup:
	go mod tidy
	@test -f .env || cp .env.example .env
	@echo "Setup concluído. Edite o .env se necessário."

migrate-up:
	go run $(MAIN) -migrate-up

migrate-down:
	go run $(MAIN) -migrate-down
