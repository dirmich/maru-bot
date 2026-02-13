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

# 5. Web Admin 설정
if [ -d "web-admin" ]; then
    echo -e "${BLUE}🌐 Web Admin 디렉토리를 초기화합니다...${NC}"
    cd web-admin
    
    if [ "$USE_BUN" = true ]; then
        echo -e "${BLUE}🍞 Bun으로 의존성 설치 및 빌드...${NC}"
        $HOME/.bun/bin/bun install
        echo -e "${BLUE}🏗️ Web Admin 빌드 중 (Bun)...${NC}"
        $HOME/.bun/bin/bun run build
    else
        echo -e "${BLUE}📦 NPM으로 의존성 설치 및 빌드...${NC}"
        npm install
        echo -e "${BLUE}🏗️ Web Admin 빌드 중 (NPM)...${NC}"
        npm run build
    fi
    cd ..
fi

# 6. 하드웨어 설정 스크립트 실행
chmod +x maru-setup.sh
export PATH="$PWD/build:$PATH"
./maru-setup.sh

# 7. PATH 등록
if ! grep -q "marubot/build" ~/.bashrc; then
    echo "export PATH=\"\$HOME/marubot/build:\$PATH\"" >> ~/.bashrc
    echo "export BUN_INSTALL=\"\$HOME/.bun\"" >> ~/.bashrc
    echo "export PATH=\"\$BUN_INSTALL/bin:\$PATH\"" >> ~/.bashrc
    echo -e "${GREEN}✅ PATH 등록 완료 (명령어: marubot)${NC}"
fi

echo -e "\n${GREEN}🎉 MaruBot 설치가 완료되었습니다!${NC}"
echo -e "명령어: ${BLUE}marubot agent${NC} (콘솔 채팅)"
echo -e "대시보드: ${BLUE}marubot dashboard${NC} (웹 관리자)"
