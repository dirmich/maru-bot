# 📋 MaruBot 구현 작업 리스트 (Tasks)

본 문서는 MaruBot 프로젝트의 단계별 구현 상태를 추적합니다. 작업 완료 시 상태를 업데이트하고 관련 내용을 커밋합니다.

---

## 🟢 Phase 1: 기반 구축 및 기본 인터랙션 (완료 및 검증)
- [x] PicoClaw 소스 코드 독립화 및 모듈 리팩토링 (`marubot` 모듈)
- [x] Raspberry Pi 하드웨어 설정 자동화 스크립트 작성 (`setup-rpi.sh`)
- [x] CSI/USB 카메라 통합 지원 도구 구현 (`camera_capture`)
- [x] 기본 설정 템플릿 및 프로젝트 구조 수립
- [x] 원클릭 설치 스크립트 (`install.sh`) 구현

---

## Phase 2: 모빌리티 및 공간 인지 (완료)
- [x] **Task 2.1: PWM 기반 모터 제어 도구 구현**
    - [x] `periph.io` 기반 PWM 라이브러리 연동
    - [x] `move_forward`, `stop`, `set_speed` 등 기본 이동 API 구현
    - [x] 하드웨어 연결 가이드 작성 (README 및 코멘트 반영)
- [x] **Task 2.2: 초음파 센서(HC-SR04) 장애물 감지 도구 구현**
    - [x] 에코/트리거 핀 제어 로직 구현
    - [x] `get_distance` 도구 추가
- [x] **Task 2.3: IMU(MPU6050) 가속도/자이로 센서 연동**
    - [x] I2C 통신 기반 센서 데이터 읽기
    - [x] 로봇의 기울기 및 방향(Bearing) 계산 도구 구현
- [x] **Task 2.4: 시각 기반 위치 추정 및 추적(Vision Tracking)**
    - [x] 표준 라이브러리 기반 경량 이미지 프로세싱 구현
    - [x] 특정 색상 추적 알고리즘 구현 (`track_color` 도구)

---

## 🟢 Phase 3: 공중 기동 및 자율 주행 확장 (완료)
- [x] **Task 3.1: MAVLink 브리지 구현**
    - [x] FC(Flight Controller)와의 시리얼 통신 연동 (Gomavlib v3 기반)
    - [x] 기본 비행 명령(이륙, 착륙, 정지비행) 도구 구현 (`drone_control`)
- [x] **Task 3.2: GPS 및 절대 좌표 기반 이동 구현**
    - [x] NMEA 데이터 파싱 및 현재 위치 도구 구현 (`get_location`)
    - [x] Waypoint 기반 경로 비행 API 추가 (`drone_control` 내 `goto` 명령)
- [x] **Task 3.3: 자율 주행 및 긴급 복구 로직**
    - [x] 자동 복귀(RTL) 및 긴급 강제 착륙(Emergency Stop) 도구 구현
    - [x] 비상 시 하드웨어 제어 인터페이스 구축

---

## 🟢 Phase 4: 자기 진화 및 동적 확장 (완료)
- [x] **Task 4.1: 실시간 도구 생성 시스템 구축**
    - [x] AI가 스스로 스크립트를 작성하고 도구로 등록하는 `create_custom_tool` 도구 구현
    - [x] `/extensions` 디렉토리를 통한 동적 도구 로드 메커니즘 구축
    - [x] 무중단 기능 확장 및 런타임 도구 업데이트 검증

---

## 🟢 Phase 5: 맞춤형 설정 및 사용자 편의성 (완료)
- [x] **Task 5.1: 설정 파일 분리 및 관리 도구 구현**
    - [x] 사용자 설정을 기록하는 `usersetting.json` 및 오버라이드 시스템 구축
    - [x] 에이전트 내에서 설정을 조회하고 수정하는 `config` 도구 추가
    - [x] 명령줄 인터페이스(CLI)용 `marubot config` 명령어 구현
- [x] **Task 5.2: 공개 레포지토리 배포 자동화 구축**
    - [x] 비공개 환경에서 공개용 파일만 선별하여 동기화하는 `make public` 타겟 구현
    - [x] 공개 배포 시 모든 명칭 자동 치환 (`marubot` -> `marubot`) 엔진 구축
    - [x] MIT 라이선스 파일 추가 및 공개 배포 가이드 수립

---

## ✅ 작업 완료 로그 (Log)
- **2026-02-10**: 프로젝트 독립화, 리팩토링 및 카메라 통합 도구 구현 완료. (Step 1~4 완료)
- **2026-02-11**: GitHub Gist 기반 설치 스크립트 적용 및 설치 가이드 개선.
- **2026-02-11**: 사용자 설정 분리(`usersetting.json`) 및 `config` 관리 도구 구현 완료. (Phase 5)
- **2026-02-11**: 공개 배포 자동화(`make public`) 및 전면 명칭 치환 프로세스 완성.
- **2026-02-13**: 배포 스크립트(`publish.sh`)의 `.git` 폴더 보존 로직 개선 (Windows 권한 문제로 이동 대신 삭제 예외 처리).
- **2026-02-13**: 설치 스크립트(`install.sh`) 개선 (Bun 설치 실패 시 Node.js/NPM으로 자동 전환).
- **2026-02-13**: Web Admin 배포 전략 변경 (로컬 빌드 -> Standalone Artifact 배포), 타겟 머신 빌드 부하 제거.
- **2026-02-18**: 설치 구조 개편 (바이너리 -> `/usr/local/bin`, 리소스 -> `~/.marubot`). 설치 후 소스 폴더 제거 가능.
- **2026-02-18**: 설치 완료 후 소스 폴더 자동 삭제 로직 추가 (Clean Install).
- **2026-02-18**: `marubot uninstall` 커맨드 구현 (자체 삭제 및 리소스 정리 기능).
- **2026-02-18**: Web Admin 빌드 시 favicon.ico 디코딩 에러 수정 및 배포 완료.
- **2026-02-18**: README 다국어 문서(EN, KO, JP, CN) 최신화 및 `marubot config` 안내 추가.
- **2026-02-18**: `maru-setup.sh` 내 오타 수정 및 불필요한 스크립트 참조 제거.
- **2026-02-18**: 32-bit ARM 환경 지원을 위해 Prisma를 Drizzle ORM으로 전면 교체. 이제 32-bit OS에서도 대시보드 사용 가능.
- **2026-02-18**: MaruBot v0.2.0 정식 배포 (Github: dirmich/maru-bot). 32-bit/64-bit 자동 감지 설치 및 Web Admin Standalone 빌드 적용 완료.
- **2026-02-19**: 32-bit 환경 빌드 에러 수정 (next.config.mjs 적용).
- **2026-02-19**: 대시보드 및 설치 스크립트 다국어화(EN, KO, JP) 완료. 자기 진화(Auto-Evolution) 시스템 프롬프트 강화.


- **2026-02-19**: MaruBot v0.3.0 정식 배포. Vite SPA 마이그레이션, 다국어 내비게이션 및 대시보드 임베딩 완료.
- **2026-02-19**: 대시보드 모드(`marubot dashboard`) 실행 시 백그라운드 서비스(Cron, Heartbeat)가 누락되는 문제 해결. 이제 대시보드에서도 예약 작업 및 자동화 기능이 정상 동작함.
- **2026-02-20**: MaruBot v0.3.2 업데이트. 원격 버전 확인 기반 스마트 업그레이드(`upgrade`) 기능 및 CLI 도움말 정렬 완료.


## 🟢 Phase 6: 웹 관리자 경량화 및 의존성 제거 (완료)
- [x] **Task 6.1: Vite + React (SPA) 마이그레이션**
    - [x] Next.js 의존성 제거 및 Vite/React Router/Tailwind 설정 (package.json, vite.config.ts)
    - [x] 진입점(src/main.tsx, index.html) 및 라우팅 구성 (App.tsx)
    - [x] 컴포넌트 및 페이지 구조 이관 (`app/` -> `src/pages`, `src/components`)
    - [x] API 연동 로직 수정 (Next.js API Routes -> Go 백엔드 직접 호출 가능한 React 로직으로 변경)
- [x] **Task 6.2: Go 바이너리 내장 (Embed) 및 서빙**
    - [x] `web-admin/dist` 빌드 결과물을 Go 바이너리에 임베딩 (`//go:embed`)
    - [x] Go 서버(`dashboard/server.go`) 구현: 정적 파일 서빙 및 기본 API 구조
    - [x] `marubot dashboard` 명령 실행 시 Go 서버 구동 연결
    - [x] API 라우팅과 정적 파일 서빙 라우팅 분리 및 연동 테스트 (Go Build 성공)
- [x] **Task 6.3: 설치 프로세스 최적화**
    - [x] `install.sh`: 사용자 기기(RPi)에서의 `npm install`, `npm run build` 단계를 Go 빌드 전 단계로 이동
    - [x] 미리 빌드된 자산(Web Admin)이 포함된 단일 바이너리 배포 구조로 변경
    - [x] `marubot dashboard` 실행 시 별도 Node.js/Bun 프로세스 없이 Go 서버만으로 동작하도록 개선

---

## 🟢 Phase 7: 다국어화 및 지능형 확장 (완료)
- [x] **Task 7.1: 대시보드 UI 다국어 지원**
    - [x] `zustand` + `localStorage` 기반 언어 상태 관리 및 영구 저장 구현
    - [x] 한국어, 영어, 일본어 번역 사전(`i18n.ts`) 구축
    - [x] 모든 대시보드 페이지(Chat, GPIO, Skills, Settings) 번역 적용
- [x] **Task 7.2: 원클릭 설치 스크립트(`install.sh`) 다국어화**
    - [x] 설치 시작 시 언어 선택 프롬프트 추가
    - [x] 모든 설치 메시지 및 로그 다국어 출력 적용
    - [x] 선택된 언어를 시스템 기본 언어로 `config.json`에 자동 주입
- [x] **Task 7.3: 자기 진화(Auto-Evolution) 기능 강화**
    - [x] `create_custom_tool` 도구의 목적과 사용법을 시스템 프롬프트에 명시
    - [x] 에이전트가 스스로 기능을 확장하도록 유도하는 지능형 컨텍스트 구축

---

## 🟢 Phase 8: 배포 및 안정화 (완료)
- [x] **Task 8.1: 버전 업그레이드 (v0.3.0)**
    - [x] `main.go`, `package.json`, `IDENTITY.md` 버전 동기화
- [x] **Task 8.2: 공개 레포지토리 배포**
    - [x] `publish.sh`를 통한 `maru-bot` 배포 자산 동기화 및 명칭 치환
    - [x] README 다국어 정비 (EN 메인, KO 별도 보관)

---

## 🟢 Phase 9: 업그레이드 시스템 강화 및 최적화 (완료)
- [x] **Task 9.1: 원격 버전 확인 및 스마트 업그레이드**
    - [x] GitHub 원격지의 `main.go`에서 최신 버전 파싱 로직 구현 (`getLatestVersion`)
    - [x] 현재 버전과 비교하여 업데이트 필요 여부 판단 및 사용자 확인 인터페이스 추가
    - [x] 업그레이드 전 기존 프로세스 자동 중단(`stopCmd`) 연동
- [x] **Task 9.2: CLI 사용자 경험 개선**
    - [x] 모든 명령어 도움말(`help`)의 서브 명령어 알파벳 순 정렬 (`cron`, `skills`, `config`)
    - [x] Task 10.1: 웹 대시보드 관리자 인증 시스템
- [x] Task 10.2: 메신저 채널 설정 인터페이스 (Web Admin)
- [x] Task 10.3: 업그레이드 안정성 개선 (v0.3.5)

---

## 🟢 Phase 11: 시스템 인터랙션 및 설치 복구 강화 (완료)
- [x] **Task 11.1: 시스템 정보 및 명령어 처리 지원 강화**
    - [x] `shell` 도구 명칭 및 설명 보완 (시스템 상태, IP 조회 등 지원 명시)
    - [x] 시스템 프롬프트 가이드 추가 (리눅스 시스템 명령어 처리 가능성 LLM에 명시)
- [x] **Task 11.2: 설치 및 업데이트 안정성 보완**
    - [x] `install.sh`: `git pull` 실패 시(bad object 등) 기존 폴더 삭제 후 재클론(Fresh Clone) 로직 추가
- [x] MaruBot v0.3.6 배포 완료

---

## 🟢 Phase 12: 사용자 경험 및 진단 시스템 고도화 (완료)
- [x] **Task 12.1: 메신저 인터랙션 개선 (입력 중 상태 표시)**
    - [x] Telegram, Discord 등 주요 채널에서 AI 처리 중 'Typing...' 표시 기능 구현
    - [x] `OutboundMessage` 에 액션 필드 추가 및 채널별 핸들러 적용
- [x] **Task 12.2: 대시보드 진단 및 로그 시스템 구축**
    - [x] `~/.marubot/dashboard.log` 실시간 로그 관리자 대시보드 내 로그 뷰어(`LogsPage`) 추가
    - [x] 시스템 프롬프트 보안 가이드 수정 (사용자 진단 목적의 IP/시스템 정보 제공 허용)
- [x] **Task 12.3: MaruBot v0.3.8 배포 완료**

---

## 🟢 Phase 13: 대시보드 홈 및 시스템 리소스 모니터링 (완료)
- [x] **Task 13.1: 시스템 대시보드(Home) 구현**
    - [x] CPU 사용량, 메모리 상태, 디스크 잔량 등 시각화
    - [x] 가동 시간(Uptime) 및 시스템 정보 요약 표시
- [x] **Task 13.2: 실시간 시스템 정보 API 구현**
    - [x] Go 백엔드에서 `/api/system/stats` 엔드포인트 구현
    - [x] RPi 가동 정보 및 네트워크 상태 제공 로직 추가

---

## 🟢 Phase 14: 인터랙션 강화 및 시스템 명령어 복구 (완료)
- [x] **Task 14.1: 텔레그램 입력 지표(Typing Indicator) 개선**
    - [x] 고루틴 기반 주기적 typing 액션 전송 (생각 중 상태 유지)
- [x] **Task 14.2: 크로스 플랫폼 쉘 명령어 지원 (Windows/Linux)**
    - [x] Windows 환경에서 `cmd /c` 사용 지원 추가
    - [x] 플랫폼별 시스템 정보 확인 명령어 가이드 보강 (System Prompt)
- [x] MaruBot v0.3.10 배포 완료

---

## 🟢 Phase 15: Webhook 채널 및 외부 인터랙션 (완료)
- [x] **Task 15.1: Webhook 채널 구현 (`pkg/channels/webhook.go`)**
    - [x] HTTP POST 요청 수신 및 메시지 처리 로직 구현
    - [x] 동기적 응답 반환 기능 (Pending Response Map) 구현
- [x] **Task 15.2: 채널 매니저 및 설정 통합**
    - [x] `manager.go`에 Webhook 채널 초기화 로직 추가
    - [x] `config.go`에 Webhook 설정 필드 추가

---

## 🟢 Phase 16: 자기 진화 및 시스템 자가 관리 (완료)
- [x] **Task 16.1: 시스템 제어 도구 구현 (`pkg/tools/system.go`)**
    - [x] 마루봇 재시작(Reload) 기능 구현
    - [x] 기술(Skills) 설치 및 목록 조회 기능 구현
- [x] **Task 16.2: 에이전트 자가 관리 역량 강화**
    - [x] `AgentLoop`에 `SystemTool` 등록
    - [x] 에이전트가 도구/기술 설치 후 스스로 상태를 갱신하도록 가이드 제공

---

## ✅ 작업 완료 로그 (Log)

... (생략) ...
- **2026-02-21**: MaruBot v0.4.0 업데이트. Webhook 채널 추가(동기 응답 지원) 및 시스템 제어 도구(`system_control`)를 통한 AI 자가 진화(기술 설치, 재로딩) 역량 강화.
- [x] 2026-02-21: Webhook 및 자가 진화 기능 구현 완료 및 정식 배포 준비.
