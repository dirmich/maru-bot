#!/bin/bash

# MaruMiniBot Public Sync Script
# 비공개 소스에서 공개용 파일만 추출하여 ../marubot 폴더로 동기화합니다.

set -e

SOURCE_DIR=$(pwd)
TARGET_DIR="../marubot"

echo "🚀 공개 배포용 파일을 $TARGET_DIR 로 동기화합니다..."

# 1. 대상 디렉토리 생성
mkdir -p "$TARGET_DIR"

# 2. 대상 디렉토리 정리 (.git 제외)
echo "🧹 대상 폴더를 정리 중..."
if [ -d "$TARGET_DIR/.git" ]; then
    # .git 폴더를 제외한 모든 파일/폴더 삭제
    find "$TARGET_DIR" -maxdepth 1 ! -name ".git" ! -name "marubot" ! -name ".." ! -name "." -exec rm -rf {} +
else
    rm -rf "${TARGET_DIR:?}"/*
fi

# 3. 선별적 파일 복사
# 공개할 항목 리스트
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
        echo "  ⚠️ $item 을 찾을 수 없어 건너뜁니다."
    fi
done

# README 파일 재정리 (공개 레포용)
echo "📝 README 다국어 정리 중..."
cp "$SOURCE_DIR/README-en.md" "$TARGET_DIR/README.md"
cp "$SOURCE_DIR/README.md" "$TARGET_DIR/README-kor.md"
# README-ja, README-cn은 이미 ITEMS에 포함되어 복사됨

# 4. 민감 정보 제거
if [ -f "$TARGET_DIR/config/usersetting.json" ]; then
    rm "$TARGET_DIR/config/usersetting.json"
    echo "  🔒 usersetting.json (비공개 설정) 제거 완료"
fi

# 5. 명칭 치환 (maruminibot -> marubot)
echo "🔄 프로젝트 명칭 치환 중 (maruminibot -> marubot)..."
cd "$TARGET_DIR"

# 파일 내용 치환 (Go 모듈명, 임포트 경로, 문서 텍스트 등)
find . -type f -not -path '*/.*' -exec sed -i 's/maruminibot/marubot/g' {} + || true
find . -type f -not -path '*/.*' -exec sed -i 's/MaruMiniBot/MaruBot/g' {} + || true
find . -type f -not -path '*/.*' -exec sed -i 's/MARUMINIBOT/MARUBOT/g' {} + || true

# GitHub 레포지토리 주소만 정식 명칭(maru-bot)으로 유지/치환
echo "🌐 GitHub 레포지토리 주소 조정 (dirmich/maru-bot)..."
find . -type f -not -path '*/.*' -exec sed -i 's/dirmich\/marubot/dirmich\/maru-bot/g' {} + || true
find . -type f -not -path '*/.*' -exec sed -i 's/maru-ai\/maru-bot/dirmich\/maru-bot/g' {} + || true

# 파일/디렉토리 이름 치환 (maruminibot -> marubot)
find . -depth -name "*maruminibot*" -not -path '*/.*' | while read -r file; do
    if [ -e "$file" ]; then
        new_file=$(echo "$file" | sed 's/maruminibot/marubot/g')
        mv "$file" "$new_file"
    fi
done

cd "$SOURCE_DIR"
echo -e "\n✅ 모든 치환 및 동기화가 완료되었습니다! 이제 $TARGET_DIR 에서 public 레포지토리 작업을 진행하세요."
