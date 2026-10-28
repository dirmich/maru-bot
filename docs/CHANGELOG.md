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
