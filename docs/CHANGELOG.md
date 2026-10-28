# 변경과 검증 기록

## 0.9.14 — 상세 기획

- 분산 실행과 사용자별 backend Wiki 요구를 PRD·PLAN·TASKS에 구체화했다.
- 원래 Ochestration.md는 내용 그대로 docs/Orchestration-original.md로 보존한다. 이후 최신 설계 기준은 PRD와 PLAN이다.
- 기존 Go 대시보드는 임베드된 frontend 산출물만 있고 admin/frontend와 admin/backend는 비어 있었다. 신규 Admin은 기존 대시보드를 보존하는 독립 Go 서버로 추가한다.
- 현재 PATH에 Go가 없어 공식 Go 배포본을 checksum 검증 후 로컬 빌드 도구 경로에 설치한다. 문서 checkpoint 검증은 diff와 버전 일치 검사다.
