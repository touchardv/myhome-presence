BINARY := myhome-presence
BUILD_DATE := $(shell date +"%Y-%m-%dT%H:%M:%S-%Z")
BUILD_DIR := $(shell pwd)/build
GIT_COMMIT_HASH := $(shell git describe --dirty --always)
GIT_VERSION_TAG := $(shell git tag --sort=-version:refname | head -n 1)
SOURCES = $(shell find . -name '*.go')

LD_ARGS := -ldflags "-X main.buildDate=$(BUILD_DATE) -X main.gitCommitHash=$(GIT_COMMIT_HASH) -X main.gitVersionTag=$(GIT_VERSION_TAG)"

.PHONY: build
build: $(BUILD_DIR)/$(BINARY)

build-linux: $(BUILD_DIR)/$(BINARY)-linux-arm

$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

$(BUILD_DIR)/$(BINARY): $(BUILD_DIR) $(SOURCES) internal/api/openapi.yaml.tmpl
	go build $(LD_ARGS) -o $(BUILD_DIR)/$(BINARY) ./cmd/myhome-presence

$(BUILD_DIR)/$(BINARY)-linux-arm: $(SOURCES) internal/api/openapi.yaml.tmpl
	$(shell export GO111MODULE=on; export GOOS=linux; export GOARCH=arm64; go build  $(LD_ARGS) -o $(BUILD_DIR)/$(BINARY)-linux-arm ./cmd/myhome-presence)

.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)

copy: $(BUILD_DIR)/$(BINARY)-linux-arm
	scp $(BUILD_DIR)/$(BINARY)-linux-arm $(TARGET):/tmp/$(BINARY)-linux-arm

deploy: test copy
	ssh $(TARGET) sudo systemctl stop myhome-presence
	ssh $(TARGET) sudo cp /tmp/$(BINARY)-linux-arm /usr/bin/myhome-presence
	ssh $(TARGET) sudo setcap 'cap_net_raw,cap_net_admin=eip' /usr/bin/myhome-presence
	ssh $(TARGET) sudo systemctl start myhome-presence

run: $(BUILD_DIR)/$(BINARY)
	$(BUILD_DIR)/$(BINARY) --config-location=`pwd` --data-location=`pwd` --log-level=debug

setup:
	ssh $(TARGET) sudo mkdir -p /etc/myhome /var/log/myhome
	ssh $(TARGET) sudo chown -R pi:pi /etc/myhome /var/log/myhome
	scp deployment/myhome-presence.*  $(TARGET):/tmp
	ssh $(TARGET) sudo mv /tmp/myhome-presence.conf /etc/sysctl.d/myhome-presence.conf
	ssh $(TARGET) sudo mv /tmp/myhome-presence.service /etc/systemd/system/myhome-presence.service
	ssh $(TARGET) sudo systemctl enable myhome-presence

test:
	go test -v -cover -timeout 10s ./...
