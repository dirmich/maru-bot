#!/bin/bash

# MaruBot One-Line Installer for Raspberry Pi
# Usage: curl -fsSL https://raw.githubusercontent.com/dirmich/maru-bot/main/install.sh | bash
set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}🤖 Starting MaruBot One-Click Installer...${NC}"

# 1. Check Architecture and OS
if [[ "$(uname -m)" != "aarch64" && "$(uname -m)" != "armv7l" ]]; then
    echo -e "${RED}❌ This script is only for Raspberry Pi (ARM) environments.${NC}"
    exit 1
fi

# 2. Install Required Packages
echo -e "${BLUE}📦 Installing required packages...${NC}"
sudo apt update
sudo apt install -y git make libcamera-apps alsa-utils vlc-plugin-base curl wget

# Install Go (1.24+)
GO_REQUIRED="1.24"
INSTALL_GO=false

    BUILD_ARCH=$(uname -m)
    if [ -f "/usr/local/go/bin/go" ]; then
        EXISTING_VERSION=$(/usr/local/go/bin/go version | awk '{print $3}' | sed 's/go//')
        # Check if version starts with required version (simple check, e.g. 1.24.0 >= 1.24)
        if [[ "$EXISTING_VERSION" == "$GO_REQUIRED"* ]] || [[ "$EXISTING_VERSION" > "$GO_REQUIRED" ]]; then
            echo -e "${GREEN}✓ Go $EXISTING_VERSION is already installed.${NC}"
            INSTALL_GO=false
        else
            echo -e "${BLUE}ℹ️ Upgrading Go from $EXISTING_VERSION to $GO_REQUIRED+...${NC}"
            INSTALL_GO=true
        fi
    else
        INSTALL_GO=true
    fi

if [ "$INSTALL_GO" = true ]; then
    echo -e "${BLUE}🐹 Installing latest Go $GO_REQUIRED+ ...${NC}"
    ARCH=$(uname -m)
    BITS=$(getconf LONG_BIT)
    if [ "$ARCH" = "aarch64" ] && [ "$BITS" = "64" ]; then GO_ARCH="arm64"; else GO_ARCH="armv6l"; fi
    WGET_URL="https://go.dev/dl/go1.24.0.linux-$GO_ARCH.tar.gz"
    wget -O go_dist.tar.gz "$WGET_URL"
    sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go_dist.tar.gz
    rm go_dist.tar.gz
    if ! grep -q "/usr/local/go/bin" ~/.bashrc; then echo 'export PATH=/usr/local/go/bin:$PATH' >> ~/.bashrc; fi
    export PATH=/usr/local/go/bin:$PATH
fi

# 3. Install Bun or Node.js (for Web Admin)
# Bun only supports ARM64 64-bit OS. Node.js will be used for 32-bit OS.
USE_BUN=false
BITS=$(getconf LONG_BIT)
if [[ "$(uname -m)" = "aarch64" && "$BITS" = "64" ]]; then
    if ! command -v bun >/dev/null 2>&1; then
        echo -e "${BLUE}🍞 Installing Bun for Web Admin...${NC}"
        curl -fsSL https://bun.sh/install | bash
        export BUN_INSTALL="$HOME/.bun"
        export PATH="$BUN_INSTALL/bin:$PATH"
    fi
    
    # Check installation and support
    if [ -f "$HOME/.bun/bin/bun" ] && "$HOME/.bun/bin/bun" --version >/dev/null 2>&1; then
        USE_BUN=true
    else
        echo -e "${RED}⚠️ Bun is not installed or not supported in this environment. Switching to Node.js.${NC}"
    fi
else
    echo -e "${BLUE}ℹ️ 32-bit OS or non-ARM64 environment detected. Using Node.js instead of Bun.${NC}"
fi

# Install Node.js if Bun is not used
if [ "$USE_BUN" = false ]; then
    if ! command -v node >/dev/null 2>&1; then
        echo -e "${BLUE}📦 Installing Node.js and NPM...${NC}"
        curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
        sudo apt-get install -y nodejs
    fi
fi

# 4. Clone Source Code
INSTALL_DIR="$HOME/marubot"
if [ -d "$INSTALL_DIR" ]; then
    echo -e "${BLUE}🔄 Updating to latest source code...${NC}"
    cd "$INSTALL_DIR"
    git pull
else
    echo -e "${BLUE}📂 Cloning MaruBot source from GitHub...${NC}"
    git clone --depth 1 https://github.com/dirmich/maru-bot.git "$INSTALL_DIR"
    cd "$INSTALL_DIR"
fi

# 5. Build Engine (with Embedded Web Admin)
echo -e "${BLUE}🛠️ Building MaruBot engine...${NC}"

# 5-1. Build Web Admin first
echo -e "${BLUE}    🏗️ Building Web Admin (Vite)...${NC}"
cd "$INSTALL_DIR/web-admin"

if [ "$USE_BUN" = true ]; then
    echo -e "${BLUE}    🍞 Installing web dependencies with Bun...${NC}"
    "$HOME/.bun/bin/bun" install
    echo -e "${BLUE}    ⚛️ Building frontend assets with Bun...${NC}"
    "$HOME/.bun/bin/bun" run build
else
    echo -e "${BLUE}    📦 Installing web dependencies with NPM...${NC}"
    npm install --legacy-peer-deps
    echo -e "${BLUE}    ⚛️ Building frontend assets with NPM...${NC}"
    npm run build
fi

# 5-2. Embed Dist to Go Source
echo -e "${BLUE}    📥 Embedding Web Admin into Go binary...${NC}"
# Copy build output to Go embedding location
mkdir -p "$INSTALL_DIR/cmd/marubot/dashboard/dist"
cp -r dist/* "$INSTALL_DIR/cmd/marubot/dashboard/dist/"

# 5-3. Go Build
cd "$INSTALL_DIR"
go mod tidy
make build

# 6. Install System and Deploy Resources
echo -e "${BLUE}🏗️ Installing system and deploying resources...${NC}"

# 6-1. Install Executable (System-wide)
if [ -f "build/marubot" ]; then
    echo "  📦 Copying executable to /usr/local/bin/marubot..."
    sudo cp build/marubot /usr/local/bin/
    sudo chmod +x /usr/local/bin/marubot
else
    echo -e "${RED}❌ marubot executable not found. Build failed.${NC}"
    exit 1
fi

# 6-2. Configure Resource Directory (~/.marubot)
RESOURCE_DIR="$HOME/.marubot"
mkdir -p "$RESOURCE_DIR"

echo "  📂 Setting up resources in ~/.marubot..."

# (1) Config
mkdir -p "$RESOURCE_DIR/config"
# Maintain existing config (copy only if it doesn't exist)
if [ ! -f "$RESOURCE_DIR/config.json" ]; then
    cp config/maru-config.json "$RESOURCE_DIR/config.json"
fi

# (2) Skills, Tools
# Removing existing folders and copying latest (update)
rm -rf "$RESOURCE_DIR/skills" "$RESOURCE_DIR/tools"
cp -r skills "$RESOURCE_DIR/"
if [ -d "tools" ]; then cp -r tools "$RESOURCE_DIR/"; fi

# (3) Web Admin (Clean up legacy files)
if [ -d "$RESOURCE_DIR/web-admin" ]; then
    echo "  🧹 Removing legacy standalone Web Admin files..."
    rm -rf "$RESOURCE_DIR/web-admin"
fi

# 7. Run Hardware Setup Script
chmod +x maru-setup.sh
# maru-setup.sh uses 'marubot' command, which should be available in /usr/local/bin.
./maru-setup.sh

# 8. Register PATH (only for Bun, MaruBot already in /usr/local/bin)
if [ "$USE_BUN" = true ]; then
    if ! grep -q "BUN_INSTALL" ~/.bashrc; then
        echo "export BUN_INSTALL=\"\$HOME/.bun\"" >> ~/.bashrc
        echo "export PATH=\"\$BUN_INSTALL/bin:\$PATH\"" >> ~/.bashrc
    fi
fi

# Clean up old PATH settings
if grep -q "marubot/build" ~/.bashrc; then
    echo "  🧹 Cleaning up old PATH settings from .bashrc..."
    sed -i '/marubot\/build/d' ~/.bashrc
fi

# 9. Migrate Existing Config (Relative -> Absolute)
if [ -f "$RESOURCE_DIR/config.json" ]; then
    if grep -q "\./workspace" "$RESOURCE_DIR/config.json"; then
        echo "  🔄 Updating workspace path in config.json to ~/.marubot/workspace..."
        sed -i 's|"\./workspace"|"~/.marubot/workspace"|g' "$RESOURCE_DIR/config.json"
    fi
fi

# 10. Consolidate Home Directory Folders
for dir in "workspace" "sessions" "extensions"; do
    if [ -d "$HOME/$dir" ]; then
        echo "  📦 Consolidating $dir folder from incorrect location to ~/.marubot/$dir..."
        mkdir -p "$RESOURCE_DIR/$dir"
        cp -an "$HOME/$dir/." "$RESOURCE_DIR/$dir/" 2>/dev/null || true
        rm -rf "$HOME/$dir"
    fi
done

echo -e "\n${GREEN}🎉 MaruBot installation complete!${NC}"
echo -e "🧹 Automatically cleaning up the source folder ($INSTALL_DIR)..."
cd "$HOME"
rm -rf "$INSTALL_DIR"

echo -e "Command: ${BLUE}marubot agent${NC} (Console Chat)"
echo -e "Dashboard: ${BLUE}marubot dashboard${NC} (Web Admin)"
