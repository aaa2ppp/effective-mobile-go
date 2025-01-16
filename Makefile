.PHONY: build run docs tidy

all: build

tidy:
	go mod tidy

docs/docs.go: cmd/app/*.go internal/handler/*.go
	swag fmt  -d cmd/app,internal/handler && \
	swag init -d cmd/app,internal/handler

docs: docs/docs.go

build: docs
	go build -o bin/app.bin cmd/app/main.go

run:
	bin/app.bin
