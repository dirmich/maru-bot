# 🦞 MaruBot (마루 미니봇) 프로젝트 이력 (History)

## 2026-04-18
### 0.6.8
- **텔레그램 마크다운 변환 버그 수정**: 여러 개의 코드 블록 또는 인라인 코드가 포함된 메시지에서 모든 블록이 동일한 내용(주로 마지막 IP 주소)으로 치환되는 심각한 오류 해결.
- **배포 자동화 적용**: 버전 상향 시 Web Admin 빌드 및 소스 코드 동기화 절차 전면 적용.

## 2026-04-06
### 0.6.7
- **Settings 모델 선택 보정**: 메인 에이전트 모델 선택 시 `agents.defaults.provider`와 `agents.defaults.model`이 항상 함께 저장되도록 UI 저장 경로를 정리.
- **중복 모델명 복원 개선**: 같은 모델명이 여러 provider에 있을 때도 현재 설정값을 `provider::model` 기준으로 정확히 복원하도록 선택값 해석 로직을 보강.
- **Fallback 스위치 동작 수정**: `Fallback Models` 항목의 스위치가 실제로 enable/disable 되도록 수정하고, 메인 모델과 중복되는 fallback 항목은 자동으로 정리.

### 0.6.6
- **Provider 활성화 스위치 추가**: 각 LLM provider와 Ollama 인스턴스를 개별적으로 enable/disable 할 수 있도록 설정 구조와 Web Admin UI를 확장.
- **Fallback 모델 식별자 명확화**: `agents.defaults.fallback_models`를 `provider::model` 형식으로 저장하도록 변경하여 provider와 모델의 대응 관계를 명시.
- **기본 모델 provider 저장 수정**: Settings에서 모델 선택 시 `agents.defaults.provider`와 `agents.defaults.model`이 함께 정확히 저장되도록 보강.
- **Publish 규칙 강화**: `publish` 요청 시 version up, web-admin build, binary build, `../maru-bot` 동기화, 두 저장소 commit/push를 자동 수행하도록 규칙 문서에 명시.

### 0.6.5
- **릴리스 경로 안정화**: `Makefile`의 `sync-ui`가 `web-admin/dist`를 잘못 참조하던 문제를 수정하여 `make public`이 UI 동기화 단계에서 중단되지 않도록 개선.
- **macOS 서명 환경 자동 로드**: `scripts/build_dmg.sh`가 기본적으로 `.env.signing`을 읽어 `SIGNING_IDENTITY`, `AC_APPLE_ID`, `AC_PASSWORD`, `AC_TEAM_ID`를 자동으로 로드하도록 정리.
- **키체인 승인 가이드 문서화**: macOS 서명 실패(`errSecInternalComponent`) 대응을 위해 개인키 접근 승인 절차를 `docs/macos-keychain-signing-access.md`로 문서화.
- **공개 릴리스 자산 재업로드**: Web Admin 재빌드 후 Windows 실행 파일과 macOS DMG를 다시 생성하여 `maru-bot` 공개 릴리스 `0.6.5` 자산을 최신 상태로 갱신.
- **web-admin provider 노출 보강**: 설정 페이지의 provider 추가 목록에 `llamacpp`를 포함하고 기본 API Base를 `http://localhost:8080/v1`로 지정.
- **llama.cpp 모델 조회 연결**: Web Admin의 `/api/config/fetch-models` 경로에서 `llamacpp`를 OpenAI 호환 provider로 처리하도록 확장.
- **로컬 서버 설정 UX 개선**: `llamacpp`는 `ollama`와 동일하게 API Key 없이도 추가 및 모델 조회가 가능하도록 정리.

### 0.6.4
- **llama.cpp 프로바이더 공식 지원 추가**: `llamacpp` 타입을 명시적으로 지원하여 로컬 LLM 서버 연동성 개선.
- **도구 스키마 호환성 최적화**: `llama.cpp`의 엄격한 문법 변환기를 고려하여 `required` 및 `additionalProperties` 제약 조건을 자동으로 제거하는 스키마 단순화 로직 구현 (400 Bad Request 에러 해결).
- **디버깅 로그 강화**: `llamacpp` 프로바이더 사용 시 실제 LLM 요청 본문을 로그로 출력하여 가시성 확보.
- **환경 변수 지원**: `MARUBOT_HOME` 환경 변수를 통한 커스텀 홈 디렉토리 지정 기능의 경로 분석 로직 강화.

## 2026-03-26
### v0.6.3
- **Go 네이티브 Browser 도구 통합**: `cheliped-browser`를 `chromedp` 기반의 네이티브 도구인 `gobrowser`로 포팅하여 외부 의존성 제거 및 성능 향상.
- **종합 관리 서비스 검증**: Webhook을 통한 동적 Cron 작업 등록 및 Heartbeat 서비스 연동성 강화.
- **빌드 안정성 개선**: `Makefile`의 `sync-ui` 구문 오류 수정 및 `install.sh` 내 저사양 기기(RPi3 등) OOM 방지를 위한 빌드 병렬성 제한(`-p=1`) 로직 추가.
- **다국어 문서 동기화**: 모든 언어별 README의 버전을 0.6.3으로 통일하고 배포 가이드(`versionup.md`)를 보강.


## 2026-03-24
###Version: 0.6.1
- **빌드 오류 수정**: Linux 및 기타 플랫폼에서 `getSysProcAttr` 식별자를 찾지 못해 빌드가 실패하던 컴파일 오류 해결 (`sys_linux.go`, `sys_default.go` 추가).

### 0.5.8
- **공개 레포지토리 빌드 안정화**: `maru-bot` 프로젝트와 같이 `web-admin` 소스 코드가 없는 환경(Linux 등)에서 `make` 실행 시 UI 빌드 단계에서 발생하던 오류 해결 (기빌드된 자산이 있을 경우 빌드 생략 로직 추가).

### 0.5.7
- **데이터 마이그레이션 도구 도입**: `marubot migrate-paths` 명령어를 추가하여 기존 `systemprofile` 경로로 저장된 스킬 및 도구 메타데이터를 현재 사용자 경로로 일괄 업데이트하는 기능 제공.
- **안정성 개선**: 도구 및 스킬 실행 시 경로 인식 오류 방지 및 버전업.

### 0.5.6
- **경로 분석 로직 정교화**: `MARUBOT_HOME` 환경 변수를 시스템 전반(config.go, agent loop, tools)에서 일관되게 인식하도록 수정.
- **Home 확장 기능 수정**: `~/.marubot`으로 시작하는 설정 경로가 커스텀 `MARUBOT_HOME` 지정 시에도 올바른 절대 경로로 변환되지 않던 버그 해결.
- **환경 변수 전파**: `--home` 명령행 인자 사용 시 해당 경로를 `MARUBOT_HOME` 환경 변수로 강제 할당하여 모든 패키지가 동일한 경로를 바라보도록 개선.

### 0.5.5
- **Windows 서비스 설정 경로 수정**: Windows 서비스 실행 시 설정 디렉토리가 시스템 프로필(`systemprofile`)로 고정되는 문제 해결.
- **경로 재정의 지원**: `MARUBOT_HOME` 환경 변수 및 `--home` 명령행 인자를 통한 설정 디렉토리 커스텀 경로 지정 기능 추가.
- **서비스 설치 개선**: 서비스 설치 시 현재 사용자의 홈 디렉토리를 자동으로 감지하여 서비스 실행 인자로 전달하도록 최적화.
- **GPIO 서버 로직 정제**: 대시보드 서버의 GPIO 설정 저장 방식 개선 및 중복 코드 제거.

## 2026-03-23
### 0.5.4
- 6단계 표준 배포 워크플로우(`RULES.md`) 수립 및 적용
- 채널 간 메시지 전송(Channel to Channel) 도구 추가: AI가 `send_channel_message` 도구를 사용하여 텔레그램↔슬랙 등 이기종 채널 간 메시지 중계 기능 지원
- 슬랙(Slack) 스레드(Thread) 응답 및 `app_mention` 이벤트 처리 강화
- 설정 자동 동기화 및 구조 정형화: `config.json` 누락 필드 자동 갱신 및 레거시 필드 제거
- 전체 저장소(`maruminibot`, `maru-bot`) 및 문서(README, IDENTITY) 버전 정합성 일원화

### 0.4.91
- 언인스톨(`uninstall`) 로직 개선: 표준 설치 경로(`~/.marubot/bin`) 외에서 실행 시 본체 파일 삭제 여부를 사용자에게 확인하도록 변경 (설치 파일 오삭제 방지)
- 버전 업데이트 및 안정화

### 0.4.89
- 설정 파일 저장 경로 불일치 수정 (`~/.marubot/config.json`으로 통일)
- 윈도우 업그레이드 시 터미널 플래시 방지 및 네이티브 다이얼로그 UI 도입
- 슬랙 연동 장애 디버깅을 위한 소켓 모드 상세 로깅 추가
- 트레이 메뉴 UX 개선 (업그레이드 확인 시 진행 상황 알림 추가)
- 트레이 메뉴(Windows/macOS)에 '업그레이드 확인' 기능 추가

### 0.4.87
- **Slack 연동 개선**: 
    - 슬랙 채널의 로깅 시스템을 메인 로거(`pkg/logger`)와 통합하여 대시보드 가시성 확보.
    - 슬랙 연결 성공, 실패 및 메시지 처리 과정에 대한 상세 디버그 로그 추가.
    - 린트 에러 수정 및 안정성 강화.

## 2026-03-21
### 0.4.86
- **Web Admin UI 개선**: 
    - 설정 페이지 레이아웃 재배치 (AI 프로바이더 우선순위 조정).
    - 시스템 언어 선택기를 드롭다운(Select) 방식으로 변경하여 UI 집약도 향상.
    - Webhook(포트/경로/보안키), Slack/WhatsApp(허용 ID) 등 누락된 상세 설정 필드 추가.
    - 페이지 최하단에 '설정 저장' 버튼 추가 및 처리 결과 피드백 강화.
- **Uninstall 로직 수정**: Windows 환경에서 프로세스 종료(`taskkill`) 및 서비스 제거(`sc delete`) 안정성 강화.
- **다국어 현지화**: 신규 설정 항목에 대해 한국어, 영어, 일본어 번역 적용.

## 2026-03-20
### 0.4.85
- **Admin System**: Bun.sh + PostgreSQL 기반 중앙 관리 서비스 구축 (Backend/Frontend).
- **Google SSO**: 서비스 다운로드 및 인스턴스 관리를 위한 Google 로그인 연동.
- **통계 대시보드**: 전체 사용자 및 플랫폼별 설치 현황을 위한 슈퍼유저 전용 UI 구현.
- **Client Integration**: Marubot 인스턴스 자동 상태 보고(30분 주기) 기능 탑재.
- **글로벌 문서**: 모든 README(KO, EN, JA, CN) 및 RULES.md 버전 동기화 및 최신화.
- **브랜딩 정제**: 모든 문서 및 소스 코드에서 외부 프로젝트(PicoClaw/nanobot) 의존적 설명 제거 및 독자 브랜드 강화.
- **링크 수정**: IDENTITY 및 소스 코드 내의 플레이스홀더 링크를 공식 공개 저장소(`maru-bot`) 링크로 전면 수정.

### 0.4.84
- **공식 홈페이지(Landing Page)**: React + Tailwind CSS + shadcn/ui 기반 멀티링구얼 지원 페이지 구축.
- **다국어 지원**: 한국어, 영어, 일본어, 스페인어 4개 국어 및 라이트/다크 테마 지원.
- **설치 가이드**: 윈도우, macOS(Intel/Silicon), Linux/RPi 전용 설치 안내 및 쉘 스크립트 제공.
- **브랜딩**: `app_icon.png`를 로고 및 파비콘으로 적용하여 일관성 확보.

### 0.4.83
- **채널 정제**: Slack(Socket Mode), WhatsApp, Telegram, Discord, Webhook 5개 핵심 채널로 집중.
- **설정 UI 개선**: 채널별 맞춤형 입력 필드 및 상세 토큰 발급 가이드(다국어) 팝업 추가.
- **백엔드 최적화**: 사용하지 않는 채널(Feishu, MaixCam) 제거 및 설정 구조 간소화.

### 0.4.82
- **아이콘 최적화**: 모든 아이콘의 흰색 여백 자동 제거 및 꽉 차게 리디자인.
- **Windows 트레이 개선**: 윈도우 환경 전용 `window_tray.ico` 적용 및 가시성 개선.

## 2026-03-17
### 0.4.74
- **GPIO 테스트 플래그 도입**: `config.json`의 `hardware.gpio_test_mode` 플래그를 통해 윈도우 등 비-기기 환경에서도 GPIO 기능을 강제 활성화 가능.
- **GPIO 시뮬레이션 (Windows)**: 하드웨어가 없는 환경에서 가상의 핀 상태(Memory-based)를 제어하고 읽을 수 있는 시뮬레이션 핸들러 구현.
- **동작 세분화**: 핀 모드(`IN`/`OUT`)를 자동 판별하여 입력 핀은 **읽기(Read)**, 출력 핀은 **토글(Toggle)** 로 동작하도록 고도화.
- **설정 파일 단일화**: `usersetting.json`을 폐기하고 모든 설정을 `config.json` 하나로 관리 (기존 설정 자동 마이그레이션 지원).
- **GPIO 그룹 가시화**: `config.json`의 중첩 구조(예: `motor_a`, `motor_b`)를 인식하여 웹 관리자에서 자동으로 **그룹 박스(Group Box)** Card UI로 렌더링.
- **로깅 추적성 강화**: 웹 관리자를 통한 GPIO 조작 시 로그에 `[WebAdmin Access]` 접두어를 추가하여 동작 출처를 명확히 함.
- **로그 가시성 개선**: 웹 관리자 로그 페이지 하단에 실제 로그 파일 경로(`~/.marubot/dashboard.log`) 명시.
- **UI 피드백 개선**: 웹 관리자 토스트 메시지를 동작 유형에 따라 차별화 (입력: 'is HIGH', 출력: 'toggled to HIGH').
- **버그 수정**: `config.json` 로딩 시 GPIO 핀 설정이 기본값(DefaultConfig)과 병합되는 현상 수정.
- **버그 수정**: 백엔드 API 응답 필드명 불일치(`is_rpi` → `is_raspberry_pi`)를 해결하여 웹 관리자 사이드바에서 GPIO 메뉴가 정상 노출되도록 수정.
- **로깅 강화**: 모든 가상 GPIO 조작 내역을 `dashboard.log`에 `[GPIO Simulation]` 접두어와 함께 실시간 기록.

### 0.4.73
- **시스템 프롬프트 갱신**: 마크다운 테이블 내 줄바꿈 가이드 보강 (`<br>` 사용 허용 정책 수립).
- **Web Admin 기능 강화**: `rehype-raw` 적용으로 테이블 내 줄바꿈(`<br>`) 실제 렌더링 지원.
- **Go 코드 안정화**: 프롬프트 문자열 내 백틱(`) 사용으로 인한 컴파일 에러 수정.
- **작업 규칙 업데이트**: `RULES.md`에 `Project.md` 기록 의무화(날짜/버전 명시) 조항 추가.

### 0.4.72
- **Web Admin 마크다운 엔진 도입**: `react-markdown` 및 `remark-gfm` 설치 및 연동.
- **스타일 고도화**: 테이블, 코드 블록, 인용구 등 마크다운 요소를 위한 전용 CSS 테마 적용.

### 0.4.71.3
- **에이전트 도구 인식 개선**: 레거시 JSON 형식(`action`, `key`) 응답 시 자동으로 `config` 도구로 매핑하여 실행하도록 파싱 로직 보강.

### 0.4.71.2
- **에이전트 블로킹 문제 해결**: `MessageBus.PublishOutbound`를 비차단(Non-blocking) 방식으로 수정하여 채널 미활성화 시 루프가 멈추는 현상 해결.
- **디버깅 강화**: Dashboard `handleChat` 핸들러 진입점부터의 상세 디버그 로깅 추가.

## 2026-03-16
### 0.4.71.1
- **서버 안정성 보호**: Dashboard 서버에 최상위 패닉 복구(Panic Recovery) 미들웨어 적용.
- **초기화 크래시 방지**: `Setup Mode`에서 `dummyAgent` 호출 시 발생하는 Nil Pointer Panic 해결을 위해 정식 초기화 루틴 적용.

## 2026-03-15
### 0.4.70.1
- **설치 스크립트 개선**: `install.sh`의 Go 아키텍처 자동 감지 및 버전 비교 로직 수정 (불필요한 재설치 방지).
- **보안 수정**: 기존 비밀번호 추출 로직 보완을 통한 401 인증 에러 원인 해결.

## 2026-03-14
### 0.4.70
- **버전 표기 버그 수정**: Windows 환경에서의 버전 정보 일관성 확보.
- **UI 최적화**: 비-라즈베리파이 환경에서 GPIO 메뉴가 보이지 않도록 필터링 로직 강화.

## 2026-03-12
### 0.4.69
- **데이터 일관성**: `IDENTITY.md`와 바이너리 간의 버전 정보 동기화.
- **파싱 로직 개선**: 텍스트와 혼용된 JSON 도구 호출 블록에 대한 추출 및 파싱 정확도 향상.
