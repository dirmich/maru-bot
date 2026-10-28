# 변경과 검증 기록

## 0.9.18 — 공개 GitHub Release 준비

- 공개 배포 marubot 실행 파일에 `admin-server` 진입점을 추가해 별도 Admin 파일 없이 사용자별 Wiki·장비·작업 관리 화면을 실행할 수 있게 했다.
- standalone Admin과 marubot은 같은 pkg/adminruntime을 사용한다. 기존 localhost:8000 설정과 환경 로딩은 보존한다.
- 릴리스 실행 파일 검증에서 발견한 토큰 화면 닫기 이벤트의 비동기 지연을 수정해 닫기·Escape 시 토큰을 즉시 제거한다.
- 릴리스 자산은 기존 배포 규칙에 따라 Windows amd64·386 실행 파일이다. 현재 Windows 환경에서 macOS DMG는 생성하지 않는다.
- 검증: 전체 Go 테스트·build·vet, Windows amd64/386 GUI 실행 파일 빌드와 버전 확인, 두 아키텍처 실행 파일의 admin-server Chrome UI smoke를 통과했다.

## 0.9.14 — 상세 기획

- 분산 실행과 사용자별 backend Wiki 요구를 PRD·PLAN·TASKS에 구체화했다.
- 원래 Ochestration.md는 내용 그대로 docs/Orchestration-original.md로 보존한다. 이후 최신 설계 기준은 PRD와 PLAN이다.
- 기존 Go 대시보드는 임베드된 frontend 산출물만 있고 admin/frontend와 admin/backend는 비어 있었다. 신규 Admin은 기존 대시보드를 보존하는 독립 Go 서버로 추가한다.
- 현재 PATH에 Go가 없어 공식 Go 배포본을 checksum 검증 후 로컬 빌드 도구 경로에 설치한다. 문서 checkpoint 검증은 diff와 버전 일치 검사다.

## 0.9.15 — 작업 엔진과 기억 저장소

- 사용자 sub 해시별 SQLite 분리, DAG 병렬 실행과 개별 승인, 기한 만료, 취소, 멱등 키 본문 충돌 검증과 원자적 저장을 구현했다.
- worker HTTPS 실행·조회·취소 프로토콜, 로컬 도구 허용 목록과 실행 journal을 추가했다.
- 단기 기억 만료, 장기/Wiki 검토 승격, private/shared 범위, revision 충돌과 삭제 시 이력 삭제를 구현했다.
- 검증: go test ./pkg/orchestration 통과. 두 작업 동시 실행·의존성·승인 재사용 방지·사용자 격리·재시작 unknown 상태를 확인했다. 실제 장비 연결은 미검증이다.
- 기본 기한을 생성하면서 요청 객체를 변경해 중복 요청 해시가 달라지는 문제를 테스트로 발견하고 깊은 복사로 수정했다.

## 0.9.16 — Admin backend와 worker 연결

- Google OAuth state·PKCE·nonce와 RS256/JWKS 검증, 12시간 HttpOnly 세션, CSRF 검증을 구현했다.
- 장비 토큰은 해시로 인증하고 backend에 필요한 원문은 AES-GCM으로 암호화했다. 90일 만료·개별 갱신·폐기와 15초 heartbeat를 구현했다.
- HTTPS worker, Tailscale 주소 제한, workspace 경로 정책, 원격 재위임 금지와 대화용 위임·기억 도구를 연결했다.
- 사용자별 Admin API·영속 SSE·기억 검토·작업 승인·unknown 작업의 장비 기록 재조회, 자연어 계획 생성을 추가했다.
- 전체 시스템 Admin frontend와 backend를 단일 Go 실행파일에 임베드했다. 화면 사용성 검증은 다음 checkpoint에서 수행한다.
- 검증: go test ./... 통과, Admin과 marubot Windows 바이너리 빌드 통과. 두 실제 HTTPS 테스트 worker의 동시 실행 및 사용자 간 작업·이벤트·기억·장비 접근 차단을 검증했다.
- 실제 Google 자격 자료와 운영 장비 배포는 수행하지 않았다.

## 0.9.17 — Admin 화면 검증과 운영 완성

- desktop/mobile 전체 현황, 장비 수정·일시중지/재개, 작업 계획·개별 승인·결과, 기억 검색·수정·Wiki 승격과 활동 기록 화면을 완성했다.
- 최근 100개 이력과 모든 미완료 run을 표시하고 큰 결과는 상세 조회로 분리했다. scheduler는 미완료 작업만 조회하며 동시 run 수를 제한한다.
- 실행 결과를 장비 전용 단기 기억으로 저장하고 기억 검색·장기 초안 도구를 제공했다. 외부 채널에서 분산 자격 자료를 사용할 때 명시적 소유자 allow_from 검증을 추가했다.
- worker 출력과 오류에서 환경·설정의 비밀을 마스킹하고 TLS 기한·장비 인증 확인·복구 동작을 보완했다.
- go test ./... 및 go vet ./... 통과. 취소·승인 만료·단기 만료·자동 분산·일시중지 제외·서명 검증·SSE 재연결·최근 활동·장비 수정 권한·목록/상세 분리 테스트를 포함한다.
- Windows Admin/marubot 빌드와 Linux amd64/arm64 교차 빌드 성공. 다른 OS에서 바이너리를 실제 실행한 것은 아니다.
- Chrome UI smoke 통과: 로그인 설정 필요 상태, 미인증 API 차단, 현황, 개별 승인, 기억 저장, 장비 수정·일시중지/재개·등록, 토큰 닫기 후 제거, 자연어 2단계 제출, 모바일 가로 넘침 없음, JavaScript 오류 없음. 인증 후 화면에는 테스트용 응답을 주입했고 실제 Google 로그인으로 대체하지 않았다.
- QA 화면은 docs/qa, 모든 신규 기획·API·실행·검증 문서는 docs에 보관했다.
- 기존 admin/backend/.env.example은 과거 PostgreSQL/4000 포트 구성 예제다. 로컬 파일은 보존하고 Git에서 제외했으며 현재 설정은 루트 .env.example로 통일했다.
- man.sh 문법 검사는 통과했다. 시작·상태·중지 복합 검증은 자동 승인 검토가 blocked by policy로 차단했으므로 실행 미검증이다. Admin 바이너리 자체의 시작·HTTP 응답은 UI smoke에서 확인했다.
- 실제 Google OAuth 자격 자료 및 Tailscale 원격 장비 연결·배포는 미수행이다.
