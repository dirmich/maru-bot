# Admin API와 실행 프로토콜

## 인증

사람은 Google OAuth 후 HttpOnly 세션 cookie를 사용한다. 변경 요청은 공개 Admin Origin 및 `/api/me`에서 받은 `X-CSRF-Token`이 필요하다. 장비는 `Authorization: Bearer <node-token>`을 사용하며 토큰에서 owner·node·역할을 결정한다. 본문이나 사용자 헤더의 owner 값을 신뢰하지 않는다.

## Admin 엔드포인트

| 경로 | 기능 | 권한 |
|---|---|---|
| GET /api/me | 현재 사용자와 CSRF | 로그인 사용자 / 장비 |
| POST /api/logout | 세션 폐기 | 사람 |
| GET /api/overview | 전체 현황과 장비·작업 | 사람 |
| GET /api/nodes | 소유 장비 목록 | 사람 / orchestrator |
| POST /api/nodes | 장비 등록과 일회 토큰 표시 | 사람 |
| PATCH /api/nodes/{id} | 장비 이름·HTTPS 주소 수정 | 사람 |
| POST /api/nodes/{id}/pause | `{paused}` 새 작업 일시중지·재개 | 사람 |
| POST /api/nodes/heartbeat | 자신의 플랫폼·도구·자원 보고 | 장비 |
| POST /api/nodes/{id}/rotate | 토큰 갱신, 기존 토큰 즉시 폐기 | 사람 |
| POST /api/nodes/{id}/revoke | 등록 폐기 | 사람 |
| POST /api/plan | 자연어 → 실행 전 검토할 DAG 초안 | 사람 / orchestrator |
| POST /api/runs | DAG 접수, 202와 영속 run 반환 | 사람 / orchestrator |
| GET /api/runs, /api/runs/{id} | 작업 목록 / 결과 | 사람 / orchestrator |
| POST /api/runs/{id}/cancel | 취소 요청 | 사람 / orchestrator |
| POST /api/runs/{id}/reconcile | unknown 작업의 worker 기록 재조회 | 사람 |
| POST /api/runs/{id}/tasks/{task}/approval | `{allow: true 또는 false}` 개별 승인 | 사람 |
| GET /api/events | 최근 500개 활동 또는 SSE | 사람 |
| GET /api/runs/{id}/events | 해당 작업 활동 또는 SSE | 사람 / orchestrator |
| GET /api/memories?q=검색어 | 현재 사용자 기억 검색, 최대 200개 | 사람 / 장비 |
| POST /api/memories | 기억 생성·revision 기반 수정 | 사람 |
| POST /api/memory-proposals | 장비 전용 장기 기억 초안 생성 | 장비 |
| GET /api/memories/{id}/history | 현재 기억의 revision 이력 | 사람 |
| POST /api/memories/{id}/promote | `{revision}` 검토된 Wiki로 승격 | 사람 |
| DELETE /api/memories/{id}?revision=N | 기억과 revision 이력 삭제 | 사람 |

장비 기억 조회에는 자기 node의 private와 동일 owner의 shared만 포함한다. 세션 전용 기억은 일반 장비 검색에 노출하지 않는다. human은 자신의 모든 기억을 관리한다. 기억 삭제는 기억 및 revision에 적용하며 출처인 작업 실행 결과까지 삭제하지는 않는다.

## DAG 예시

```json
{
  "title": "두 장비 점검 후 결과 파일 작성",
  "idempotency_key": "user-request-unique-key",
  "tasks": [
    {"id":"windows","node_id":"등록된 Windows ID","tool":"system_status","args":{},"approval_level":"read-only"},
    {"id":"linux","node_id":"등록된 Linux ID","tool":"system_status","args":{},"approval_level":"read-only"},
    {"id":"write","node_id":"등록된 Linux ID","tool":"write_file","args":{"path":"checked.txt","content":"점검 완료"},"depends_on":["windows","linux"],"approval_level":"local-write"}
  ]
}
```

작업 수는 1~64, 동시에 대기·실행하는 run은 사용자별 64개, 최대 병렬 수는 기본 4, 기한은 기본 1시간·최대 24시간이다. 노드 지명은 강제 제약이다. 자동 선택 시 온라인·프로토콜·도구를 먼저 확인하고 실행 개수·CPU·메모리로 분산한다. 실행 인자는 승인 후 수정하지 않는다. shell 등 쓰기 가능한 도구는 read-only 요청을 받아도 승인 대기로 전환한다.

작업 목록은 모든 진행 중 run과 최근 100개 run을 반환하며 큰 도구 출력은 생략한다. 개별 run 조회에는 전체 결과가 포함된다. 오래된 run도 ID로 조회할 수 있다. scheduler는 미완료 run만 읽어 완료 이력이 늘어나도 매초 전체 결과를 다시 읽지 않는다.

`depends_on`은 실행 순서를 지정한다. 앞 단계 결과를 다음 단계 인자에 자동 문자열 삽입하지 않는다. 결과 종합은 대화 모델의 `orchestration_status` 조회 또는 Admin의 장비별 결과 보기로 수행한다.

요청 멱등 키와 정규 JSON 본문 해시를 원자적으로 저장한다. 같은 키의 다른 본문은 거부하고, 같은 본문은 기존 run을 반환한다. 요청 본문 크기는 256KiB, 도구 출력은 64KiB로 제한한다. 도구 출력은 신뢰하지 않는 참고 데이터다.

## worker HTTPS

| 경로 | 기능 |
|---|---|
| POST /v1/tasks | protocol, owner_id, run_id, task, deadline, idempotency_key를 검증하고 실행 |
| GET /v1/tasks/{run_id}-{task_id} | 영속 실행 journal 조회 |
| POST /v1/tasks/{run_id}-{task_id}/cancel | 안전한 중단 요청 |

worker는 등록 시 발급받은 장비별 토큰만 받으며 모든 연결은 TLS 서버 인증서를 검증한다. 프로토콜 버전은 1이다. idempotency_key는 `run_id/task_id`이고 헤더 `Idempotency-Key`와 동일해야 한다. 로컬 정책은 도구 allowlist, workspace 경로, 메타 도구 금지를 검증한다.

worker 실행 응답은 `{status, output, error}`다. 긴 작업은 HTTP 응답을 기다리며, coordinator의 진행 이벤트는 영속 SSE로 제공한다. 원시 도구 로그를 SSE로 보내지 않는다. 프로세스·네트워크 장애로 결과를 모르면 unknown으로 남기고 장비 기록 재조회만 허용한다. 다른 장비에서 자동 재실행하지 않는다.

## 이벤트와 복구

SSE 요청은 `Accept: text/event-stream`으로 보낸다. 각 이벤트는 증가하는 `id`, `event: update`, JSON `data`를 가진다. 재연결 시 `Last-Event-ID` 이후 이벤트부터 전송한다. 한 번에 최대 500개를 읽고 이후 cursor부터 계속 전송한다. 서버는 연결을 최대 5분 유지한 뒤 재연결하도록 한다. 작업 기한과 SSE 연결 수명은 별개다. UI는 15초 보조 갱신도 수행한다.

이벤트는 현재 삭제 없이 영속 보존한다. 보존량이 커지면 일관된 백업 이후 별도 운영 정책으로 정리해야 한다. 자동 결과 파일 청크 전송과 marumesh는 후속 확장이다.
