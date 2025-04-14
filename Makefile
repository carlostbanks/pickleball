# pickle/Makefile

# Variables
PROTOC := protoc
GO := go
NPM := npm
BACKEND_DIR := ./backend
FRONTEND_DIR := ./frontend
PROTO_DIR := $(BACKEND_DIR)/proto

# Default target
.PHONY: all
all: build

# Setup the project
.PHONY: setup
setup: setup-tools setup-proto setup-backend setup-frontend

# Setup tools
.PHONY: setup-tools
setup-tools:
	$(GO) install google.golang.org/protobuf/cmd/protoc-gen-go@latest
	$(GO) install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
	$(GO) install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest

# Generate protobuf files
.PHONY: setup-proto
setup-proto:
	$(PROTOC) --proto_path=$(PROTO_DIR) \
		--go_out=$(BACKEND_DIR) --go_opt=paths=source_relative \
		--go-grpc_out=$(BACKEND_DIR) --go-grpc_opt=paths=source_relative \
		--grpc-gateway_out=$(BACKEND_DIR) --grpc-gateway_opt=paths=source_relative \
		$(PROTO_DIR)/*.proto

# Setup backend
.PHONY: setup-backend
setup-backend:
	cd $(BACKEND_DIR) && $(GO) mod tidy

# Setup frontend
.PHONY: setup-frontend
setup-frontend:
	cd $(FRONTEND_DIR) && $(NPM) install

# Build the project
.PHONY: build
build: build-backend build-frontend

# Build backend
.PHONY: build-backend
build-backend:
	cd $(BACKEND_DIR) && $(GO) build -o pickle-server .

# Build frontend
.PHONY: build-frontend
build-frontend:
	cd $(FRONTEND_DIR) && $(NPM) run build

# Run the backend
.PHONY: run-backend
run-backend:
	cd $(BACKEND_DIR) && $(GO) run .

# Run the frontend
.PHONY: run-frontend
run-frontend:
	cd $(FRONTEND_DIR) && $(NPM) start

# Run both backend and frontend
.PHONY: run
run:
	make -j2 run-backend run-frontend

# Clean the project
.PHONY: clean
clean: clean-backend clean-frontend

# Clean backend
.PHONY: clean-backend
clean-backend:
	cd $(BACKEND_DIR) && $(GO) clean
	rm -f $(BACKEND_DIR)/pickle-server

# Clean frontend
.PHONY: clean-frontend
clean-frontend:
	cd $(FRONTEND_DIR) && rm -rf build

# Create a new migration
.PHONY: migrate-create
migrate-create:
	cd $(BACKEND_DIR) && $(GO) run -tags 'postgres' \
		github.com/golang-migrate/migrate/v4/cmd/migrate create -ext sql -dir db/migrations -seq $(name)

# Apply migrations
.PHONY: migrate-up
migrate-up:
	cd $(BACKEND_DIR) && $(GO) run -tags 'postgres' \
		github.com/golang-migrate/migrate/v4/cmd/migrate up -path db/migrations -database "postgres://postgres:postgres@localhost:5432/pickle?sslmode=disable"

# Rollback migrations
.PHONY: migrate-down
migrate-down:
	cd $(BACKEND_DIR) && $(GO) run -tags 'postgres' \
		github.com/golang-migrate/migrate/v4/cmd/migrate down -path db/migrations -database "postgres://postgres:postgres@localhost:5432/pickle?sslmode=disable"

# Init PostgreSQL
.PHONY: db-init
db-init:
	psql -U postgres -c "CREATE DATABASE pickle;"
	psql -U postgres -c "CREATE EXTENSION IF NOT EXISTS earthdistance CASCADE;" -d pickle
	psql -U postgres -c "CREATE EXTENSION IF NOT EXISTS cube CASCADE;" -d pickle

# Drop PostgreSQL database
.PHONY: db-drop
db-drop:
	psql -U postgres -c "DROP DATABASE IF EXISTS pickle;"

# Generate mock data
.PHONY: db-mock
db-mock:
	cd $(BACKEND_DIR) && $(GO) run scripts/mock_data.go

# Test backend
.PHONY: test-backend
test-backend:
	cd $(BACKEND_DIR) && $(GO) test ./... -v

# Test frontend
.PHONY: test-frontend
test-frontend:
	cd $(FRONTEND_DIR) && $(NPM) test

# Run linter for backend
.PHONY: lint-backend
lint-backend:
	cd $(BACKEND_DIR) && golangci-lint run

# Run linter for frontend
.PHONY: lint-frontend
lint-frontend:
	cd $(FRONTEND_DIR) && $(NPM) run lint

# Help
.PHONY: help
help:
	@echo "Available targets:"
	@echo "  all             - Build the project"
	@echo "  setup           - Setup the project"
	@echo "  setup-tools     - Setup development tools"
	@echo "  setup-proto     - Generate protobuf files"
	@echo "  setup-backend   - Setup backend dependencies"
	@echo "  setup-frontend  - Setup frontend dependencies"
	@echo "  build           - Build both backend and frontend"
	@echo "  build-backend   - Build the backend"
	@echo "  build-frontend  - Build the frontend"
	@echo "  run             - Run both backend and frontend"
	@echo "  run-backend     - Run the backend"
	@echo "  run-frontend    - Run the frontend"
	@echo "  clean           - Clean build artifacts"
	@echo "  migrate-create  - Create a new migration"
	@echo "  migrate-up      - Apply migrations"
	@echo "  migrate-down    - Rollback migrations"
	@echo "  db-init         - Initialize PostgreSQL database"
	@echo "  db-drop         - Drop PostgreSQL database"
	@echo "  db-mock         - Generate mock data"
	@echo "  test-backend    - Run backend tests"
	@echo "  test-frontend   - Run frontend tests"
	@echo "  lint-backend    - Run linter for backend"
	@echo "  lint-frontend   - Run linter for frontend"
	@echo "  help            - Show this help message"