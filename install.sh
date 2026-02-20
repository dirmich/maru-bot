#!/bin/bash

# MaruBot One-Line Installer for Raspberry Pi
# Usage: curl -fsSL https://raw.githubusercontent.com/dirmich/maru-bot/main/install.sh | bash
set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}🤖 Starting MaruBot One-Click Installer...${NC}"

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

# 0-1. Admin Password Selection
if [ "$MARUBOT_LANG" = "ko" ]; then
    PROMPT_PWD="웹 대시보드 관리자 암호를 설정하세요 [기본값: admin]: "
elif [ "$MARUBOT_LANG" = "ja" ]; then
    PROMPT_PWD="Webダッシュボードの管理パスワードを設定してください [デフォルト: admin]: "
else
    PROMPT_PWD="Set Admin Password for Web Dashboard [Default: admin]: "
fi

if [ -c /dev/tty ]; then
    read -p "$PROMPT_PWD" MARUBOT_PWD < /dev/tty
else
    if [ -t 0 ]; then
        read -p "$PROMPT_PWD" MARUBOT_PWD
    else
        MARUBOT_PWD="admin"
    fi
fi

if [ -z "$MARUBOT_PWD" ]; then
    MARUBOT_PWD="admin"
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
    MSG_DASHBOARD="대시보드 실행: marubot dashboard"
elif [ "$MARUBOT_LANG" = "ja" ]; then
    MSG_ARCH_ERR="❌ このスクリプトはRaspberry Pi(ARM)環境専用です。"
    MSG_PKG_INST="📦 必須パッケージをインストール中..."
    MSG_GO_INST="🐹 最新のGoをインストール中..."
    MSG_CLONE="📂 MaruBotソースコードをクローン中..."
    MSG_WEB_BUILD="🏗️ Web管理画面(Vite)をビルド中..."
    MSG_GO_BUILD="🛠️ MaruBotエンジンをビルド中..."
    MSG_SUCCESS="🎉 MaruBotのインストールが完了しました！"
    MSG_DASHBOARD="ダッシュボードの実行: marubot dashboard"
else
    MSG_ARCH_ERR="❌ This script is only for Raspberry Pi (ARM) environments."
    MSG_PKG_INST="📦 Installing required packages..."
    MSG_GO_INST="🐹 Installing latest Go..."
    MSG_CLONE="📂 Cloning MaruBot source from GitHub..."
    MSG_WEB_BUILD="🏗️ Building Web Admin (Vite)..."
    MSG_GO_BUILD="🛠️ Building MaruBot engine..."
    MSG_SUCCESS="🎉 MaruBot installation complete!"
    MSG_DASHBOARD="Run dashboard: marubot dashboard"
fi

# 1. Check Architecture and OS
if [[ "$(uname -m)" != "aarch64" && "$(uname -m)" != "armv7l" ]]; then
    echo -e "${RED}${MSG_ARCH_ERR}${NC}"
    exit 1
fi

# 2. Install Required Packages
echo -e "${BLUE}${MSG_PKG_INST}${NC}"
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
    echo -e "${BLUE}${MSG_CLONE}${NC}"
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
make GO="$GO_CMD" build

# 6. Install System and Deploy Resources
echo -e "${BLUE}🏗️ Installing system and deploying resources...${NC}"

if [ -f "build/marubot" ]; then
    echo "  📦 Copying executable to /usr/local/bin/marubot..."
    sudo rm -f /usr/local/bin/marubot
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

echo -e "\n${GREEN}${MSG_SUCCESS}${NC}"
echo -e "🧹 Automatically cleaning up the source folder ($INSTALL_DIR)..."
cd "$HOME"
rm -rf "$INSTALL_DIR"

echo -e "Command: ${BLUE}marubot agent${NC}"
echo -e "${MSG_DASHBOARD}: ${BLUE}marubot dashboard${NC}"
