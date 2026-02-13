#!/bin/bash

# MaruBot Public Sync Script
# 비공개 소스(maruminibot -> 현재 marubot으로 명칭 변경됨)에서 공개용 파일만 추출하여 ../marubot 폴더로 동기화합니다.

set -e

SOURCE_DIR=$(pwd)
TARGET_DIR="../maru-bot"

echo "🚀 공개 배포용 파일을 $TARGET_DIR 로 동기화합니다..."

# 1. 대상 디렉토리 생성
mkdir -p "$TARGET_DIR"

# 2. 대상 디렉토리 정리 ( .git 보존을 위한 안전한 방식 )
echo "🧹 대상 폴더를 정리 중..."
GIT_BACKUP_DIR="${TARGET_DIR}_git_backup.tmp"

# 이전 백업이 남아있다면 삭제 (잔여 파일 정리)
if [ -d "$GIT_BACKUP_DIR" ]; then
    rm -rf "$GIT_BACKUP_DIR"
fi

# .git 폴더 백업 (이동)
if [ -d "$TARGET_DIR/.git" ]; then
    echo "🔒 .git 폴더를 안전한 곳으로 임시 이동 중..."
    mv "$TARGET_DIR/.git" "$GIT_BACKUP_DIR"
else
    echo "⚠️  주의: $TARGET_DIR/.git 폴더가 존재하지 않음 (신규 생성?)"
fi

# 대상 폴더 내용 전체 삭제
# (이제 .git이 없으므로 안심하고 삭제 가능)
# 단, 대상 폴더 자체는 남겨둠
find "$TARGET_DIR" -mindepth 1 -maxdepth 1 -exec rm -rf {} +

# .git 폴더 복원 (이동)
if [ -d "$GIT_BACKUP_DIR" ]; then
    echo "🔓 .git 폴더 복원 중..."
    mv "$GIT_BACKUP_DIR" "$TARGET_DIR/.git"
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
