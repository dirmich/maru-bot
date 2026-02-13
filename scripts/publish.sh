#!/bin/bash

# MaruBot Public Sync Script
# 비공개 소스(maruminibot -> 현재 marubot으로 명칭 변경됨)에서 공개용 파일만 추출하여 ../marubot 폴더로 동기화합니다.

set -e

SOURCE_DIR=$(pwd)
TARGET_DIR="../maru-bot"

echo "🚀 공개 배포용 파일을 $TARGET_DIR 로 동기화합니다..."

# 1. 대상 디렉토리 생성
mkdir -p "$TARGET_DIR"

# 2. 대상 디택토리 정리 (.git 및 필요한 파일 제외)
echo "🧹 대상 폴더를 정리 중..."
if [ -d "$TARGET_DIR/.git" ]; then
    # .git 폴더와 기존에 있을지 모를 maru-bot 관련 메타파일 제외하고 삭제
    find "$TARGET_DIR" -maxdepth 1 ! -name ".git" ! -name ".." ! -name "." -exec rm -rf {} +
else
    rm -rf "${TARGET_DIR:?}"/*
fi

# 3. 선별적 파일 복사
ITEMS=(
    "cmd"
    "pkg"
    "config"
    "skills"
    "Makefile"
    "go.mod"
    "go.sum"
    "README.md"
    "README-en.md"
    "README-ja.md"
    "README-cn.md"
    ".gitignore"
    "install.sh"
    "maru-setup.sh"
    "LICENSE"
    "web-admin"
)

echo "📂 파일 복사 중..."
for item in "${ITEMS[@]}"; do
    if [ -e "$SOURCE_DIR/$item" ]; then
        if [ "$item" == "web-admin" ]; then
            echo "  📦 web-admin 소스 복사 중..."
            mkdir -p "$TARGET_DIR/web-admin"
            # 무거운 폴더 제외하고 복사 (이미 빌드와 상관없이 소스만 배포)
            tar -c --exclude='.git' --exclude='node_modules' --exclude='.next' --exclude='.env*' --exclude='*.db*' -C "$SOURCE_DIR/web-admin" . | tar -x -C "$TARGET_DIR/web-admin"
            echo "  ✓ web-admin 복사 완료"
        else
            cp -R "$SOURCE_DIR/$item" "$TARGET_DIR/"
            echo "  ✓ $item 복사 완료"
        fi
    else
        echo "  ⚠️ $item 을 찾을 수 없어 건너뜜"
    fi
done

# README 파일 재정리 (공개 레포용: EN이 메인)
echo "📝 README 다국어 정리 중..."
cp "$SOURCE_DIR/README-en.md" "$TARGET_DIR/README.md"
cp "$SOURCE_DIR/README.md" "$TARGET_DIR/README-kor.md"

# 4. 민감 정보 제거 (설정 파일 등)
if [ -f "$TARGET_DIR/config/usersetting.json" ]; then
    rm "$TARGET_DIR/config/usersetting.json"
    echo "  🔒 usersetting.json (비공개 설정) 제거 완료"
fi

# 5. 명칭 최종 체크 및 치환 (혹시 남았을지 모를 maruminibot -> marubot)
echo "🔄 명칭 최종 확인 중 (maruminibot -> marubot)..."
cd "$TARGET_DIR"

# 파일 내용 치환
find . -type f -not -path '*/.*' -not -path '*/node_modules/*' -exec sed -i 's/maruminibot/marubot/g' {} + || true
find . -type f -not -path '*/.*' -not -path '*/node_modules/*' -exec sed -i 's/MaruMiniBot/MaruBot/g' {} + || true
find . -type f -not -path '*/.*' -not -path '*/node_modules/*' -exec sed -i 's/MARUMINIBOT/MARUBOT/g' {} + || true

# GitHub 레포지토리 주소 조정 (정식 명칭 maru-bot)
echo "🌐 GitHub 레포지토리 주소 조정 (dirmich/maru-bot)..."
find . -type f -not -path '*/.*' -exec sed -i 's/dirmich\/marubot/dirmich\/maru-bot/g' {} + || true
find . -type f -not -path '*/.*' -exec sed -i 's/maru-ai\/maru-bot/dirmich\/maru-bot/g' {} + || true

cd "$SOURCE_DIR"
echo -e "\n✅ 모든 동기화가 완료되었습니다!"
echo "📍 위치: $TARGET_DIR"
