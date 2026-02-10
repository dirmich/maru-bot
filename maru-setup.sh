#!/bin/bash

# MaruMiniBot RPi Hardware Setup Script
# Version: 1.0.0

echo "🚀 MaruMiniBot 설정을 시작합니다..."

# 1. MaruMiniBot 엔진 확인
if command -v maruminibot > /dev/null; then
    echo "✅ MaruMiniBot 엔진이 감지되었습니다."
else
    echo "❌ MaruMiniBot 엔진을 찾을 수 없습니다. MaruMiniBot 설치 후 다시 시도해주세요."
    exit 1
fi

# 2. 하드웨어 접근 권한 설정 (라즈베리 파이 전용)
echo "📦 하드웨어 접근 권한을 설정합니다..."
# GPIO 사용 권한 추가
sudo usermod -aG gpio $USER 2>/dev/null
# I2C/SPI 인터페이스 활성화 가이드
echo "ℹ️ I2C 및 SPI 인터페이스가 활성화되어 있는지 raspi-config에서 확인하세요."

# 3. 필수 도구 설치 확인
echo "🛠️ 필수 멀티미디어 도구를 확인합니다..."
for tool in libcamera-apps alsa-utils; do
    if dpkg -s $tool > /dev/null 2>&1; then
        echo "✅ $tool 이 설치되어 있습니다."
    else
        echo "⚠️ $tool 이 없습니다. 설치를 권장합니다: sudo apt install $tool"
    fi
done

# 4. 설정 파일 연결
echo "📝 MaruMiniBot 설정을 MaruMiniBot에 적용합니다..."
mkdir -p ~/.maruminibot
cp ./config/maru-config.json ~/.maruminibot/config.json
echo "✅ 설정 완료! 이제 'maruminibot agent' 또는 'maru-run.sh'로 에드워드와 소통하세요."

echo "🎉 MaruMiniBot 준비 완료!"
