TARGET=webserver
SECRETS=secrets.env

.PHONY: all
all: build check

.PHONY: build
build: bin/$(TARGET)

.PHONY: check
check:
	@echo -e "\n# Running unit tests..."
	@go test -cover ./pkg/$(TARGET)

.PHONY: run
run: all
	@echo -e "\n# Running $(TARGET)..."
	@env $$(cat $(SECRETS) | xargs) ./bin/$(TARGET)

.PHONY: clean
clean:
	@echo -e "\n# Removing build..."
	rm -rf bin

.PHONY: docker
docker:
	@echo -e "\n# Building docker image..."
	docker build . -t torvalds_number:multistage
	@echo -e "\n# Deploying docker image..."
	docker run --env-file $(SECRETS) --rm -p 80:80 torvalds_number:multistage

bin/%: cmd/% pkg/% Makefile
	@echo -e "\n# Building $@..."
	go build -o $@ ./$<
