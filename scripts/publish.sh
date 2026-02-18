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

                # 명시적으로 필요한 파일들만 복사 (숨김 폴더 .next 포함)
                echo "    - server.js 및 package.json 복사"
                cp .next/standalone/server.js "$TARGET_DIR/web-admin/"
                cp .next/standalone/package.json "$TARGET_DIR/web-admin/"
                
                if [ -d ".next/standalone/.next" ]; then
                    echo "    - .next (standalone build data) 폴더 복사"
                    cp -a .next/standalone/.next "$TARGET_DIR/web-admin/"
                else
                    echo "    ⚠️ .next/standalone/.next 폴더를 찾을 수 없습니다!"
                fi
                
                # 2. Static 리소스 복사 (.next/static -> .next/static)
                echo "    - .next/static 리소스 복사"
                mkdir -p "$TARGET_DIR/web-admin/.next/static"
                if [ -d ".next/static" ]; then
                    cp -a .next/static/. "$TARGET_DIR/web-admin/.next/static/"
                fi
                
                # 3. Public 폴더 복사
                echo "    - public 폴더 복사"
                if [ -d "public" ]; then
                    cp -a public "$TARGET_DIR/web-admin/"
                fi

                # 4. Prisma 폴더 복사 (스키마 재생성을 위해 필요)
                echo "    - prisma 폴더 복사"
                if [ -d "prisma" ]; then
                    cp -a prisma "$TARGET_DIR/web-admin/"
                fi

                # node_modules는 타켓에서 설치하도록 제외
                if [ -d "$TARGET_DIR/web-admin/node_modules" ]; then
                    echo "    - 기존 node_modules 제거"
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
