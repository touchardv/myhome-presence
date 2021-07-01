BINARY := myhome-presence
BUILD_DIR := $(shell pwd)/build
SOURCES = $(shell find . -name '*.go')

.PHONY: build
build: $(BUILD_DIR)/$(BINARY)

$(BUILD_DIR):
	mkdir -p $(BUILD_DIR)

$(BUILD_DIR)/$(BINARY): $(BUILD_DIR) $(SOURCES) docs/openapi.yaml.tmpl
	go build -o $(BUILD_DIR)/$(BINARY) .

$(BUILD_DIR)/$(BINARY)-linux-arm: $(SOURCES) docs/openapi.yaml.tmpl
	$(shell export GO111MODULE=on; export GOOS=linux; export GOARCH=arm; export GOARM=5; go build -o $(BUILD_DIR)/$(BINARY)-linux-arm .)

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
	$(BUILD_DIR)/$(BINARY) --config-location=`pwd` --log-level=debug

setup:
	ssh $(TARGET) sudo mkdir -p /etc/myhome /var/log/myhome
	ssh $(TARGET) sudo chown -R pi:pi /etc/myhome /var/log/myhome
	scp myhome-presence.*  $(TARGET):/tmp
	ssh $(TARGET) sudo mv /tmp/myhome-presence.conf /etc/sysctl.d/myhome-presence.conf
	ssh $(TARGET) sudo mv /tmp/myhome-presence.service /etc/systemd/system/myhome-presence.service
	ssh $(TARGET) sudo systemctl enable myhome-presence

test:
	go test -v -cover ./...
