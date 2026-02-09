# P2P Chat Makefile
# Provides build, install, and uninstall targets for Linux systems

# Variables
BINARY_NAME = p2pchat
SOURCE_DIR = cmd/p2pchat
INSTALL_PATH = /usr/local/bin
GO_VERSION = 1.21

# Default target
.PHONY: all
all: build

# Build the binary
.PHONY: build
build:
	@echo "🔨 Building P2P Chat..."
	@go build -o $(BINARY_NAME) $(SOURCE_DIR)/main.go
	@echo "✅ Built $(BINARY_NAME)"

# Install system-wide (Linux only)
.PHONY: install
install: build
	@echo "🚀 Installing P2P Chat system-wide..."
	@if [ "$$(uname)" != "Linux" ]; then \
		echo "❌ System install only supported on Linux"; \
		echo "💡 Use './$(BINARY_NAME)' directly on other platforms"; \
		exit 1; \
	fi
	@if [ "$$(id -u)" -eq 0 ]; then \
		cp $(BINARY_NAME) $(INSTALL_PATH)/$(BINARY_NAME); \
		chmod +x $(INSTALL_PATH)/$(BINARY_NAME); \
		echo "✅ Installed $(BINARY_NAME) to $(INSTALL_PATH)"; \
		echo "🎉 You can now run 'p2pchat' from anywhere!"; \
	else \
		echo "❌ Installation requires root privileges"; \
		echo "💡 Run: sudo make install"; \
		exit 1; \
	fi

# Uninstall system-wide
.PHONY: uninstall
uninstall:
	@echo "🗑️  Uninstalling P2P Chat..."
	@if [ -f "$(INSTALL_PATH)/$(BINARY_NAME)" ]; then \
		if [ "$$(id -u)" -eq 0 ]; then \
			rm -f $(INSTALL_PATH)/$(BINARY_NAME); \
			echo "✅ Removed $(BINARY_NAME) from $(INSTALL_PATH)"; \
		else \
			echo "❌ Uninstallation requires root privileges"; \
			echo "💡 Run: sudo make uninstall"; \
			exit 1; \
		fi \
	else \
		echo "ℹ️  $(BINARY_NAME) not found in $(INSTALL_PATH)"; \
	fi

# Clean build artifacts
.PHONY: clean
clean:
	@echo "🧹 Cleaning build artifacts..."
	@rm -f $(BINARY_NAME)
	@echo "✅ Cleaned"

# Run tests
.PHONY: test
test:
	@echo "🧪 Running tests..."
	@go test ./...
	@echo "✅ Tests passed"

# Development build and run
.PHONY: run
run: build
	@echo "🚀 Starting P2P Chat..."
	@./$(BINARY_NAME)

# Development with debug
.PHONY: debug
debug: build
	@echo "🔍 Starting P2P Chat in debug mode..."
	@./$(BINARY_NAME) -debug

# Show help
.PHONY: help
help:
	@echo "P2P Chat Build System"
	@echo "===================="
	@echo ""
	@echo "Targets:"
	@echo "  build     - Build the p2pchat binary"
	@echo "  install   - Install system-wide (Linux, requires sudo)"
	@echo "  uninstall - Remove system installation (requires sudo)"
	@echo "  clean     - Remove build artifacts"
	@echo "  test      - Run all tests"
	@echo "  run       - Build and run locally"
	@echo "  debug     - Build and run with debug logging"
	@echo "  help      - Show this help message"
	@echo ""
	@echo "Examples:"
	@echo "  make build               # Build locally"
	@echo "  sudo make install        # Install system-wide (Linux)"
	@echo "  make run                 # Quick test run"
	@echo "  ./p2pchat                # Run local build directly"
	@echo ""
	@echo "Portfolio Note:"
	@echo "This project is designed for local compilation and evaluation."
	@echo "System installation is provided as a convenience for Linux users."

# Check if system installation exists
.PHONY: status
status:
	@echo "P2P Chat Installation Status"
	@echo "============================"
	@if [ -f "$(INSTALL_PATH)/$(BINARY_NAME)" ]; then \
		echo "✅ System installation found at $(INSTALL_PATH)/$(BINARY_NAME)"; \
		echo "🎉 Run 'p2pchat' from anywhere"; \
	else \
		echo "❌ No system installation found"; \
		echo "💡 Run 'sudo make install' or use './p2pchat' locally"; \
	fi
	@if [ -f "./$(BINARY_NAME)" ]; then \
		echo "✅ Local build found: ./$(BINARY_NAME)"; \
	else \
		echo "❌ No local build found - run 'make build'"; \
	fi
