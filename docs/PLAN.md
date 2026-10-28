# 구현 계획과 기술 설계

## checkpoint

1. 상세 기획·원문 이관·버전 관리 도구 정비.
2. 사용자별 SQLite 저장소, DAG scheduler, 장비·승인·기억·Wiki·감사와 복구 구현 및 테스트.
3. Admin backend 인증·API·SSE, HTTPS worker·heartbeat·정책, marubot 자연어 위임 도구 연결 및 통합 테스트.
4. Admin frontend 전체 현황·작업 작성·장비·승인·기억 관리, 빌드·전체 검증·운영 문서 완성.

각 checkpoint에서 docs/CHANGELOG.md와 docs/TASKS.md를 갱신하고 patch 버전을 올린 뒤 검증·commit·현재 브랜치 push한다. 개발 Wiki 기록은 사용자 지시에 따라 이 세션에서는 docs에 보관한다.

## 구조

- admin/backend: Go Admin 서버 실행 진입점. 기본 localhost:8000, reverse proxy HTTPS 지원.
- admin/frontend: 기존 Go 임베드 배포와 호환되는 독립 운영 화면. 기존 로컬 대시보드는 보존한다.
- pkg/orchestration: 저장소, 타입, DAG 조율, 사용자별 기억, worker 프로토콜과 전송.
- pkg/controlplane: Google OAuth와 세션, 사용자/노드별 권한, Admin API.
- marubot worker: 기존 ToolRegistry를 사용하되 요청 컨텍스트에 도구 allowlist와 제한을 부여한다.

## 저장과 권한

backend root/auth.sqlite에는 세션과 해시된 장비 자격 자료를 둔다. 사용자별 root/users/SHA256(sub)/state.sqlite에는 노드, 작업 DAG, 이벤트, 감사, 기억과 문서 버전을 둔다. Google sub를 파일 경로로 직접 사용하지 않는다. 토큰 원문은 DB·감사·Wiki에 저장하지 않는다.

노드 토큰은 로그인 사용자가 발급하고 90일 만료·폐기를 지원한다. worker는 토큰을 환경 또는 접근 제한된 파일에서 읽는다. TLS 서버 인증을 검증하며 InsecureSkipVerify는 사용하지 않는다. 자체 CA는 명시적으로 전달한다. Google 로그인은 state+PKCE 및 검증된 sub를 사용한다.

## 실행 모델

작업은 id, instruction, target_node_id, tool, args, depends_on, approval_level, deadline, idempotency_key를 가진다. 서버가 ID·소유자를 확정하고 DAG 순환·미존재 의존성·중복 키 충돌을 거부한다. 승인 시 도구와 인자는 저장된 작업에 결합한다. 승인 이후 작업 본문 수정은 허용하지 않는다.

준비된 task를 제한된 병렬도로 실행한다. 실패한 의존 작업의 후속은 blocked가 된다. 취소는 컨텍스트 전파와 worker 취소 API로 처리한다. 프로세스 재시작 시 running은 unknown으로 보존해 부작용 재실행을 막는다. 읽기 작업도 자동 재라우팅하지 않고 사용자가 결과를 확인한 후 새 작업으로 제출한다.

## 기억 lifecycle

short: 만료 필수(기본 24시간), long: 사용자 검토 후 유지. Wiki는 long+검증된 내용이다. private: node_id 또는 session_id 필수, shared: 같은 owner의 장비에서 조회 가능. worker는 shared와 자기 node의 private만 읽으며 다른 세션 기억은 받지 않는다. 승격은 원본과 연결된 새 Wiki revision으로 기록한다. 삭제·갱신에는 revision 검사를 적용한다. 검색은 사용자 DB 내 제목·본문 키워드 검색이다.

## Admin 정보 구조

상단에 온라인 장비/전체 장비, 실행 작업, 승인 대기, 기억 수와 최근 갱신을 표시한다. 장비 목록, 작업 DAG와 결과, 개별 승인, 검색 가능한 기억/Wiki, 감사 이력을 탭으로 제공한다. 사용자가 설정할 값은 인라인 오류를 제공하고 파괴 작업은 대상별 확인한다.
