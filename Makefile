APP_NAME := gomments
DOCKER_TAG := less-coffee/$(APP_NAME):latest
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
		-v $(PWD)/$(DATA_DIR):/home/appuser/data \
		-e BASE_URL=$(BASE_URL) -e PORT=$(PORT) -e ALLOW_ORIGIN=$(ALLOW_ORIGIN) \
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
		-v $(PWD)/$(DATA_DIR):/home/appuser/data \
		-e BASE_URL=$(BASE_URL) -e PORT=$(PORT) -e ALLOW_ORIGIN=$(ALLOW_ORIGIN) GIN_MODE=release \
		$(DOCKER_TAG)

.PHONY: docker
docker: build
	aws ecr get-login-password --region ap-southeast-2 | docker login --username AWS --password-stdin 874239376509.dkr.ecr.ap-southeast-2.amazonaws.com
	docker tag less-coffee/gomments:latest 874239376509.dkr.ecr.ap-southeast-2.amazonaws.com/less-coffee/gomments:latest
	docker push 874239376509.dkr.ecr.ap-southeast-2.amazonaws.com/less-coffee/gomments:latest

