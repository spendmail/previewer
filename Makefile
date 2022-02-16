BIN := "./bin/previewer"
DOCKER_IMG="previewer:develop"
DOCKER_CONTAINER="previewer"
CONFIG_FILE_NAME="previewer.docker"
GIT_HASH := $(shell git log --format="%h" -n 1)
LDFLAGS := -X main.release="develop" -X main.buildDate=$(shell date -u +%Y-%m-%dT%H:%M:%S) -X main.gitHash=$(GIT_HASH)

build:
	go build -v -o $(BIN) -ldflags "$(LDFLAGS)" ./cmd/previewer

launch: build
	$(BIN) -config ./configs/previewer.toml

build-img:
	docker build \
		--build-arg=CONFIG_FILE_NAME="$(CONFIG_FILE_NAME)" \
		--build-arg=LDFLAGS="$(LDFLAGS)" \
		-t $(DOCKER_IMG) \
		-f build/Dockerfile .

run-img: build-img
	docker run -d --rm -p 8888:8888 --name $(DOCKER_CONTAINER) $(DOCKER_IMG)

stop-img:
	docker stop $(DOCKER_CONTAINER)

run:
	LDFLAGS="$(LDFLAGS)" \
	CONFIG_FILE_NAME=$(CONFIG_FILE_NAME) \
	docker-compose -f deployments/docker-compose.yaml up -d

stop:
	LDFLAGS="$(LDFLAGS)" \
	CONFIG_FILE_NAME=$(CONFIG_FILE_NAME) \
	docker-compose -f deployments/docker-compose.yaml down

.PHONY: build run build-img run-img
