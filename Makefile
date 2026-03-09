APP := proxy-sw
PREFIX ?= $(HOME)/.local
BIN_DIR := $(PREFIX)/bin

.PHONY: build test install-local uninstall-local

build:
	go build ./cmd/$(APP)

test:
	go test ./...

install-local:
	mkdir -p "$(BIN_DIR)"
	go build -o "$(BIN_DIR)/$(APP)" ./cmd/$(APP)
	@echo "Installed $(APP) to $(BIN_DIR)/$(APP)"
	@echo "Make sure $(BIN_DIR) is in your PATH"

uninstall-local:
	rm -f "$(BIN_DIR)/$(APP)"
	@echo "Removed $(BIN_DIR)/$(APP)"
