APP_NAME := less-coffee/gomments
DOCKER_TAG := $(APP_NAME):latest
DATA_DIR := data

.PHONY: build
build:
	docker build -t $(DOCKER_TAG) .

.PHONY: run
run: build
	@mkdir -p $(DATA_DIR)
	-docker stop $(APP_NAME)
	-docker rm $(APP_NAME)
	@echo "Running at http://localhost:$(PORT)"
	docker run \
		--name $(APP_NAME) \
		-p $(PORT):$(PORT) \
		-v $(PWD)/$(DATA_DIR):/root/data \
		-e BASE_URL=$(BASE_URL) -e PORT=$(PORT) \
		$(DOCKER_TAG)

.PHONY: run-release
run-release: build
	@mkdir -p $(DATA_DIR)
	-docker stop $(APP_NAME)
	-docker rm $(APP_NAME)
	@echo "Running at http://localhost:$(PORT)"
	docker run \
		--name $(APP_NAME) \
		-p $(PORT):$(PORT) \
		-v $(PWD)/$(DATA_DIR):/root/data \
		-e BASE_URL=$(BASE_URL) -e PORT=$(PORT) -e GIN_MODE=release \
		$(DOCKER_TAG)
