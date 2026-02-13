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
if ! command -v go >/dev/null 2>&1; then
    echo -e "${BLUE}🐹 Go 1.24+ 최신 버전을 설치합니다...${NC}"
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

# Bun 설치 (Web Admin용)
if ! command -v bun >/dev/null 2>&1; then
    echo -e "${BLUE}🍞 Web Admin 실행을 위해 Bun을 설치합니다...${NC}"
    curl -fsSL https://bun.sh/install | bash
    export BUN_INSTALL="$HOME/.bun"
    export PATH="$BUN_INSTALL/bin:$PATH"
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
make build

# 5. Web Admin 설정
if [ -d "web-admin" ]; then
    echo -e "${BLUE}🌐 Web Admin 디렉토리를 초기화합니다...${NC}"
    cd web-admin
    bun install
    # 빌드는 첫 실행 시 자동으로 수행되거나 개발자가 직접 할 수 있도록 유지
    # 사용자 편의를 위해 여기서 빌드도 수행 (처음엔 좀 걸림)
    echo -e "${BLUE}🏗️ Web Admin 빌드 중 (최초 1회, 수 분 소요)...${NC}"
    bun run build
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
