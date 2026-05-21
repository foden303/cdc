APP_NAME := cdc
BIN_DIR := bin
CONFIG_FILE := deploy/app/config.yaml
PROTO_DIR := proto
PROTO_IMAGE_NAME := cdc-proto-gen
PROTO_DOCKERFILE := $(PROTO_DIR)/Dockerfile

.PHONY: all build run test tidy up down fix-perms clean gen-proto .docker-check .proto-image fe-install fe-dev fe-build fe-lint

all: tidy build

build:
	@mkdir -p $(BIN_DIR)
	go build -o $(BIN_DIR)/$(APP_NAME) ./cmd/cdc

run: build
	./$(BIN_DIR)/$(APP_NAME) -config $(CONFIG_FILE)

test:
	go test -v ./...

tidy:
	go mod tidy

up:
	docker compose up -d

down:
	docker compose down

fix-perms:
	@echo "Fixing nats-data permissions..."
	sudo chown -R $(shell id -u):$(shell id -g) nats-data

clean:
	rm -rf $(BIN_DIR)

.docker-check:
	@docker info > /dev/null 2>&1 || (echo "Error: Docker daemon is not running. Please start Docker and try again." >&2; exit 1)

.proto-image:
	@docker image inspect $(PROTO_IMAGE_NAME) > /dev/null 2>&1 || \
		(echo "Building proto generation image..."; docker build -f $(PROTO_DOCKERFILE) -t $(PROTO_IMAGE_NAME) $(PROTO_DIR))

gen-proto: .docker-check .proto-image
	@docker run --rm \
		-v $(PWD):/workspace \
		--user $(shell id -u):$(shell id -g) \
		-e BUF_CACHE_DIR=/tmp/buf-cache \
		$(PROTO_IMAGE_NAME) sh /workspace/$(PROTO_DIR)/generate.sh


# ─── Frontend ────────────────────────────────────────────────────────

fe-install:
	cd website && npm install

fe-dev:
	cd website && npm run dev

fe-build:
	cd website && npm run build

fe-lint:
	cd website && npm run lint
