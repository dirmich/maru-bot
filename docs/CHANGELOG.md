# 변경과 검증 기록

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
