# MaruBot 프로젝트 이력

## 2026-06-27
### 0.9.13
- **날짜/시간 검증 harness 보강**: 시스템 현황 및 날짜/시간 질문에 실제 `date`/`timedatectl` 출력과 UTC 시간을 포함해 모델이 현재 날짜/시간을 추정하지 않도록 수정.

### 0.9.12
- **Web Admin 언어 설정 동기화 수정**: 설정 화면 진입 시 서버 `config.language`를 UI language store에 반영해 localStorage 기본값(`en`)이 저장 시 서버 언어를 덮어쓰지 않도록 수정.

### 0.9.11
- **설정 언어 hot-reload 보강**: Web Admin에서 언어를 저장하면 실행 중인 AgentLoop의 시스템 프롬프트 컨텍스트도 즉시 재생성해 다음 응답부터 선택 언어를 반영.

### 0.9.10
- **응답 언어 지시 강화**: 설정 언어 코드(`ko` 등)를 실제 언어명으로 풀어 시스템 프롬프트에 주입해 harness 응답이 영어로 새지 않도록 보강.

### 0.9.9
- **시스템 현황 검증 harness 추가**: 시스템/서버 현황 질문에서는 LLM 호출 전 실제 `shell` 명령을 실행해 검증된 상태 출력을 컨텍스트에 주입.
- **거짓 시스템 정보 방지 프롬프트 강화**: `VERIFIED SYSTEM STATUS HARNESS`가 있으면 해당 출력만 근거로 답하도록 시스템 프롬프트 보강.

### 0.9.8
- **WebAdmin 업그레이드 복구**: Linux/RPi 업그레이드 중 실행 중인 user systemd 서비스를 먼저 중지해 설치가 끊기던 문제를 수정.
- **비대화형 설치 지원**: 대시보드/AI 업그레이드에서 `install.sh`가 `/dev/tty` 입력을 기다리지 않도록 `MARUBOT_NONINTERACTIVE=1` 경로를 추가.

### 0.9.7
- **GPIO 텍스트 호출 실행 보정**: 모델이 `call:gpio_control{action: "status"}` 형식으로 응답해도 실제 `gpio_control status` 도구 호출로 변환하도록 수정.
- **GPIO 기본 스킬 내장 fallback**: 바이너리 업그레이드 환경에서 `skills/gpio` 파일이 누락돼도 기본 `gpio` 스킬이 로드되도록 보강.
- **WebAdmin 기본 스킬 목록 보정**: WebAdmin `/api/skills`가 워크스페이스 스킬뿐 아니라 기본 스킬 디렉터리도 조회하도록 수정.

### 0.9.6
- **GPIO 상태 조회 도구화**: `gpio_control`에 전체 현재 GPIO 레벨을 반환하는 `status` 액션을 추가.
- **RPi GPIO 기본 스킬 추가**: GPIO가 활성화된 환경에서 `gpio` 스킬을 자동 로드하여 상태 질문이 `gpio_control status`를 사용하도록 보강.
- **도구 호출 복구 강화**: 모델이 `hardware.gpio.pins.*` 설정 조회 JSON을 본문에 출력하는 경우 `gpio_control status` 호출로 복구.

## 2026-04-21
### 0.9.0
- **GPIO 지능형 대시보드**: 충돌 감지와 실시간 상태 확인 기능을 갖춘 GPIO 대시보드 추가.
- **설치/업그레이드 안정화**: 설치 스크립트의 정리 로직과 릴리즈 다운로드 경로를 보강.
- **라즈베리파이 빌드 안정화**: Go 1.25+ 환경에서 의존성 문제가 발생하지 않도록 replace 지시어와 패키지 버전을 고정.
- **버전 동기화 SOP 정리**: `versionup.md` 기준으로 Identity, 문서, UI 버전을 함께 관리하도록 정리.

### 0.7.2
- **대시보드 상세 정보 강화**: Windows 환경의 상세 하드웨어 메트릭과 진행 UI를 개선.
- **히스토리 서버 지원**: 자동 저장과 사이드바 이력 관리 기능을 추가.
- **빌드 안정화**: 의존성 정리와 Windows 빌드 보정을 적용.

## 2026-04-18
### 0.6.8
- **텔레그램 마크다운 변환 수정**: 여러 코드 블록이나 인라인 코드가 포함된 메시지에서 마지막 값으로 모두 치환되던 오류 수정.
- **배포 자동화 적용**: 버전 상승, Web Admin 빌드, 소스 동기화 절차를 배포 흐름에 반영.

## 2026-04-06
### 0.6.7
- **Settings 모델 선택 보정**: 메인 에이전트 모델 선택 시 `agents.defaults.provider`와 `agents.defaults.model`이 함께 저장되도록 수정.
- **중복 모델명 복원 개선**: 같은 모델명이 여러 provider에 있어도 `provider::model` 기준으로 정확히 복원하도록 보강.
- **Fallback 스위치 동작 수정**: fallback 모델 스위치가 실제로 enable/disable 되도록 수정.

### 0.6.6
- **Provider 활성화 스위치 추가**: LLM provider와 Ollama 인스턴스를 개별 enable/disable 할 수 있도록 설정 구조와 Web Admin UI 확장.
- **Fallback 모델 구조 명확화**: `agents.defaults.fallback_models`를 `provider::model` 형식으로 저장하도록 변경.
- **기본 모델 provider 저장 보정**: Settings에서 모델 선택 시 provider와 model이 함께 저장되도록 수정.
- **Publish 규칙 강화**: `publish` 요청 시 version up, web-admin build, binary build, 공개 저장소 동기화, 양쪽 저장소 commit/push를 수행하도록 문서화.

### 0.6.5
- **릴리즈 경로 안정화**: `Makefile`의 `sync-ui`가 올바른 `web-admin/dist`를 참조하도록 수정.
- **macOS 서명 환경 자동 로드**: `scripts/build_dmg.sh`가 `.env.signing` 값을 자동으로 읽도록 정리.
- **llama.cpp Web Admin 노출**: Settings provider 목록에 `llamacpp`를 추가하고 기본 API Base를 `http://localhost:8080/v1`로 지정.
- **llama.cpp 모델 조회 연결**: `/api/config/fetch-models`에서 `llamacpp`를 OpenAI 호환 provider로 처리.

### 0.6.4
- **llama.cpp provider 지원 추가**: 로컬 LLM 서버 연동을 위한 `llamacpp` provider 타입 추가.
- **도구 스키마 호환성 개선**: llama.cpp가 거부하는 일부 JSON Schema 제약을 자동으로 제거하도록 보강.
- **디버그 로그 강화**: llama.cpp 요청 본문을 확인할 수 있도록 provider 로그를 개선.
- **`MARUBOT_HOME` 경로 지원 강화**: 사용자 지정 홈 디렉터리 처리 로직을 보강.

## 2026-03-26
### 0.6.3
- **Go 네이티브 Browser 도구 통합**: `cheliped-browser`를 `chromedp` 기반 `gobrowser`로 대체.
- **Cron/Heartbeat 서비스 검증 강화**: Webhook 기반 작업 등록과 서비스 연동을 보강.
- **빌드 안정화**: `Makefile`의 `sync-ui` 오류 수정과 저사양 기기 빌드 옵션을 추가.
- **다국어 문서 동기화**: README와 배포 가이드를 최신 버전 기준으로 정리.

## 2026-03-24
### 0.6.1
- **빌드 오류 수정**: Linux 등 일부 플랫폼에서 `getSysProcAttr` 관련 컴파일 오류가 나던 문제 수정.

### 0.5.8
- **공개 저장소 빌드 안정화**: 공개 저장소에서 Web Admin 소스 없이도 기존 빌드 산출물을 사용할 수 있도록 보강.

### 0.5.7
- **데이터 마이그레이션 도구 도입**: `marubot migrate-paths` 명령을 추가하여 기존 경로의 메모리와 작업 데이터를 현재 경로로 이전.

### 0.5.6
- **경로 분석 로직 정리**: `MARUBOT_HOME`과 `--home` 인자를 통한 설정 디렉터리 지정 기능을 개선.

### 0.5.5
- **Windows 서비스 설정 경로 수정**: 서비스 실행 시 설정 디렉터리가 `systemprofile`로 고정되던 문제 수정.
- **GPIO 서버 로직 정리**: 대시보드 GPIO 설정 저장 방식과 중복 코드를 개선.

## 2026-03-23
### 0.5.4
- **표준 배포 워크플로 정리**: `RULES.md` 기준 배포 절차를 적용.
- **채널 간 메시지 전송 도구 추가**: AI가 `send_channel_message` 도구로 채널 간 메시지를 중계할 수 있도록 지원.
- **Slack 스레드 응답 강화**: `app_mention` 이벤트와 스레드 응답 처리를 개선.
- **설정 자동 동기화**: 누락 설정 필드 자동 보정과 오래된 필드 제거를 적용.

### 0.4.91
- **Uninstall 개선**: 설치 경로 외부의 실행 파일 삭제 여부를 사용자에게 확인하도록 변경.

### 0.4.89
- **설정 파일 경로 통일**: 설정 저장 위치를 `~/.marubot/config.json`으로 정리.
- **Windows 업그레이드 안정화**: 업그레이드 관련 플래그와 다이얼로그 UI를 추가.
- **트레이 메뉴 UX 개선**: 업그레이드 확인 및 진행 상태 알림을 보강.

### 0.4.87
- **Slack 연동 개선**: Slack 채널 로깅을 메인 로거와 통합하고 상세 디버그 로그를 추가.

## 2026-03-21
### 0.4.86
- **Web Admin UI 개선**: 설정 페이지 배치, 시스템 언어 선택, Webhook/Slack/WhatsApp 세부 설정을 보강.
- **Uninstall 안정화**: Windows 환경에서 프로세스 종료와 서비스 제거를 강화.
- **다국어 지원**: 신규 설정 항목에 한국어, 영어, 일본어 번역을 적용.
