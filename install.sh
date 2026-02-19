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

if [ -f "/usr/local/go/bin/go" ]; then
    EXISTING_VERSION=$(/usr/local/go/bin/go version | awk '{print $3}' | sed 's/go//')
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
fi

# Ensure /usr/local/go/bin is at the front of PATH for this script session
export PATH=/usr/local/go/bin:$PATH
GO_CMD="/usr/local/go/bin/go"
if [ ! -f "$GO_CMD" ]; then GO_CMD="go"; fi

# 3. Clone Source Code
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

# 4. Install Optional Web Admin Build Tools (Only if source is present)
USE_BUN=false
HAS_WEB_SOURCE=false
if [ -d "$INSTALL_DIR/web-admin" ]; then
    HAS_WEB_SOURCE=true
    echo -e "${BLUE}⚛️ Web Admin source detected. Preparing build tools...${NC}"
    
    BITS=$(getconf LONG_BIT)
    if [[ "$(uname -m)" = "aarch64" && "$BITS" = "64" ]]; then
        if ! command -v bun >/dev/null 2>&1; then
            echo -e "${BLUE}🍞 Installing Bun for Web Admin...${NC}"
            curl -fsSL https://bun.sh/install | bash
            export BUN_INSTALL="$HOME/.bun"
            export PATH="$BUN_INSTALL/bin:$PATH"
        fi
        if [ -f "$HOME/.bun/bin/bun" ] && "$HOME/.bun/bin/bun" --version >/dev/null 2>&1; then
            USE_BUN=true
        fi
    fi

    if [ "$USE_BUN" = false ]; then
        if ! command -v node >/dev/null 2>&1; then
            echo -e "${BLUE}📦 Installing Node.js and NPM...${NC}"
            curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
            sudo apt-get install -y nodejs
        fi
    fi
else
    echo -e "${BLUE}ℹ️ Using pre-built Web Admin (Single Binary Mode).${NC}"
fi

# 5. Build Engine
echo -e "${BLUE}🛠️ Building MaruBot engine...${NC}"

# 5-1. Build Web Admin (If source exists)
if [ "$HAS_WEB_SOURCE" = true ]; then
    echo -e "${BLUE}    🏗️ Building Web Admin (Vite)...${NC}"
    cd "$INSTALL_DIR/web-admin"
    if [ "$USE_BUN" = true ]; then
        "$HOME/.bun/bin/bun" install
        "$HOME/.bun/bin/bun" run build
    else
        npm install --legacy-peer-deps
        npm run build
    fi
    # Embed Dist to Go Source
    echo -e "${BLUE}    📥 Embedding Web Admin into Go source...${NC}"
    mkdir -p "$INSTALL_DIR/cmd/marubot/dashboard/dist"
    cp -r dist/* "$INSTALL_DIR/cmd/marubot/dashboard/dist/"
    cd "$INSTALL_DIR"
fi

# 5-2. Go Build
echo -e "${BLUE}    🐹 Running Go build...${NC}"
$GO_CMD mod tidy
make GO="$GO_CMD" build

# 6. Install System and Deploy Resources
echo -e "${BLUE}🏗️ Installing system and deploying resources...${NC}"

if [ -f "build/marubot" ]; then
    echo "  📦 Copying executable to /usr/local/bin/marubot..."
    sudo cp build/marubot /usr/local/bin/
    sudo chmod +x /usr/local/bin/marubot
else
    echo -e "${RED}❌ marubot executable not found. Build failed.${NC}"
    exit 1
fi

RESOURCE_DIR="$HOME/.marubot"
mkdir -p "$RESOURCE_DIR"
mkdir -p "$RESOURCE_DIR/config"
if [ ! -f "$RESOURCE_DIR/config.json" ]; then
    cp config/maru-config.json "$RESOURCE_DIR/config.json"
fi

rm -rf "$RESOURCE_DIR/skills" "$RESOURCE_DIR/tools"
cp -r skills "$RESOURCE_DIR/"
if [ -d "tools" ]; then cp -r tools "$RESOURCE_DIR/"; fi

# Clean up legacy
if [ -d "$RESOURCE_DIR/web-admin" ]; then
    rm -rf "$RESOURCE_DIR/web-admin"
fi

# 7. Hardware Setup
chmod +x maru-setup.sh
./maru-setup.sh

# 8. Finalize PATH and Config
if grep -q "marubot/build" ~/.bashrc 2>/dev/null; then
    sed -i '/marubot\/build/d' ~/.bashrc
fi

# Migrate Workspace Path
if [ -f "$RESOURCE_DIR/config.json" ]; then
    if grep -q "\./workspace" "$RESOURCE_DIR/config.json"; then
        sed -i 's|"\./workspace"|"~/.marubot/workspace"|g' "$RESOURCE_DIR/config.json"
    fi
fi

# Consolidate Folders
for dir in "workspace" "sessions" "extensions"; do
    if [ -d "$HOME/$dir" ]; then
        mkdir -p "$RESOURCE_DIR/$dir"
        cp -an "$HOME/$dir/." "$RESOURCE_DIR/$dir/" 2>/dev/null || true
        rm -rf "$HOME/$dir"
    fi
done

echo -e "\n${GREEN}🎉 MaruBot installation complete!${NC}"
echo -e "🧹 Automatically cleaning up the source folder ($INSTALL_DIR)..."
cd "$HOME"
rm -rf "$INSTALL_DIR"

echo -e "Command: ${BLUE}marubot agent${NC}"
echo -e "Dashboard: ${BLUE}marubot dashboard${NC}"
