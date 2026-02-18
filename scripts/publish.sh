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
            echo "  🍳 Web Admin 로컬 빌드 시작..."
            
            # 빌드 디렉토리로 이동
            if cd "$SOURCE_DIR/web-admin"; then
                # 의존성 설치 및 빌드 (Next.js Standalone 모드)
                echo "    - 기존 빌드 정리 (.next 삭제)..."
                rm -rf .next

                echo "    - 의존성 확인 및 빌드 (npm)..."
                # 윈도우 환경 호환성을 위해 bun 대신 npm 사용 권장
                npm install
                npm run build
                
                echo "  📦 Web Admin 빌드 결과물(Standalone) 배포 중..."
                mkdir -p "$TARGET_DIR/web-admin"

                # 1. Standalone 결과물 복사 (server.js, package.json 및 .next 숨김 폴더 포함)
                # cp -a ./. 사용 시 점(.)으로 시작하는 숨김 폴더(.next 등)도 모두 복사됨
                cp -a .next/standalone/. "$TARGET_DIR/web-admin/"
                
                # 2. Static 리소스 복사 (.next/static -> .next/static)
                # Standalone 모드는 static 파일을 별도로 서빙해야 하므로 구조를 맞춰줘야 함
                mkdir -p "$TARGET_DIR/web-admin/.next/static"
                cp -R .next/static/* "$TARGET_DIR/web-admin/.next/static/"
                
                # 3. Public 폴더 복사
                cp -R public "$TARGET_DIR/web-admin/"

                # 중요: 로컬(Windows/Mac)의 node_modules는 리눅스(ARM)와 호환되지 않을 수 있음
                # 특히 sharp, sqlite3 같은 네이티브 모듈.
                # 따라서 node_modules는 제외하고, 타겟 머신에서 'bun install'을 수행하도록 유도해야 함.
                # 하지만 standalone 폴더는 node_modules를 포함하고 있음.
                # 안전을 위해 복사된 node_modules를 삭제함.
                if [ -d "$TARGET_DIR/web-admin/node_modules" ]; then
                    echo "    - 플랫폼 호환성을 위해 로컬 node_modules 제거 (타겟에서 재설치 필요)"
                    rm -rf "$TARGET_DIR/web-admin/node_modules"
                fi

                echo "  ✓ web-admin 빌드 및 배포 완료"
                
                # 다시 루트로 복귀
                cd "$SOURCE_DIR"
            else
                echo "  ❌ web-admin 디렉토리 진입 실패"
                exit 1
            fi
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
