#!/bin/bash

# MaruBot One-Line Installer for Raspberry Pi
# Usage: curl -fsSL https://raw.githubusercontent.com/dirmich/maru-bot/main/install.sh | bash
set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[0;33m'
NC='\033[0m'

# 0. Parse Arguments
FORCE_RPI=false
for arg in "$@"; do
    if [ "$arg" == "--rpi" ]; then
        FORCE_RPI=true
    fi
done

echo -e "${BLUE}[*] Starting MaruBot One-Click Installer...${NC}"
if [ "$FORCE_RPI" = true ]; then
    echo -e "${BLUE}[i] Raspberry Pi mode forced via parameter.${NC}"
fi

# 0. Language Selection
echo "Select Application Language:"
echo "1) English (en)"
echo "2) Korean (ko)"
echo "3) Japanese (ja)"

# Check if running interactively or verify /dev/tty availability
if [ -c /dev/tty ]; then
    read -p "Select (1-3) [Default: 1]: " LANG_CHOICE < /dev/tty
else
    if [ -t 0 ]; then
        read -p "Select (1-3) [Default: 1]: " LANG_CHOICE
    else
        # Fallback for non-interactive environments
        LANG_CHOICE=1
    fi
fi

if [ -z "$LANG_CHOICE" ]; then
    LANG_CHOICE=1
fi

case $LANG_CHOICE in
    2) MARUBOT_LANG="ko" ;;
    3) MARUBOT_LANG="ja" ;;
    *) MARUBOT_LANG="en" ;;
esac

# Check for existing password in config.json
EXISTING_PWD=""
if [ -f "$HOME/.marubot/config.json" ]; then
    # Try to extract existing password using more compatible sed/grep
    EXISTING_PWD=$(grep "\"admin_password\":" "$HOME/.marubot/config.json" | sed -E 's/.*"admin_password":\s*"([^"]+)".*/\1/' || true)
fi

PROMPT_PWD_SUFFIX=" [Default: admin]"
if [ ! -z "$EXISTING_PWD" ]; then
    PROMPT_PWD_SUFFIX=" [Default: (Current)]"
fi

PROMPT_PWD="Set Admin Password for Web Dashboard$PROMPT_PWD_SUFFIX: "

if [ -c /dev/tty ]; then
    read -p "$PROMPT_PWD" MARUBOT_PWD < /dev/tty
else
    if [ -t 0 ]; then
        read -p "$PROMPT_PWD" MARUBOT_PWD
    else
        MARUBOT_PWD=""
    fi
fi

if [ -z "$MARUBOT_PWD" ]; then
    if [ ! -z "$EXISTING_PWD" ]; then
        MARUBOT_PWD="$EXISTING_PWD"
        echo "Keeping existing password."
    else
        MARUBOT_PWD="admin"
        echo "Defaulting to 'admin'. (Initial login requires 'admin')"
    fi
fi

# Messages (English Only for Shell Stability)
MSG_ARCH_ERR="[!] This script is only for Raspberry Pi (ARM) environments."
MSG_PKG_INST="[*] Installing required packages..."
MSG_GO_INST="[*] Installing latest Go..."
MSG_CLONE="[*] Cloning MaruBot source from GitHub..."
MSG_WEB_BUILD="[*] Building Web Admin (Vite)..."
MSG_GO_BUILD="[*] Building MaruBot engine..."
MSG_SUCCESS="[+] MaruBot installation complete!"
MSG_DASHBOARD="Run dashboard: marubot start"
MSG_UPGRADE="[*] Attempting MaruBot upgrade..."

# 1. Check Architecture and OS
IS_PI=false
if [[ "$(uname -m)" == "aarch64" || "$(uname -m)" == "armv7l" ]]; then
    IS_PI=true
fi

if [ "$IS_PI" = false ]; then
    echo -e "${YELLOW}[!] Notice: This environment is not Raspberry Pi (ARM). Hardware-specific features (GPIO, etc.) will be disabled.${NC}"
fi

# 2. Install Required Packages
echo -e "${BLUE}${MSG_PKG_INST}${NC}"
if command -v apt >/dev/null 2>&1; then
    sudo apt update
    sudo apt install -y git make curl wget
    if [ "$IS_PI" = true ]; then
        sudo apt install -y libcamera-apps alsa-utils vlc-plugin-base
    fi
else
    echo -e "${YELLOW}[!] 'apt' not found. Skipping system package installation.${NC}"
fi

# Install Go (1.24+)
GO_REQUIRED="1.24"
INSTALL_GO=false

# 1. Prioritize MaruBot-installed Go or current PATH
if [ -f "/usr/local/go/bin/go" ]; then
    DETECTED_GO="/usr/local/go/bin/go"
elif command -v go >/dev/null 2>&1; then
    DETECTED_GO=$(command -v go)
fi

if [ ! -z "$DETECTED_GO" ]; then
    if ! "$DETECTED_GO" version >/dev/null 2>&1; then
        echo -e "${YELLOW}[!] Existing Go installation at $DETECTED_GO is broken. Forcing reinstall...${NC}"
        INSTALL_GO=true
    else
        EXISTING_VERSION=$("$DETECTED_GO" version | awk '{print $3}' | sed 's/go//' | cut -d. -f1-2)
        REQUIRED_MAJOR_MINOR=$(echo $GO_REQUIRED | cut -d. -f1-2)
        
        # Simple string comparison for major.minor
        if [ "$EXISTING_VERSION" = "$REQUIRED_MAJOR_MINOR" ] || [ "$(printf '%s\n%s' "$REQUIRED_MAJOR_MINOR" "$EXISTING_VERSION" | sort -V | head -n1)" = "$REQUIRED_MAJOR_MINOR" ]; then
            echo -e "${GREEN}[v] Go $EXISTING_VERSION is already installed and sufficient.${NC}"
            INSTALL_GO=false
        else
            echo -e "${BLUE}[i] Upgrading Go from $EXISTING_VERSION to $GO_REQUIRED+...${NC}"
            INSTALL_GO=true
        fi
    fi
else
    INSTALL_GO=true
fi

if [ "$INSTALL_GO" = true ]; then
    echo -e "${BLUE}[*] Installing latest Go $GO_REQUIRED+ ...${NC}"
    ARCH=$(uname -m)
    BITS=$(getconf LONG_BIT)
    if [[ "$ARCH" == "aarch64" || "$ARCH" == "arm64" ]]; then 
        if [ "$BITS" = "64" ]; then
            GO_ARCH="arm64"
        else
            GO_ARCH="armv6l"
        fi
    elif [[ "$ARCH" == *"arm"* || "$ARCH" == *"aarch32"* ]]; then
        GO_ARCH="armv6l"
    elif [[ "$ARCH" == "x86_64" || "$ARCH" == "amd64" ]]; then
        GO_ARCH="amd64"
    elif [[ "$ARCH" == "i686" || "$ARCH" == "i386" ]]; then
        GO_ARCH="386"
    else
        GO_ARCH="armv6l"
        echo -e "${YELLOW}[!] Unknown architecture ($ARCH). Defaulting to armv6l.${NC}"
    fi
    WGET_URL="https://go.dev/dl/go1.24.0.linux-$GO_ARCH.tar.gz"
    echo "Downloading from $WGET_URL ..."
    wget -O go_dist.tar.gz "$WGET_URL"
    sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go_dist.tar.gz
    rm go_dist.tar.gz
    if ! grep -q "/usr/local/go/bin" ~/.bashrc; then echo 'export PATH=/usr/local/go/bin:$PATH' >> ~/.bashrc; fi
fi

if [ -f "/usr/local/go/bin/go" ]; then
    export PATH=/usr/local/go/bin:$PATH
    GO_CMD="/usr/local/go/bin/go"
else
    GO_CMD=$(command -v go)
fi

if [ -z "$GO_CMD" ]; then
    echo -e "${RED}[x] Go not found even after installation attempt.${NC}"
    exit 1
fi

# 3. Clone Source Code
INSTALL_DIR="$HOME/marubot"
DOWNLOAD_ARCHIVE() {
    echo -e "${BLUE}[i] Downloading archive via curl/wget...${NC}"
    mkdir -p "$INSTALL_DIR"
    if command -v curl >/dev/null 2>&1; then
        curl -fsSL https://github.com/dirmich/maru-bot/archive/refs/tags/v0.7.2.2.tar.gz | tar -xz -C "$INSTALL_DIR" --strip-components=1
    elif command -v wget >/dev/null 2>&1; then
        wget -qO- https://github.com/dirmich/maru-bot/archive/refs/tags/v0.7.2.2.tar.gz | tar -xz -C "$INSTALL_DIR" --strip-components=1
    else
        echo -e "${RED}[x] Neither curl nor wget are available.${NC}"
        return 1
    fi
}

if [ -d "$INSTALL_DIR" ]; then
    echo -e "${BLUE}[*] Updating to latest source code...${NC}"
    cd "$INSTALL_DIR"
    if command -v git >/dev/null 2>&1 && [ -d ".git" ]; then
        if ! git pull; then
            echo -e "${YELLOW}[!] Git update failed. Retrying fresh download...${NC}"
            cd "$HOME"
            rm -rf "$INSTALL_DIR"
            if ! DOWNLOAD_ARCHIVE; then exit 1; fi
            cd "$INSTALL_DIR"
        fi
    else
        echo -e "${YELLOW}[!] Not a git repository. Downloading fresh archive...${NC}"
        cd "$HOME"
        rm -rf "$INSTALL_DIR"
        if ! DOWNLOAD_ARCHIVE; then exit 1; fi
        cd "$INSTALL_DIR"
    fi
else
    echo -e "${BLUE}${MSG_CLONE}${NC}"
    if command -v git >/dev/null 2>&1; then
        if ! git clone --depth 1 https://github.com/dirmich/maru-bot.git "$INSTALL_DIR"; then
            echo -e "${YELLOW}[!] Git clone failed. Falling back to archive download...${NC}"
            rm -rf "$INSTALL_DIR"
            if ! DOWNLOAD_ARCHIVE; then exit 1; fi
        fi
        cd "$INSTALL_DIR"
    else
        echo -e "${YELLOW}[!] 'git' not found. Downloading archive...${NC}"
        if ! DOWNLOAD_ARCHIVE; then exit 1; fi
        cd "$INSTALL_DIR"
    fi
fi

# 4. Install Optional Web Admin Build Tools
USE_BUN=false
HAS_WEB_SOURCE=false
if [ -d "$INSTALL_DIR/web-admin" ]; then
    HAS_WEB_SOURCE=true
    echo -e "${BLUE}[*] Web Admin source detected. Preparing build tools...${NC}"
    BITS=$(getconf LONG_BIT)
    if [[ "$(uname -m)" = "aarch64" && "$BITS" = "64" ]]; then
        if ! command -v bun >/dev/null 2>&1; then
            echo -e "${BLUE}[*] Installing Bun for Web Admin...${NC}"
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
            echo -e "${BLUE}[i] Installing Node.js and NPM...${NC}"
            curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
            sudo apt-get install -y nodejs
        fi
    fi
else
    echo -e "${BLUE}[i] Using pre-built Web Admin (Single Binary Mode).${NC}"
fi

# 5. Build Engine
echo -e "${BLUE}[*] Building MaruBot engine...${NC}"

if [ "$HAS_WEB_SOURCE" = true ]; then
    echo -e "${BLUE}    ${MSG_WEB_BUILD}${NC}"
    cd "$INSTALL_DIR/web-admin"
    if [ "$USE_BUN" = true ]; then
        "$HOME/.bun/bin/bun" install
        "$HOME/.bun/bin/bun" run build
    else
        npm install --legacy-peer-deps
        npm run build
    fi
    echo -e "${BLUE}    [i] Embedding Web Admin into Go source...${NC}"
    rm -rf "$INSTALL_DIR/cmd/marubot/dashboard/dist"
    mkdir -p "$INSTALL_DIR/cmd/marubot/dashboard/dist"
    cp -r dist/* "$INSTALL_DIR/cmd/marubot/dashboard/dist/"
    cd "$INSTALL_DIR"
fi

echo -e "${BLUE}    ${MSG_GO_BUILD}${NC}"
TOTAL_MEM_KB=$(grep MemTotal /proc/meminfo | awk '{print $2}' || echo "0")
EXTRA_GOFLAGS=""
if [ "$TOTAL_MEM_KB" -gt 0 ] && [ "$TOTAL_MEM_KB" -lt 1500000 ]; then
    echo -e "${YELLOW}[!] Low memory detected. Limiting build parallelism...${NC}"
    EXTRA_GOFLAGS="-p=1"
fi

# skip tidy to use vendor

BUILD_TMPDIR="$HOME/.marubot/tmp"
mkdir -p "$BUILD_TMPDIR"
export TMPDIR="$BUILD_TMPDIR"
make GO="$GO_CMD" GOFLAGS="-v $EXTRA_GOFLAGS" clean build
rm -rf "$BUILD_TMPDIR"
unset TMPDIR

# 6. Install System
echo -e "${BLUE}[*] Installing system and deploying resources...${NC}"

if [ -f "build/marubot" ]; then
    INSTALL_BIN_DIR="$HOME/.marubot/bin"
    mkdir -p "$INSTALL_BIN_DIR"
    rm -f "$INSTALL_BIN_DIR/marubot"
    cp build/marubot "$INSTALL_BIN_DIR/marubot"
    chmod +x "$INSTALL_BIN_DIR/marubot"
    
    if [ -f "/usr/local/bin/marubot" ]; then
        sudo rm -f "/usr/local/bin/marubot"
    fi
    
    if [[ ":$PATH:" != *":$INSTALL_BIN_DIR:"* ]]; then
        case "$SHELL" in
            */zsh) echo "export PATH=\"$INSTALL_BIN_DIR:\$PATH\"" >> ~/.zshrc ;;
            *) echo "export PATH=\"$INSTALL_BIN_DIR:\$PATH\"" >> ~/.bashrc ;;
        esac
        export PATH="$INSTALL_BIN_DIR:$PATH"
    fi
else
    echo -e "${RED}[x] Build failed.${NC}"
    exit 1
fi

RESOURCE_DIR="$HOME/.marubot"
mkdir -p "$RESOURCE_DIR/config"
if [ ! -f "$RESOURCE_DIR/config.json" ]; then
    cp config/maru-config.json.example "$RESOURCE_DIR/config.json"
fi

if [ -f "$RESOURCE_DIR/config.json" ]; then
    if grep -q "\"language\":" "$RESOURCE_DIR/config.json"; then
        sed -i "s/\"language\": \".*\"/\"language\": \"$MARUBOT_LANG\"/" "$RESOURCE_DIR/config.json"
    else
        sed -i "0,/{/s/{/{\n  \"language\": \"$MARUBOT_LANG\",/" "$RESOURCE_DIR/config.json"
    fi

    if grep -q "\"admin_password\":" "$RESOURCE_DIR/config.json"; then
        sed -i "s/\"admin_password\": \".*\"/\"admin_password\": \"$MARUBOT_PWD\"/" "$RESOURCE_DIR/config.json"
    else
        sed -i "0,/{/s/{/{\n  \"admin_password\": \"$MARUBOT_PWD\",/" "$RESOURCE_DIR/config.json"
    fi
fi

rm -rf "$RESOURCE_DIR/skills" "$RESOURCE_DIR/tools"
cp -r skills "$RESOURCE_DIR/"
if [ -d "tools" ]; then cp -r tools "$RESOURCE_DIR/"; fi

if [ "$IS_PI" = true ] && [ -f "./maru-setup.sh" ]; then
    chmod +x maru-setup.sh
    ./maru-setup.sh
fi

echo -e "\n${GREEN}${MSG_SUCCESS}${NC}"
echo -e "[*] Cleaning up source folder..."
cd "$HOME"
rm -rf "$INSTALL_DIR"

echo -e "Command: ${BLUE}marubot agent${NC}"
echo -e "${MSG_DASHBOARD}: ${BLUE}marubot start${NC}"
