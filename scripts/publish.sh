#!/bin/bash

# MaruBot Public Sync Script
# 비공개 소스(maruminibot)에서 공개용 파일만 추출하여 ../maru-bot 폴더로 동기화합니다.
# Web Admin은 소스 코드 대신 빌드된 결과물(dist)만 포함합니다.

set -e

SOURCE_DIR=$(pwd)
TARGET_DIR=$(readlink -f "../maru-bot")

echo "🚀 공개 배포용 파일을 $TARGET_DIR 로 동기화합니다..."

# 1. 대상 디렉토리 생성
mkdir -p "$TARGET_DIR"

# 2. 대상 디렉토리 정리 ( .git 보존 )
echo "🧹 대상 폴더를 정리 중..."
if [ -d "$TARGET_DIR" ]; then
    find "$TARGET_DIR" -mindepth 1 -maxdepth 1 -not -name ".git" -exec rm -rf {} +
else
    mkdir -p "$TARGET_DIR"
fi

# 3. 선별적 파일 복사 (소스 코드)
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
)

echo "📂 파일 복사 중..."
for item in "${ITEMS[@]}"; do
    if [ -e "$SOURCE_DIR/$item" ]; then
        cp -R "$SOURCE_DIR/$item" "$TARGET_DIR/"
        echo "  ✓ $item 복사 완료"
    else
        echo "  ⚠️ $item 을 찾을 수 없어 건너뜜"
    fi
done

# 4. Web Admin 빌드 결과물 동기화
# Web Admin 소스 코드는 비공개로 유지하고, 빌드된 정적 자산만 Go 바이너리에 포함되도록 전파합니다.
echo "🏗️ Web Admin 빌드 결과물(dist)을 Go 대시보드 경로로 복사 중..."
if [ -d "$SOURCE_DIR/web-admin/dist" ]; then
    # Go embed 타겟 경로
    DEST_DIST="$TARGET_DIR/cmd/marubot/dashboard/dist"
    mkdir -p "$DEST_DIST"
    cp -r "$SOURCE_DIR/web-admin/dist/"* "$DEST_DIST/"
    echo "  ✓ Web Admin 빌드 자산 복사 완료 (Path: $DEST_DIST)"
else
    echo "  ❌ Web Admin 빌드 결과(dist)를 찾을 수 없습니다. 먼저 build를 수행하세요."
    exit 1
fi

# 5. README 파일 재정리 (공개 레포용: EN이 메인)
echo "📝 README 다국어 정리 중..."
cp "$SOURCE_DIR/README-en.md" "$TARGET_DIR/README.md"
cp "$SOURCE_DIR/README.md" "$TARGET_DIR/README-kor.md"

# 6. 민감 정보 제거 (설정 파일 등)
if [ -f "$TARGET_DIR/config/usersetting.json" ]; then
    rm "$TARGET_DIR/config/usersetting.json"
    echo "  🔒 usersetting.json (비공개 설정) 제거 완료"
fi

# 7. 명칭 최종 체크 및 치환
echo "🔄 명칭 최종 확인 중 (maruminibot -> marubot)..."
cd "$TARGET_DIR"
find . -type f -not -path '*/.*' -not -path '*/node_modules/*' -exec sed -i 's/maruminibot/marubot/g' {} + || true
find . -type f -not -path '*/.*' -not -path '*/node_modules/*' -exec sed -i 's/MaruMiniBot/MaruBot/g' {} + || true

cd "$SOURCE_DIR"
echo -e "\n✅ 모든 동기화가 완료되었습니다!"
echo "📍 위치: $TARGET_DIR"
