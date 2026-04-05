.PHONY: all build install uninstall clean help test

# Build variables
BINARY_NAME=marubot
BUILD_DIR=build
CMD_DIR=cmd/$(BINARY_NAME)
MAIN_GO=$(CMD_DIR)/main.go

# Version
VERSION?=$(shell git describe --tags --exact-match 2>/dev/null || echo "")
# If VERSION is empty, main.go's harcoded version will be used because we'll conditionalize LDFLAGS
BUILD_TIME=$(shell date +%FT%T%z)
# Base LDFLAGS
LDFLAGS_BASE=$(if $(VERSION),-X main.Version=$(VERSION) -X main.buildTime=$(BUILD_TIME),-X main.buildTime=$(BUILD_TIME))
# For Linux/Darwin (Console)
LDFLAGS_CONSOLE=-ldflags "$(LDFLAGS_BASE)"
# For Windows (GUI - to hide CMD window)
LDFLAGS_WINDOWSGUI=-ldflags "$(LDFLAGS_BASE) -H windowsgui"
# OS detection
UNAME_S:=$(shell uname -s)
UNAME_M:=$(shell uname -m)

# CGO settings
ifeq ($(UNAME_S),Darwin)
	CGO_ENABLED=1
else
	CGO_ENABLED=0
endif
export CGO_ENABLED

# Go variables
GO?=go
GOFLAGS?=-v

# MARUBOT_HOME is the base directory for resources and binary
MARUBOT_HOME?=$(HOME)/.marubot

# Installation
INSTALL_BIN_DIR=$(MARUBOT_HOME)/bin
WORKSPACE_DIR?=$(MARUBOT_HOME)/workspace
WORKSPACE_SKILLS_DIR=$(WORKSPACE_DIR)/skills
BUILTIN_SKILLS_DIR=$(CURDIR)/skills

# OS detection

# Platform-specific settings
ifeq ($(UNAME_S),Linux)
	PLATFORM=linux
	ifeq ($(UNAME_M),x86_64)
		ARCH=amd64
	else ifeq ($(UNAME_M),aarch64)
		ifeq ($(shell getconf LONG_BIT),64)
			ARCH=arm64
		else
			ARCH=arm
		endif
	else ifeq ($(UNAME_M),riscv64)
		ARCH=riscv64
	else
		ARCH=arm
	endif
else ifeq ($(UNAME_S),Darwin)
	PLATFORM=darwin
	ifeq ($(UNAME_M),x86_64)
		ARCH=amd64
	else ifeq ($(UNAME_M),arm64)
		ARCH=arm64
	else
		ARCH=$(UNAME_M)
	endif
else
	PLATFORM=$(UNAME_S)
	ARCH=$(UNAME_M)
endif

BINARY_PATH=$(BUILD_DIR)/$(BINARY_NAME)-$(PLATFORM)-$(ARCH)

# internal helper to sync UI assets
sync-ui:
	@echo "Checking web-admin assets..."
	@if [ -d "web-admin" ]; then \
		echo "Building UI (Clean build)..."; \
		cd web-admin && npm run build && cd ..; \
		echo "Syncing web-admin assets..."; \
		rm -rf cmd/marubot/dashboard/dist; \
		mkdir -p cmd/marubot/dashboard/dist; \
		cp -rv web-admin/dist/* cmd/marubot/dashboard/dist/; \
	else \
		echo "Skipping UI build (source not found). Checking for pre-built assets..."; \
		if [ ! -f "cmd/marubot/dashboard/dist/index.html" ]; then \
			echo "Error: cmd/marubot/dashboard/dist/index.html is missing!"; \
			exit 1; \
		fi; \
		echo "✓ Pre-built assets found in dashboard/dist"; \
	fi

## build: Build the marubot binary for current platform
build: sync-ui
	@echo "Building $(BINARY_NAME) for $(PLATFORM)/$(ARCH)..."
	@mkdir -p $(BUILD_DIR)
	$(GO) build $(GOFLAGS) $(LDFLAGS_CONSOLE) -o $(BINARY_PATH) ./$(CMD_DIR)
	@echo "Build complete: $(BINARY_PATH)"
	@ln -sf $(BINARY_NAME)-$(PLATFORM)-$(ARCH) $(BUILD_DIR)/$(BINARY_NAME)

## build-all: Build marubot for Windows and macOS
build-all: sync-ui
	@echo "Building for Windows and macOS..."
	@mkdir -p $(BUILD_DIR)
	@# Windows
	CGO_ENABLED=0 GOOS=windows GOARCH=amd64 $(GO) build $(LDFLAGS_WINDOWSGUI) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-amd64.exe ./$(CMD_DIR)
	CGO_ENABLED=0 GOOS=windows GOARCH=386 $(GO) build $(LDFLAGS_WINDOWSGUI) -o $(BUILD_DIR)/$(BINARY_NAME)-windows-386.exe ./$(CMD_DIR)
	@# Darwin
	@echo "Building for macOS (CGO required for Tray Icon)..."
	CGO_ENABLED=1 GOOS=darwin GOARCH=amd64 $(GO) build $(LDFLAGS_CONSOLE) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-amd64 ./$(CMD_DIR)
	CGO_ENABLED=1 GOOS=darwin GOARCH=arm64 $(GO) build $(LDFLAGS_CONSOLE) -o $(BUILD_DIR)/$(BINARY_NAME)-darwin-arm64 ./$(CMD_DIR)
	@echo "Packaging DMGs..."
	@$(MAKE) package-dmg
	@echo "All targeted builds and packages complete"

## package-win: Collect Windows binaries (No ZIP)
package-win: build-all
	@echo "📦 Collecting Windows binaries (x64 & x86)..."
	@mkdir -p build/
	@rm -f build/*.zip
	@echo "✓ Windows binaries ready in build/."

## package-dmg: Package macOS binaries into DMG files
package-dmg:
	@echo "📦 Packaging macOS DMGs..."
	@chmod +x scripts/build_dmg.sh
	@./scripts/build_dmg.sh amd64
	@./scripts/build_dmg.sh arm64
	@rm -f build/*.zip
	@echo "✓ macOS DMGs created."

## install: Install marubot to system and copy builtin skills
install: build
	@echo "Installing $(BINARY_NAME)..."
	@mkdir -p $(INSTALL_BIN_DIR)
	@cp $(BUILD_DIR)/$(BINARY_NAME) $(INSTALL_BIN_DIR)/$(BINARY_NAME)
	@chmod +x $(INSTALL_BIN_DIR)/$(BINARY_NAME)
	@echo "Installed binary to $(INSTALL_BIN_DIR)/$(BINARY_NAME)"
	@echo "Installing builtin skills to $(WORKSPACE_SKILLS_DIR)..."
	@mkdir -p $(WORKSPACE_SKILLS_DIR)
	@for skill in $(BUILTIN_SKILLS_DIR)/*/; do \
		if [ -d "$$skill" ]; then \
			skill_name=$$(basename "$$skill"); \
			if [ -f "$$skill/SKILL.md" ]; then \
				cp -r "$$skill" $(WORKSPACE_SKILLS_DIR); \
				echo "  ✓ Installed skill: $$skill_name"; \
			fi; \
		fi; \
	done
	@echo "Installation complete!"

## install-skills: Install builtin skills to workspace
install-skills:
	@echo "Installing builtin skills to $(WORKSPACE_SKILLS_DIR)..."
	@mkdir -p $(WORKSPACE_SKILLS_DIR)
	@for skill in $(BUILTIN_SKILLS_DIR)/*/; do \
		if [ -d "$$skill" ]; then \
			skill_name=$$(basename "$$skill"); \
			if [ -f "$$skill/SKILL.md" ]; then \
				mkdir -p $(WORKSPACE_SKILLS_DIR)/$$skill_name; \
				cp -r "$$skill" $(WORKSPACE_SKILLS_DIR); \
				echo "  ✓ Installed skill: $$skill_name"; \
			fi; \
		fi; \
	done
	@echo "Skills installation complete!"

## public: Sync public files to ../marubot (for public repo maintenance)
public: package-win package-dmg
	@echo "🚀 Syncing to public repository (Windows: EXE only, macOS: DMG only)..."
	@chmod +x scripts/publish.sh
	@./scripts/publish.sh

## uninstall: Remove marubot from system
uninstall:
	@echo "Uninstalling $(BINARY_NAME)..."
	-@$(INSTALL_BIN_DIR)/$(BINARY_NAME) uninstall
	@rm -f $(INSTALL_BIN_DIR)/$(BINARY_NAME)
	@echo "Removed binary from $(INSTALL_BIN_DIR)/$(BINARY_NAME)"
	@echo "Note: Core configuration and workspace are kept unless you run 'make uninstall-all'"

## uninstall-all: Remove marubot and all data
uninstall-all:
	@echo "Removing workspace and skills..."
	@rm -rf $(MARUBOT_HOME)
	@echo "Removed workspace: $(MARUBOT_HOME)"
	@echo "Complete uninstallation done!"

## clean: Remove build artifacts
clean:
	@echo "Cleaning build artifacts..."
	@rm -rf $(BUILD_DIR)
	@rm -f *.zip
	@echo "Clean complete"

## fmt: Format Go code
fmt:
	@$(GO) fmt ./...

## deps: Update dependencies
deps:
	@$(GO) get -u ./...
	@$(GO) mod tidy

## run: Build and run marubot
run: build
	@$(BUILD_DIR)/$(BINARY_NAME) $(ARGS)

## help: Show this help message
help:
	@echo "marubot Makefile"
	@echo ""
	@echo "Usage:"
	@echo "  make [target]"
	@echo ""
	@echo "Targets:"
	@grep -E '^## ' $(MAKEFILE_LIST) | sed 's/## /  /'
	@echo ""
	@echo "Examples:"
	@echo "  make build              # Build for current platform"
	@echo "  make install            # Install to $(INSTALL_BIN_DIR)"
	@echo "  make uninstall          # Remove from $(INSTALL_BIN_DIR)"
	@echo "  make install-skills     # Install skills to workspace"
	@echo ""
	@echo "Environment Variables:"
	@echo "  INSTALL_PREFIX          # Installation prefix (default: /usr/local)"
	@echo "  WORKSPACE_DIR           # Workspace directory (default: ~/.marubot/workspace)"
	@echo "  VERSION                 # Version string (default: git describe)"
	@echo ""
	@echo "Current Configuration:"
	@echo "  Platform: $(PLATFORM)/$(ARCH)"
	@echo "  Binary: $(BINARY_PATH)"
	@echo "  Install Prefix: $(INSTALL_PREFIX)"
	@echo "  Workspace: $(WORKSPACE_DIR)"
