BUILD_DATE := $(shell date +"%Y-%m-%dT%H:%M:%S-%Z")
BUILD_DIR := $(shell pwd)/build
GIT_COMMIT_HASH := $(shell git describe --dirty --always)
GIT_VERSION_TAG := $(shell git tag --sort=-version:refname | head -n 1)
GOARCH := $(shell go env GOARCH)
GOOS := $(shell go env GOOS)
SOURCES := $(shell find . -name '*.go')
TARGET ?= $(shell uname -m)

BINARY := myhome-presence-$(GOOS)-$(GOARCH)
IMAGE := quay.io/touchardv/myhome-presence
LD_ARGS := -ldflags "-X main.buildDate=$(BUILD_DATE) -X main.gitCommitHash=$(GIT_COMMIT_HASH) -X main.gitVersionTag=$(GIT_VERSION_TAG)"
TAG := latest

ifeq ($(GOARCH), arm)
 DOCKER_BUILDX_PLATFORM := linux/arm/v7
else ifeq ($(GOARCH), arm64)
 DOCKER_BUILDX_PLATFORM := linux/arm64/v8
else ifeq ($(GOARCH), amd64)
 DOCKER_BUILDX_PLATFORM := linux/amd64
endif

.DEFAULT_GOAL := build
.PHONY: build
build: $(BUILD_DIR)/$(BINARY)

build-image: $(BUILD_DIR)/$(BINARY)
	docker buildx build --progress plain \
	--platform $(DOCKER_BUILDX_PLATFORM) \
	--tag $(IMAGE):$(TAG) --load -f deployment/Dockerfile .

$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

$(BUILD_DIR)/$(BINARY): $(BUILD_DIR) $(SOURCES) internal/api/openapi.yaml.tmpl
	go build $(LD_ARGS) -o $(BUILD_DIR)/$(BINARY) ./cmd/myhome-presence

.PHONY: clean
clean:
	rm -rf $(BUILD_DIR)
	docker image rm -f $(IMAGE)/$(TAG)
	go clean

deploy-systemd-service: $(BUILD_DIR)/$(BINARY) test
	scp $(BUILD_DIR)/$(BINARY) $(TARGET):/tmp/$(BINARY)
	ssh $(TARGET) sudo systemctl stop myhome-presence
	ssh $(TARGET) sudo cp /tmp/$(BINARY) /usr/bin/myhome-presence
	ssh $(TARGET) sudo setcap 'cap_net_raw,cap_net_admin=eip' /usr/bin/myhome-presence
	ssh $(TARGET) sudo systemctl start myhome-presence

run: $(BUILD_DIR)/$(BINARY)
	$(BUILD_DIR)/$(BINARY) --config-location=`pwd` --data-location=`pwd` --log-level=debug

run-image:
	docker run -it --rm $(IMAGE):$(TAG)

setup-systemd-service:
	ssh $(TARGET) sudo mkdir -p /etc/myhome /var/lib/myhome /var/log/myhome
	ssh $(TARGET) sudo chown -R pi:pi /etc/myhome /var/lib/myhome /var/log/myhome
	scp deployment/myhome-presence.*  $(TARGET):/tmp
	ssh $(TARGET) sudo mv /tmp/myhome-presence.conf /etc/sysctl.d/myhome-presence.conf
	ssh $(TARGET) sudo mv /tmp/myhome-presence.service /etc/systemd/system/myhome-presence.service
	ssh $(TARGET) sudo systemctl enable myhome-presence

test:
	go test -v -cover -timeout 10s ./...
