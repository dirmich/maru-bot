# 🦞 MaruBot 표준 배포 운영 절차서 (PUBLISH SOP)

이 문서는 MaruBot의 소스 수정부터 최종 릴리즈까지의 모든 과정을 엄격히 규정합니다. 모든 에이전트(AI)는 배포 및 업데이트 작업 시작 전 이 문서를 반드시 정독하고 각 단계를 준수해야 합니다.

---

## 🚫 불변의 규칙 (Golden Rules)
1.  **비공개 저장소(`maruminibot`)**: 오직 소스 코드의 Source of Truth 역할을 수행합니다. **태그(Tag)나 릴리즈(Release)를 절대 생성하지 마십시오.**
2.  **공개 저장소(`maru-bot`)**: 공식적인 배포 지점입니다. **태깅과 GitHub Release는 오직 이곳에서만 수행합니다.**
3.  **정밀 동기화**: 파일을 복사할 때 저장소 루트가 아닌, **개별 하위 경로(`cmd/marubot/`, `pkg/` 등)**로 정밀하게 이식하십시오.

---

## 🚀 배포 7단계 절대 공정 (7-Step Workflow)

### [1단계] Modify (수정 및 검증)
- 비공개 저장소(`maruminibot`)에서 기능을 수정합니다.
- 로컬 환경에서 크로스 컴파일(`GOOS=linux GOARCH=arm` 등)을 통해 빌드 무결성을 검증합니다.

### [2단계] Version Up (버전 승격)
- 다음 5개 지점의 버전 번호를 동일하게 갱신합니다.
    1.  `pkg/config/version.go`: `const Version = "0.x.x"`
    2.  `cmd/marubot/main.go`: `Version: 0.x.x` (IDENTITY.md 템플릿 내)
    3.  `web-admin/package.json`: `"version": "0.x.x"`
    4.  `Project.md`: 최신 버전 패치 노트 추가.
    5.  `README.md` (및 국어별 README): 헤더 버전 기표 수정.

### [3단계] Internal Push (비공개 커밋)
- 비공개 저장소(`maruminibot`)에서 `git add`, `git commit`, `git push`를 수행합니다.
- **주의**: 태그를 생성하지 마십시오.

### [4단계] Precision Sync (공개 저장소 이식)
- 비공개 저장소의 파일을 공개 저장소(`maru-bot`)의 정위치로 복사합니다.
- **명령 예시 (PowerShell)**:
    - `Copy-Item -Path "cmd/marubot/main.go" -Destination "../maru-bot/cmd/marubot/" -Force`
    - `Copy-Item -Path "pkg/config/version.go" -Destination "../maru-bot/pkg/config/" -Force`
    - `Copy-Item -Path "install.sh", "Project.md" -Destination "../maru-bot/" -Force`

### [5단계] Public Push & Tag (공개 커밋 및 태깅)
- 공개 저장소(`maru-bot`)에서 커밋을 생성하고 **공식 버전 태그**를 생성합니다.
- `git tag 0.x.x` -> `git push origin 0.x.x`

### [6단계] Build (바이너리 패키징)
- 아래 4가지 플랫폼/아키텍쳐만 빌드 및 패키징힙니다. (다른 플랫폼은 절대 빌드/패키징하지 마십시오)
  1. **macOS**: `marubot-macos-amd64.dmg`, `marubot-macos-arm64.dmg` (반드시 DMG 형태로 패키징 및 Notarization 수행)
  2. **Windows**: `marubot-windows-amd64.exe`, `marubot-windows-386.exe`
- **Linux (Raspberry Pi)**: 별도의 바이너리를 릴리즈에 올리지 않습니다. (설치 스크립트가 현장에서 소스 빌드 수행)
- `make public` 명령어를 사용하면 위 4개 에셋이 자동 생성 및 Sync 됩니다.

### [7단계] Publish (최종 고시)
- `gh release create 0.x.x` 또는 `gh release upload`를 수행합니다.
- **업로드 자산**: 오직 아래 4개 파일만 포함하십시오.
  - `marubot-macos-amd64.dmg`
  - `marubot-macos-arm64.dmg`
  - `marubot-windows-amd64.exe`
  - `marubot-windows-386.exe`

---

## 🛠️ 수시 점검 사항
- 파일 인코딩은 항상 **UTF-8 (BOM 없음)**을 유지하십시오.
- `install.sh`의 아카이브 다운로드 주소에 캐시 무효화 파라미터(`?v=버전`)가 포함되었는지 확인하십시오.
