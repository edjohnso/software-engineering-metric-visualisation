TARGET=webserver

export GHO_CLIENT_ID
export GHO_CLIENT_SECRET
export GHO_PAT

.PHONY: all
all: build check

.PHONY: build
build: bin/$(TARGET)

.PHONY: check
check:
	@echo -e "\n# Running go vet..."
	go vet ./cmd/$(TARGET)
	@echo -e "\n# Running unit tests..."
	go test -cover ./pkg/$(TARGET)

.PHONY: run
run: all
	@echo -e "\n# Running $(TARGET)..."
	./bin/$(TARGET)

.PHONY: clean
clean:
	@echo -e "\n# Removing build..."
	rm -rf bin

.PHONY: docker
docker:
	@echo -e "\n# Building docker image..."
	docker build . -t torvalds_number:multistage --build-arg GHO_CLIENT_ID --build-arg GHO_CLIENT_SECRET --build-arg GHO_PAT
	@echo -e "\n# Deploying docker image..."
	docker run --rm -p 80:80 torvalds_number:multistage -e GHO_CLIENT_ID -e GHO_CLIENT_SECRET -e GHO_PAT

bin/%: cmd/% pkg/% Makefile
	@echo -e "\n# Building $@..."
	go build -o $@ ./$<
