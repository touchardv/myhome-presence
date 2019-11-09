BINARY := myhome-presence
BUILD_DIR := $(shell pwd)/build
SOURCES = $(shell find . -name '*.go')

.PHONY: build
build: $(BUILD_DIR)/$(BINARY)

$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

$(BUILD_DIR)/$(BINARY): $(BUILD_DIR) $(SOURCES)
	go build -o $(BUILD_DIR)/$(BINARY) .

.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)

run: $(BUILD_DIR)/$(BINARY)
	$(BUILD_DIR)/$(BINARY)	

test:
	go test -v -cover ./...
