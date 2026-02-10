#!/bin/bash

# MaruMiniBot One-Line Installer for Raspberry Pi
# Usage: curl -fsSL https://raw.githubusercontent.com/maru-ai/maruminibot/main/install.sh | bash

set -e

RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

echo -e "${BLUE}🤖 MaruMiniBot 원클릭 설치를 시작합니다...${NC}"

# 1. 아키텍처 및 OS 확인
if [[ "$(uname -m)" != "aarch64" && "$(uname -m)" != "armv7l" ]]; then
    echo -e "${RED}❌ 이 스크립트는 Raspberry Pi (ARM) 환경 전용입니다.${NC}"
    exit 1
fi

# 2. 시스템 업데이트 및 필수 패키지 설치
echo -e "${BLUE}📦 시스템 업데이트 및 필수 패키지를 설치합니다...${NC}"
sudo apt update
sudo apt install -y git make golang libcamera-apps alsa-utils vlc-plugin-base

# 3. 소스 코드 클론
INSTALL_DIR="$HOME/maruminibot"
if [ -d "$INSTALL_DIR" ]; then
    echo -e "${BLUE}🔄 기존 설치 폴더가 발견되어 업데이트를 진행합니다...${NC}"
    cd "$INSTALL_DIR"
    git pull
else
    echo -e "${BLUE}📂 GitHub에서 소스 코드를 가져옵니다...${NC}"
    git clone https://github.com/maru-ai/maruminibot.git "$INSTALL_DIR"
    cd "$INSTALL_DIR"
fi

# 4. 바이너리 빌드
echo -e "${BLUE}🛠️ MaruMiniBot 엔진을 빌드합니다...${NC}"
make build

# 5. 실행 권한 부여 및 시스템 경로 등록
chmod +x build/maruminibot
chmod +x maru-setup.sh

# 6. 하드웨어 설정 스크립트 실행
echo -e "${BLUE}⚙️ 하드웨어 초기 설정을 시작합니다...${NC}"
./maru-setup.sh

# 7. 환경 변수 등록 (.bashrc)
if ! grep -q "maruminibot" ~/.bashrc; then
    echo 'export PATH="$HOME/maruminibot/build:$PATH"' >> ~/.bashrc
    echo -e "${GREEN}✅ PATH에 maruminibot이 등록되었습니다. (새 터미널에서 적용)${NC}"
fi

echo -e "\n${GREEN}🎉 MaruMiniBot 설치가 완료되었습니다!${NC}"
echo -e "명령어: ${BLUE}maruminibot agent${NC} 를 입력하여 AI 에이전트를 실행하세요."
echo -e "설정 파일 위치: ${BLUE}~/.maruminibot/config.json${NC}"
