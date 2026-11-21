SHELL := /bin/bash

DATABASE_URL ?= postgres://postgres:postgres@localhost:5432/study_companion?sslmode=disable

-include .env
export

.PHONY: run build vet test seed-superadmin

run:
	cd backend && go run ./cmd/server

build:
	cd backend && go build ./...

vet:
	cd backend && go vet ./...

test:
	cd backend && go test ./...

seed-superadmin:
	cd backend && go run ./cmd/server seed-superadmin
