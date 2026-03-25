# MaruBot (마루봇) 전방위 버전업 및 릴리즈 절차 (Version Upgrade Guide)

항상 아래 순서를 엄격히 준수하여 모든 플랫폼과 UI에서 버전이 동기화되도록 합니다.

## 1. 개요
*   **목표**: 소스 코드, 바이너리, 웹 관리자(WebAdmin), 그리고 공개 저장소(GitHub)의 버전을 하나로 일치시킴.
*   **기준 버전**: `v0.6.2` (예시)

## 2. 사전 준비
1.  작업 중인 모든 저장소를 최신 상태로 유지 (`git pull`).
2.  이전 테스트 잔여물 삭제 (`rm -rf tmp/`).
3.  Git 잠금 파일 확인 및 제거 (`rm -f .git/index.lock`).

## 3. 소스 코드 버전 업데이트 (Manual)
반드시 다음 파일들을 찾아 직접 수정합니다:
*   [ ] **Back-end Core**: `pkg/config/version.go` -> `const Version = "0.6.2"`
*   [ ] **Identity Template**: `cmd/marubot/main.go` 내 `IDENTITY.md` 섹션의 `Version` 필드 수정. (매우 중요: WebAdmin 초기화 시 사용됨)
*   [ ] **Front-end**: `web-admin/package.json` -> `"version": "0.6.2"`
    *   *주의*: UI 빌드 전 `web-admin/dist`를 반드시 삭제해야 최신 버전이 반영됩니다.
*   [ ] **Documentation**: `README.md`, `README-en.md` 등의 첫 줄 버전 번호 수정.
*   [ ] **History**: `Project.md`에 새 버전 일시 및 변경 사항 기록.
*   [ ] **Makefile**: `LDFLAGS`의 `-X main.Version` 대소문자 확인 (반드시 대문자 V).

## 4. UI 빌드 및 자산 동기화 (Critical)
웹 관리자의 버전을 반영하기 위해 클린 빌드가 필수입니다.
```bash
cd web-admin
rm -rf dist
npm install
npm run build
cd ..

# 백엔드 임베드 경로로 복사
rm -rf cmd/marubot/dashboard/dist
mkdir -p cmd/marubot/dashboard/dist
cp -r web-admin/dist/* cmd/marubot/dashboard/dist/
```

## 5. 내부 저장소 (`maruminibot`) 동기화 및 릴리즈
1.  빌드 캐시 정리: `go clean -cache`
2.  커밋 및 푸시:
    ```bash
    git add .
    git commit -m "Release v0.6.2"
    git push
    ```
3.  **태그 지정 (중요)**:
    ```bash
    git tag v0.6.2
    git push origin v0.6.2
    ```

## 6. 공개 저장소 (`maru-bot`) 동기화 및 GitHub 릴리즈
1.  동기화 스크립트 실행: `bash scripts/publish.sh`
2.  공개 저장소 소스 커밋:
    ```bash
    cd ../maru-bot
    git add .
    git commit -m "Sync with latest maruminibot (v0.6.2)"
    git push
    ```
3.  **공개 태그 및 릴리즈**:
    ```bash
    git tag v0.6.2
    git push origin v0.6.2
    # gh CLI가 설치된 경우 publish.sh가 자동으로 수행하나 태그 푸시는 별도로 확인 요망
    ```

## 7. 최종 확인
1.  **로컬 버전**: `marubot version` 명령으로 확인.
2.  **WebAdmin**: 브라우저 접속 후 **Ctrl + F5** 강제 새로고침하여 하단 또는 설정 페이지의 버전 확인.
3.  **업그레이드 체크**: `marubot upgrade` 시 최신 버전이 정상 인식되는지 확인.
