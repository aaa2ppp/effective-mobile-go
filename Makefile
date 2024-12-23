.PHONY: build run docs tidy

all: tidy docs build

tidy:
	go mod tidy

docs:
	swag init -d cmd/app,internal/handler && \
	swag fmt  -d cmd/app,internal/handler

build:
	go build -o bin/app.bin cmd/app/main.go

run:
	bin/app.bin
