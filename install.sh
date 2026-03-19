#!/bin/bash

# MaruBot One-Line Installer for Raspberry Pi
# Usage: curl -fsSL https://raw.githubusercontent.com/dirmich/maru-bot/main/install.sh | bash
set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

# 0. Parse Arguments
FORCE_RPI=false
for arg in "$@"; do
    if [ "$arg" == "--rpi" ]; then
        FORCE_RPI=true
    fi
done

echo -e "${BLUE}🤖 Starting MaruBot One-Click Installer...${NC}"
if [ "$FORCE_RPI" = true ]; then
    echo -e "${BLUE}ℹ️ Raspberry Pi mode forced via parameter.${NC}"
fi
# 0. Language Selection
# 0. Language Selection
echo "Language / 언어 / 言語:"
echo "1) English (en)"
echo "2) 한국어 (ko)"
echo "3) 日本語 (ja)"

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

if [ "$MARUBOT_LANG" = "ko" ]; then
    PROMPT_PWD="웹 대시보드 관리자 암호를 설정하세요$PROMPT_PWD_SUFFIX: "
elif [ "$MARUBOT_LANG" = "ja" ]; then
    PROMPT_PWD="Webダッシュボードの管理パスワードを設定してください$PROMPT_PWD_SUFFIX: "
else
    PROMPT_PWD="Set Admin Password for Web Dashboard$PROMPT_PWD_SUFFIX: "
fi

if [ -c /dev/tty ]; then
    read -p "$PROMPT_PWD" MARUBOT_PWD < /dev/tty
else
    if [ -t 0 ]; then
        read -p "$PROMPT_PWD" MARUBOT_PWD
    else
        MARUBOT_PWD="" # Will check later
    fi
fi

if [ -z "$MARUBOT_PWD" ]; then
    if [ ! -z "$EXISTING_PWD" ]; then
        MARUBOT_PWD="$EXISTING_PWD"
        echo "Keeping existing password."
    else
        MARUBOT_PWD="admin"
        echo "Defaulting to 'admin'."
    fi
fi

# Translations
if [ "$MARUBOT_LANG" = "ko" ]; then
    MSG_ARCH_ERR="❌ 이 스크립트는 라즈베리 파이(ARM) 환경 전용입니다."
    MSG_PKG_INST="📦 필수 패키지 설치 중..."
    MSG_GO_INST="🐹 최신 Go 설치 중..."
    MSG_CLONE="📂 MaruBot 소스 코드 클론 중..."
    MSG_WEB_BUILD="🏗️ 웹 관리자 페이지(Vite) 빌드 중..."
    MSG_GO_BUILD="🛠️ MaruBot 엔진 빌드 중..."
    MSG_SUCCESS="🎉 MaruBot 설치가 완료되었습니다!"
    MSG_DASHBOARD="대시보드 실행: marubot start"
    MSG_UPGRADE="🚀 MaruBot 업그레이드 시도 중..."
elif [ "$MARUBOT_LANG" = "ja" ]; then
    MSG_ARCH_ERR="❌ このスクリプトはRaspberry Pi(ARM)環境専用です。"
    MSG_PKG_INST="📦 必須パッケージをインストール中..."
    MSG_GO_INST="🐹 最新のGoをインストール中..."
    MSG_CLONE="📂 MaruBotソースコードをクローン中..."
    MSG_WEB_BUILD="🏗️ Web管理画面(Vite)をビルド中..."
    MSG_GO_BUILD="🛠️ MaruBotエンジンをビルド中..."
    MSG_SUCCESS="🎉 MaruBotのインストールが完了しました！"
    MSG_DASHBOARD="ダッシュボードの実行: marubot start"
    MSG_UPGRADE="🚀 MaruBotアップグレードを試行中..."
else
    MSG_ARCH_ERR="❌ This script is only for Raspberry Pi (ARM) environments."
    MSG_PKG_INST="📦 Installing required packages..."
    MSG_GO_INST="🐹 Installing latest Go..."
    MSG_CLONE="📂 Cloning MaruBot source from GitHub..."
    MSG_WEB_BUILD="🏗️ Building Web Admin (Vite)..."
    MSG_GO_BUILD="🛠️ Building MaruBot engine..."
    MSG_SUCCESS="🎉 MaruBot installation complete!"
    MSG_DASHBOARD="Run dashboard: marubot start"
    MSG_UPGRADE="🚀 Attempting MaruBot upgrade..."
fi

# 1. Check Architecture and OS
IS_PI=false
if [[ "$(uname -m)" == "aarch64" || "$(uname -m)" == "armv7l" ]]; then
    IS_PI=true
fi

if [ "$IS_PI" = false ]; then
    echo -e "${YELLOW}⚠️ Notice: This environment is not Raspberry Pi (ARM). Hardware-specific features (GPIO, etc.) will be disabled.${NC}"
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
    echo -e "${YELLOW}⚠️ 'apt' not found. Skipping system package installation. Ensure git, make, curl, and wget are installed manually.${NC}"
fi

# Install Go (1.24+)
GO_REQUIRED="1.24"
INSTALL_GO=false

# Check for go in common locations
if command -v go >/dev/null 2>&1; then
    DETECTED_GO=$(command -v go)
elif [ -f "/usr/local/go/bin/go" ]; then
    DETECTED_GO="/usr/local/go/bin/go"
fi

if [ ! -z "$DETECTED_GO" ]; then
    if ! "$DETECTED_GO" version >/dev/null 2>&1; then
        echo -e "${YELLOW}⚠️ Existing Go installation is broken (e.g., Exec format error). Forcing reinstall...${NC}"
        INSTALL_GO=true
    else
        EXISTING_VERSION=$("$DETECTED_GO" version | awk '{print $3}' | sed 's/go//')
        # More robust version comparison
        if [[ "$EXISTING_VERSION" == "$GO_REQUIRED"* ]] || [ "$(printf '%s\n%s' "$GO_REQUIRED" "$EXISTING_VERSION" | sort -V | head -n1)" = "$GO_REQUIRED" ]; then
            echo -e "${GREEN}✓ Go $EXISTING_VERSION is already installed at $DETECTED_GO.${NC}"
            INSTALL_GO=false
        else
            echo -e "${BLUE}ℹ️ Upgrading Go from $EXISTING_VERSION to $GO_REQUIRED+...${NC}"
            INSTALL_GO=true
        fi
    fi
else
    INSTALL_GO=true
fi

if [ "$INSTALL_GO" = true ]; then
    echo -e "${BLUE}🐹 Installing latest Go $GO_REQUIRED+ ...${NC}"
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
        GO_ARCH="armv6l" # Default to armv6l on unknown for RPi safety, or we could leave as amd64 but print a warning
        echo -e "${YELLOW}⚠️ Unknown architecture ($ARCH). Defaulting to armv6l for Raspberry Pi.${NC}"
    fi
    WGET_URL="https://go.dev/dl/go1.24.0.linux-$GO_ARCH.tar.gz"
    echo "Downloading from $WGET_URL ..."
    wget -O go_dist.tar.gz "$WGET_URL"
    sudo rm -rf /usr/local/go && sudo tar -C /usr/local -xzf go_dist.tar.gz
    rm go_dist.tar.gz
    if ! grep -q "/usr/local/go/bin" ~/.bashrc; then echo 'export PATH=/usr/local/go/bin:$PATH' >> ~/.bashrc; fi
fi

# Ensure Go is in PATH for this script session
if [ -f "/usr/local/go/bin/go" ]; then
    export PATH=/usr/local/go/bin:$PATH
    GO_CMD="/usr/local/go/bin/go"
else
    GO_CMD=$(command -v go)
fi

if [ -z "$GO_CMD" ]; then
    echo -e "${RED}❌ Go not found even after installation attempt.${NC}"
    exit 1
fi

# 3. Clone Source Code
INSTALL_DIR="$HOME/marubot"
DOWNLOAD_ARCHIVE() {
    echo -e "${BLUE}ℹ️ Downloading archive via curl/wget...${NC}"
    mkdir -p "$INSTALL_DIR"
    if command -v curl >/dev/null 2>&1; then
        curl -fsSL https://github.com/dirmich/maru-bot/archive/refs/heads/main.tar.gz | tar -xz -C "$INSTALL_DIR" --strip-components=1
    elif command -v wget >/dev/null 2>&1; then
        wget -qO- https://github.com/dirmich/maru-bot/archive/refs/heads/main.tar.gz | tar -xz -C "$INSTALL_DIR" --strip-components=1
    else
        echo -e "${RED}❌ Neither curl nor wget are available.${NC}"
        return 1
    fi
}

if [ -d "$INSTALL_DIR" ]; then
    echo -e "${BLUE}🔄 Updating to latest source code...${NC}"
    cd "$INSTALL_DIR"
    if command -v git >/dev/null 2>&1 && [ -d ".git" ]; then
        if ! git pull; then
            echo -e "${YELLOW}⚠️ Git update failed. Retrying fresh download...${NC}"
            cd "$HOME"
            rm -rf "$INSTALL_DIR"
            if ! DOWNLOAD_ARCHIVE; then exit 1; fi
            cd "$INSTALL_DIR"
        fi
    else
        echo -e "${YELLOW}⚠️ Not a git repository or git missing. Downloading fresh archive...${NC}"
        cd "$HOME"
        rm -rf "$INSTALL_DIR"
        if ! DOWNLOAD_ARCHIVE; then exit 1; fi
        cd "$INSTALL_DIR"
    fi
else
    echo -e "${BLUE}${MSG_CLONE}${NC}"
    if command -v git >/dev/null 2>&1; then
        if ! git clone --depth 1 https://github.com/dirmich/maru-bot.git "$INSTALL_DIR"; then
            echo -e "${YELLOW}⚠️ Git clone failed. Falling back to archive download...${NC}"
            rm -rf "$INSTALL_DIR"
            if ! DOWNLOAD_ARCHIVE; then exit 1; fi
        fi
        cd "$INSTALL_DIR"
    else
        echo -e "${YELLOW}⚠️ 'git' not found. Downloading archive...${NC}"
        if ! DOWNLOAD_ARCHIVE; then exit 1; fi
        cd "$INSTALL_DIR"
    fi
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
    echo -e "${BLUE}    ${MSG_WEB_BUILD}${NC}"
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
echo -e "${BLUE}    ${MSG_GO_BUILD}${NC}"
$GO_CMD mod tidy
$GO_CMD clean -cache # Ensure new code is recompiled
make GO="$GO_CMD" clean build

# 6. Install System and Deploy Resources
echo -e "${BLUE}🏗️ Installing system and deploying resources...${NC}"

if [ -f "build/marubot" ]; then
    INSTALL_BIN_DIR="$HOME/.marubot/bin"
    mkdir -p "$INSTALL_BIN_DIR"
    echo "  📦 Installing executable to $INSTALL_BIN_DIR/marubot..."
    
    # Avoid 'Text file busy' by removing the existing binary first
    rm -f "$INSTALL_BIN_DIR/marubot"
    cp build/marubot "$INSTALL_BIN_DIR/marubot"
    chmod +x "$INSTALL_BIN_DIR/marubot"
    
    # Remove old global installation if it exists
    if [ -f "/usr/local/bin/marubot" ]; then
        echo "  🧹 Removing old global binary from /usr/local/bin..."
        sudo rm -f "/usr/local/bin/marubot"
    fi
    
    # Add to PATH if not already there
    if [[ ":$PATH:" != *":$INSTALL_BIN_DIR:"* ]]; then
        echo "  🌐 Adding $INSTALL_BIN_DIR to PATH..."
        case "$SHELL" in
            */zsh)
                echo "export PATH=\"$INSTALL_BIN_DIR:\$PATH\"" >> ~/.zshrc
                ;;
            *)
                echo "export PATH=\"$INSTALL_BIN_DIR:\$PATH\"" >> ~/.bashrc
                ;;
        esac
        export PATH="$INSTALL_BIN_DIR:$PATH"
    fi
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

# Set selected language in config
if [ -f "$RESOURCE_DIR/config.json" ]; then
    # Simple check and replace for "language" field
    if grep -q "\"language\":" "$RESOURCE_DIR/config.json"; then
        sed -i "s/\"language\": \".*\"/\"language\": \"$MARUBOT_LANG\"/" "$RESOURCE_DIR/config.json"
    else
        # Add after opening brace if not exists
        sed -i "0,/{/s/{/{\n  \"language\": \"$MARUBOT_LANG\",/" "$RESOURCE_DIR/config.json"
    fi

    # Set admin_password field
    if grep -q "\"admin_password\":" "$RESOURCE_DIR/config.json"; then
        sed -i "s/\"admin_password\": \".*\"/\"admin_password\": \"$MARUBOT_PWD\"/" "$RESOURCE_DIR/config.json"
    else
        # Add after opening brace
        sed -i "0,/{/s/{/{\n  \"admin_password\": \"$MARUBOT_PWD\",/" "$RESOURCE_DIR/config.json"
    fi

    # Set is_raspberry_pi field if forced
    if [ "$FORCE_RPI" = true ]; then
        if grep -q "\"is_raspberry_pi\":" "$RESOURCE_DIR/config.json"; then
            sed -i "s/\"is_raspberry_pi\": .*/\"is_raspberry_pi\": true,/" "$RESOURCE_DIR/config.json"
        else
            # Add after opening brace (under hardware section if possible, but simple add for now)
            # Find hardware section and add or just add at start
            sed -i "0,/{/s/{/{\n  \"hardware\": {\n    \"is_raspberry_pi\": true\n  },/" "$RESOURCE_DIR/config.json"
        fi
    fi
fi

rm -rf "$RESOURCE_DIR/skills" "$RESOURCE_DIR/tools"
cp -r skills "$RESOURCE_DIR/"
if [ -d "tools" ]; then cp -r tools "$RESOURCE_DIR/"; fi

# Clean up legacy
if [ -d "$RESOURCE_DIR/web-admin" ]; then
    rm -rf "$RESOURCE_DIR/web-admin"
fi

# 7. Hardware Setup (Only for Raspberry Pi)
if [ "$IS_PI" = true ] && [ -f "./maru-setup.sh" ]; then
    chmod +x maru-setup.sh
    ./maru-setup.sh
fi

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

echo -e "\n${GREEN}${MSG_SUCCESS}${NC}"
echo -e "🧹 Automatically cleaning up the source folder ($INSTALL_DIR)..."
cd "$HOME"
rm -rf "$INSTALL_DIR"

echo -e "Command: ${BLUE}marubot agent${NC}"
echo -e "${MSG_DASHBOARD}: ${BLUE}marubot start${NC}"
