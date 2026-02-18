#!/bin/bash

# MaruBot One-Line Installer for Raspberry Pi
# Usage: curl -fsSL https://raw.githubusercontent.com/dirmich/maru-bot/main/install.sh | bash
set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m'

echo -e "${BLUE}🤖 MaruBot 원클릭 설치를 시작합니다...${NC}"

# 1. 아키텍처 및 OS 확인
if [[ "$(uname -m)" != "aarch64" && "$(uname -m)" != "armv7l" ]]; then
    echo -e "${RED}❌ 이 스크립트는 Raspberry Pi (ARM) 환경 전용입니다.${NC}"
    exit 1
fi

# 2. 필수 패키지 설치
echo -e "${BLUE}📦 필수 패키지를 설치합니다...${NC}"
sudo apt update
sudo apt install -y git make libcamera-apps alsa-utils vlc-plugin-base curl wget

# Go 설치 (1.24+)
GO_REQUIRED="1.24"
INSTALL_GO=false

if command -v go >/dev/null 2>&1; then
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//' | cut -d' ' -f1)
    if [ "$(printf '%s\n' "$GO_REQUIRED" "$GO_VERSION" | sort -V | head -n1)" != "$GO_REQUIRED" ]; then
        INSTALL_GO=true
    fi
else
    INSTALL_GO=true
fi

if [ "$INSTALL_GO" = true ]; then
    echo -e "${BLUE}Hamster🐹 Go $GO_REQUIRED+ 최신 버전을 설치합니다...${NC}"
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

# 51. Bun 또는 Node.js 설치 (Web Admin용)
# Bun은 ARM64만 지원하므로, ARMv7 등에서는 Node.js를 사용해야 함
USE_BUN=false
if [ "$(uname -m)" = "aarch64" ]; then
    if ! command -v bun >/dev/null 2>&1; then
        echo -e "${BLUE}🍞 Web Admin 실행을 위해 Bun을 설치합니다...${NC}"
        curl -fsSL https://bun.sh/install | bash
        export BUN_INSTALL="$HOME/.bun"
        export PATH="$BUN_INSTALL/bin:$PATH"
    fi
    
    # 설치 확인
    if [ -f "$HOME/.bun/bin/bun" ]; then
        USE_BUN=true
    else
        echo -e "${RED}⚠️ Bun 설치에 실패했습니다. Node.js로 전환합니다.${NC}"
    fi
else
    echo -e "${BLUE}ℹ️ 32-bit 환경(또는 비-ARM64)이 감지되었습니다. Bun 대신 Node.js를 사용합니다.${NC}"
fi

# Bun을 사용할 수 없는 경우 Node.js 설치
if [ "$USE_BUN" = false ]; then
    if ! command -v node >/dev/null 2>&1; then
        echo -e "${BLUE}📦 Node.js 및 NPM을 설치합니다...${NC}"
        curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
        sudo apt-get install -y nodejs
    fi
fi

# 3. 소스 코드 클론
INSTALL_DIR="$HOME/marubot"
if [ -d "$INSTALL_DIR" ]; then
    echo -e "${BLUE}🔄 최신 코드로 업데이트합니다...${NC}"
    cd "$INSTALL_DIR"
    git pull
else
    echo -e "${BLUE}📂 GitHub에서 MaruBot 소스를 가져옵니다...${NC}"
    git clone --depth 1 https://github.com/dirmich/maru-bot.git "$INSTALL_DIR"
    cd "$INSTALL_DIR"
fi

# 4. 엔지 빌드
echo -e "${BLUE}🛠️ MaruBot 엔진을 빌드합니다...${NC}"
go mod tidy
make build

# 5. 시스템 설치 및 리소스 배치
echo -e "${BLUE}🏗️ 시스템에 설치 및 리소스를 배치합니다...${NC}"

# 5-1. 실행 파일 설치 (시스템 전역)
if [ -f "build/marubot" ]; then
    echo "  📦 실행 파일(/usr/local/bin/marubot) 복사 중..."
    sudo cp build/marubot /usr/local/bin/
    sudo chmod +x /usr/local/bin/marubot
else
    echo -e "${RED}❌ 빌드된 marubot 실행 파일이 없습니다. 빌드 실패.${NC}"
    exit 1
fi

# 5-2. 리소스 디렉토리 구성 (~/.marubot)
RESOURCE_DIR="$HOME/.marubot"
mkdir -p "$RESOURCE_DIR"

echo "  📂 리소스(~/.marubot) 설정 중..."

# (1) Config
mkdir -p "$RESOURCE_DIR/config"
# 기존 설정 유지 (없을 때만 복사)
if [ ! -f "$RESOURCE_DIR/config.json" ]; then
    cp config/maru-config.json "$RESOURCE_DIR/config.json"
fi

# (2) Skills, Tools
# 기존 폴더 제거 후 최신 복사 (업데이트)
rm -rf "$RESOURCE_DIR/skills" "$RESOURCE_DIR/tools"
cp -r skills "$RESOURCE_DIR/"
if [ -d "tools" ]; then cp -r tools "$RESOURCE_DIR/"; fi

# (3) Web Admin
# 기존 web-admin 제거 (clean install)
rm -rf "$RESOURCE_DIR/web-admin"
if [ -d "web-admin" ]; then
    echo "  🌐 Web Admin 리소스 복사..."
    cp -r web-admin "$RESOURCE_DIR/"
    
    # 의존성 설치 (이동된 위치에서 수행)
    cd "$RESOURCE_DIR/web-admin"
    if [ "$USE_BUN" = true ]; then
        echo -e "${BLUE}    🍞 Bun으로 런타임 의존성 설치...${NC}"
        $HOME/.bun/bin/bun install --production
    else
        echo -e "${BLUE}    📦 NPM으로 런타임 의존성 설치...${NC}"
        npm install --production
    fi
    cd "$INSTALL_DIR"
fi

# 6. 하드웨어 설정 스크립트 실행
chmod +x maru-setup.sh
# maru-setup.sh 내부에서 'marubot' 명령어를 사용하므로 PATH 등록 없이 바로 실행 가능해야 함 (/usr/local/bin)
./maru-setup.sh

# 7. PATH 등록 (Bun만 필요, MaruBot은 이미 /usr/local/bin)
if [ "$USE_BUN" = true ]; then
    if ! grep -q "BUN_INSTALL" ~/.bashrc; then
        echo "export BUN_INSTALL=\"\$HOME/.bun\"" >> ~/.bashrc
        echo "export PATH=\"\$BUN_INSTALL/bin:\$PATH\"" >> ~/.bashrc
    fi
fi

# 레거시 PATH 제거 (혹시 이전에 설치했다면)
# sed -i '/marubot\/build/d' ~/.bashrc  <-- 위험할 수 있으므로 사용자에게 맡김

echo -e "\n${GREEN}🎉 MaruBot 설치가 완료되었습니다!${NC}"
echo -e "🧹 설치에 사용된 소스 폴더($INSTALL_DIR)를 자동으로 정리합니다..."
cd "$HOME"
rm -rf "$INSTALL_DIR"

echo -e "명령어: ${BLUE}marubot agent${NC} (콘솔 채팅)"
echo -e "대시보드: ${BLUE}marubot dashboard${NC} (웹 관리자)"
