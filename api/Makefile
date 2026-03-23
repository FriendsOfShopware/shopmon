.PHONY: generate generate-sql generate-api build run run-worker tidy test

generate: generate-sql generate-api

generate-sql:
	sqlc generate

generate-api:
	oapi-codegen --config oapi-codegen-api.yaml openapi/spec.yaml
	oapi-codegen --config oapi-codegen-auth.yaml openapi/spec.yaml

build:
	go build -o bin/shopmon .

run:
	go run . server

run-worker:
	go run . worker

tidy:
	go mod tidy

test:
	go test ./internal/... -count=1 -timeout 120s

.DEFAULT_GOAL := build
