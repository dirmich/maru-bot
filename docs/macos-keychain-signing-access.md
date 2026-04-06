# macOS Keychain 개인키 사용 승인 가이드

이 문서는 MaruBot macOS 배포 시 `codesign`과 `notarytool`이 사용할 인증서 개인키 접근 권한을 승인하는 절차를 정리합니다.

대상 인증서:

- `Developer ID Application: Highmaru, Inc. (X4V7W6GE8X)`

관련 증상:

- `codesign ... errSecInternalComponent`
- `security show-keychain-info ...: User interaction is not allowed.`
- 인증서는 `security find-identity -v -p codesigning` 에서 보이지만 실제 서명이 실패함

## 1. 사전 확인

현재 로그인 키체인에 서명 인증서가 있는지 확인합니다.

```bash
security find-identity -v -p codesigning
```

정상이라면 아래와 비슷한 항목이 보여야 합니다.

```text
"Developer ID Application: Highmaru, Inc. (X4V7W6GE8X)"
```

## 2. 로그인 키체인 잠금 해제

키체인이 잠겨 있으면 개인키 사용 승인이 진행되지 않습니다.

```bash
security unlock-keychain ~/Library/Keychains/login.keychain-db
```

필요하면 기본 키체인도 다시 지정합니다.

```bash
security default-keychain -s ~/Library/Keychains/login.keychain-db
```

## 3. CLI 도구의 개인키 접근 권한 허용

아래 명령은 `codesign`과 Apple 도구가 개인키를 사용할 수 있도록 Access Control 목록을 갱신합니다.

```bash
security set-key-partition-list -S apple-tool:,apple:,codesign: -s -k '<macOS 로그인 암호>' ~/Library/Keychains/login.keychain-db
```

주의:

- `<macOS 로그인 암호>`는 실제 로그인 암호로 바꿔야 합니다.
- 이 단계가 `errSecInternalComponent` 해결의 핵심인 경우가 많습니다.

## 4. Keychain Access 앱에서 수동 승인

터미널 명령만으로 해결되지 않으면 GUI에서 직접 승인합니다.

1. `Keychain Access` 앱을 엽니다.
2. `login` 키체인을 선택합니다.
3. `My Certificates` 로 이동합니다.
4. `Developer ID Application: Highmaru, Inc. (X4V7W6GE8X)` 인증서를 찾습니다.
5. 인증서 아래의 개인키를 펼칩니다.
6. 개인키를 더블클릭한 뒤 `Access Control` 탭을 엽니다.
7. 아래 둘 중 하나를 선택합니다.

- `Confirm before allowing access`
- `Allow all applications to access this item`

배포 자동화를 위해서는 아래가 더 실용적입니다.

- `Allow all applications to access this item`

보수적으로 운영하려면 개별 앱만 추가합니다.

- `/usr/bin/codesign`
- `Xcode`
- 사용하는 터미널 앱 (`Terminal.app`, `iTerm.app` 등)

## 5. 서명 환경 변수 확인

MaruBot 릴리스 스크립트는 기본적으로 루트의 `.env.signing` 을 읽습니다.

예시:

```bash
cat .env.signing
```

```text
export SIGNING_IDENTITY="Developer ID Application: Highmaru, Inc. (X4V7W6GE8X)"
export AC_APPLE_ID="apps@highmaru.com"
export AC_PASSWORD="..."
export AC_TEAM_ID="X4V7W6GE8X"
```

다른 파일을 쓰고 싶으면 `MARUBOT_SIGNING_ENV` 로 경로를 덮어쓸 수 있습니다.

```bash
MARUBOT_SIGNING_ENV=/path/to/.env.signing ./scripts/build_dmg.sh amd64
```

## 6. 서명 테스트

`.env.signing` 자동 로드가 동작하는지 포함해 바로 테스트합니다.

```bash
env -u SIGNING_IDENTITY -u AC_APPLE_ID -u AC_PASSWORD -u AC_TEAM_ID \
GOCACHE=/tmp/marubot-gocache \
GOPATH=/tmp/marubot-gopath \
./scripts/build_dmg.sh amd64
```

정상일 때 기대 흐름:

1. `Loading signing environment from .../.env.signing`
2. `Signing binary and app bundle with identity: Developer ID Application: Highmaru, Inc. ...`
3. notarization submit
4. stapling 완료
5. `Created build/marubot-macos-amd64.dmg`

## 7. 전체 공개 배포 테스트

서명 테스트가 통과하면 전체 퍼블리시를 실행합니다.

```bash
GOCACHE=/tmp/marubot-gocache GOPATH=/tmp/marubot-gopath make public
```

이 경로는 다음을 수행합니다.

- Web Admin 빌드
- Windows 실행 파일 빌드
- macOS DMG 빌드
- `../maru-bot/releases` 동기화
- GitHub Release 자산 업로드

## 8. 문제 해결 체크리스트

`codesign` 이 계속 실패하면 아래를 순서대로 점검합니다.

1. 로그인 키체인이 실제로 잠금 해제되어 있는지 확인
2. 인증서와 개인키가 같은 키체인에 있는지 확인
3. 개인키 `Access Control` 에 `codesign` 또는 터미널 앱이 허용되어 있는지 확인
4. `security set-key-partition-list` 를 다시 실행
5. macOS 세션을 다시 로그인하거나 터미널을 재실행
6. Xcode Command Line Tools가 정상인지 확인

## 9. 현재 MaruBot 릴리스 스크립트 동작

현재 [build_dmg.sh](/Volumes/2T-SSD/work/0.Project/0.ai/maruminibot/scripts/build_dmg.sh) 는 아래 순서로 동작합니다.

1. `SIGNING_IDENTITY` 가 비어 있으면 `.env.signing` 자동 로드
2. macOS 앱 번들 생성
3. `codesign`
4. `notarytool submit`
5. `stapler staple`
6. 최종 DMG 생성 및 필요 시 다시 서명/노타리제이션

따라서 `.env.signing` 이 있어도 키체인 권한이 막혀 있으면 자동 배포는 실패합니다.
