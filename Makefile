TARGET=webserver

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
	@./bin/$(TARGET)

.PHONY: clean
clean:
	@echo -e "\n# Removing build..."
	rm -rf bin

bin/%: cmd/% pkg/% Makefile
	@echo -e "\n# Building $@..."
	go build -o $@ ./$<
