#!/bin/bash

# MaruBot Public Sync Script
# 비공개 소스(maruminibot -> 현재 marubot으로 명칭 변경됨)에서 공개용 파일만 추출하여 ../marubot 폴더로 동기화합니다.

set -e

SOURCE_DIR=$(pwd)
# TARGET_DIR을 절대 경로로 변환 (하위 폴더 이동 시에도 경로 유지)
TARGET_DIR=$(readlink -f "../maru-bot")

echo "🚀 공개 배포용 파일을 $TARGET_DIR 로 동기화합니다..."

# 1. 대상 디렉토리 생성
mkdir -p "$TARGET_DIR"

# 2. 대상 디렉토리 정리 ( .git 보존을 위한 안전한 방식 )
echo "🧹 대상 폴더를 정리 중..."

# .git 폴더 이동(mv)이 윈도우 환경에서 권한 오류(Permission denied)를 일으킬 수 있으므로,
# .git을 제외한 나머지 파일들을 하나씩 찾아서 삭제하는 방식으로 변경합니다.

if [ -d "$TARGET_DIR" ]; then
    # .git이 아닌 모든 파일과 디렉토리를 찾아서 삭제
    # -mindepth 1: 타겟 디렉토리 자체는 포함하지 않음
    # -maxdepth 1: 직계 자식만 대상
    find "$TARGET_DIR" -mindepth 1 -maxdepth 1 -not -name ".git" -exec rm -rf {} +
else
    mkdir -p "$TARGET_DIR"
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
            echo "  � Web Admin 소스 코드 복사 중 (타겟 머신 빌드용)..."
            
            # 1. web-admin 디렉토리 생성 (기존 내용 초기화)
            rm -rf "$TARGET_DIR/web-admin"
            mkdir -p "$TARGET_DIR/web-admin"

            # 2. 소스 파일 복사 (node_modules, .next, .git, dist 제외)
            if command -v rsync >/dev/null 2>&1; then
                echo "  Using rsync..."
                rsync -av --exclude 'node_modules' --exclude '.next' --exclude 'dist' --exclude '.git' "$SOURCE_DIR/web-admin/" "$TARGET_DIR/web-admin/" > /dev/null
            else
                echo "  ⚠️ rsync not found."
                # 윈도우/Git Bash 환경에서 tar 파이프라인이 에러를 낼 수 있음.
                # 복사 후 삭제하는 방식이 가장 안정적일 수 있음.
                echo "  Creating directory..."
                mkdir -p "$TARGET_DIR/web-admin"
                
                # find를 이용해 파일 복사 (node_modules, .next, dist 제외)
                # 주의: 윈도우에서 심볼릭 링크나 경로 문제 발생 가능성 최소화
                echo "  Copying files (excluding node_modules)..."
                
                # web-admin 내부의 항목들 순회
                find "$SOURCE_DIR/web-admin" -mindepth 1 -maxdepth 1 \( -name 'node_modules' -o -name '.next' -o -name 'dist' -o -name '.git' \) -prune -o -exec cp -r {} "$TARGET_DIR/web-admin/" \;
            fi
            
            echo "  ✓ web-admin 소스 복사 완료"
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
# echo "🌐 GitHub 레포지토리 주소 조정 (dirmich/maru-bot)..."
# find . -type f -not -path '*/.*' -exec sed -i 's/dirmich\/marubot/dirmich\/maru-bot/g' {} + || true
# find . -type f -not -path '*/.*' -exec sed -i 's/maru-ai\/maru-bot/dirmich\/maru-bot/g' {} + || true

cd "$SOURCE_DIR"
echo -e "\n✅ 모든 동기화가 완료되었습니다!"
echo "📍 위치: $TARGET_DIR"
