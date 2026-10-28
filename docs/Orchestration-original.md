marubot 오케스트레이터 설계 — Tailscale 1단계, marumesh 교체 가능
작성일: 2026-10-28
상태: 설계 문서 (구현 미시작, marubot 소스 수정 없음)
결정: 2026-10-28 사용자 선택 — 1단계 전송은 Tailscale, marumesh는 나중에 교체 가능하게 설계한다. marumesh는 아직 완성된 상태가 아니라는 사용자 확인을 전제로 한다.
결정: 2026-10-28 사용자 선택 — 에이전트 간 와이어 프로토콜은 HTTP+JSON+SSE로 확정한다 (15절).
전임 분석: 2026-10-28 marumesh 분석 — private dirmich/marumesh가 소스 정본(Go 1.26.1, 2026-09-22 기준 v0.11.95), public dirmich/maru-mesh는 배포·문서용(2026-09-18 기준 v0.11.89, 루트에 소스 없음).
1. 목표와 비목표
목표
사용자는 오케스트레이터 1명하고만 대화하고, 오케스트레이터가 여러 기기의 marubot을 조율해 목적을 달성한다.
marubot의 기존 강점을 유지한다: Go 단일 바이너리, RPi/Windows/macOS/Linux 멀티플랫폼, 로컬 도구 직접 실행.
전송 수단은 교체 가능하게 한다. 1단계는 검증된 Tailscale, 나중에는 marumesh로 밑단만 바꾼다. 태스크 프로토콜, 능력 등록, 인증, 감사 형식은 전송과 무관하게 유지한다.
비목표
OpenBot 전체 스택(TS/Bun/React/Hono/PostgreSQL/CopilotKit Intelligence)을 marubot에 이식하지 않는다.
다중 테넌트, 팀 역할 관리, 공개 노출을 만들지 않는다. 단일 소유자 전제이다.
오케스트레이터가 원격 기기에 직접 셸 명령을 날리는 구조를 만들지 않는다. 작업은 항상 해당 기기의 에이전트에게 위임하고, 그 에이전트가 자기 정책으로 실행한다.
2. 전체 구조
사용자 (Telegram / Discord / Slack / Web Admin)
        │
        ▼
오케스트레이터 marubot (role=orchestrator, 권장 배치: main 서버)
  - 계획 수립, 라우팅, 결과 집계, 승인 취합, 감사 기록
        │  Transport 인터페이스 (아래 4절)
        ├─ Tailscale 1단계 ──────────────┐
        └─ marumesh (후기 단계, 교체) ──┘
        │
        ▼
worker marubot (각 기기, role=worker)
  - 자기 기기의 도구만 실행: shell, browser(chromedp), gpio, camera,
    filesystem, ssh, system 등 (pkg/tools)
  - 자기 로컬 정책으로 거부 가능, 비밀은 외부로 내보내지 않음
오케스트레이터도 worker도 같은 marubot 바이너리이다. 역할은 설정으로만 구분한다. 별도 프로그램이 아니다.
오케스트레이터 권장 배치: 상시 가동되는 main(Linux 서버). window-office, macmini, mbp-home/office, dev, pgx, RPi 노드는 worker로 둔다. 오케스트레이터 호스트는 구현 전 최종 확인 사항이다.
사용자는 필요 시 특정 노드를 지명할 수 있다 ("이번 건은 window-office에서 해줘"). 지명이 없으면 오케스트레이터가 능력·온라인 상태·부하로 라우팅한다.
3. 신원: Google SSO로 동일 소유자에 묶기
3.1 원칙
소유자 식별자는 이메일이 아니라 Google 고유 사용자 ID(sub) 기준으로 한다. 이메일은 변경·재사용될 수 있다.
모든 노드(오케스트레이터·worker)는 같은 소유자에 속해야 통신한다. 소유자가 다르면 네트워크가 닿아도 작업이 거부된다.
인간 로그인(SSO)과 노드 간 인증은 분리한다. 이것이 이 설계에서 가장 중요한 구분이다.
3.2 등록(Enrollment) 흐름
노드를 처음 켤 때 사람이 Google SSO로 로그인한다.
오케스트레이터가 해당 노드의 등록을 승인한다. 승인 전에는 어떤 작업도 받지 않는다.
승인된 노드에만 노드별 인증 자료를 발급한다. 노드마다 다르고, 만료·갱신·폐기가 가능해야 한다.
노드 삭제·소유자 변경 시 오케스트레이터에서 해당 노드 인증을 폐기하고, 세션을 종료한다.
3.3 실행 시 인증 자료
SSO 토큰 자체를 매 작업마다 쓰지 않는다. SSO 토큰은 수명이 짧고, 무인 노드(RPi, 서버)는 사람이 재로그인할 수 없기 때문이다.
실행 시에는 등록 단계에서 발급한 노드 인증을 쓴다. 권장 형태는 오케스트레이터를 내부 CA로 하는 mTLS 클라이언트 인증서, 또는 노드별 범위 제한 API 토큰 중 하나이다. 둘 중 하나를 구현 전 확정한다.
인증 자료는 로컬에 암호화 저장하고, 로그·감사 기록·작업 결과에 절대 포함하지 않는다. 평문으로 설정 파일에 남기지 않는다.
만료가 가까워지면 노드가 자동 갱신을 시도하고, 갱신 실패 시 해당 노드는 신규 작업을 받지 않고 오케스트레이터에 상태를 보고한다.
3.4 worker의 검증 체크리스트 (매 요청)
소유자(owner_id) 일치 여부
발신자가 등록된 오케스트레이터 노드인지 여부
작업 범위(scope)와 도구 허용 목록(tool allowlist) 준수 여부
인증 만료 여부, 작업 만료(deadline) 여부
하나라도 어긋나면 실행하지 않고 거부 사유를 감사 로그에 남긴다.
3.5 marumesh와의 신원 매핑 (후기 단계 참고)
marumesh는 이미 Google SSO 기기 등록, 소유자별 기기명 유일성, 대시보드 정책·그룹을 갖추고 있다. marumesh로 교체할 때는 marumesh의 소유자·기기 신원을 marubot의 owner_id·node_id에 매핑한다.
단, 메시 네트워크 신원을 그대로 marubot 작업 인증으로 쓰지 않는다. marubot 작업마다 위 3.4 검증을 별도로 수행한다.
4. 전송 추상화: Tailscale 1단계, marumesh 교체
4.1 인터페이스 (Go, 개념)
// Transport는 "어디로"가 아니라 "어떻게 닿는지"만 담당한다.
// 태스크 형식, 인증, 능력, 감사는 이 위 계층에서 동일하게 유지한다.
type Transport interface {
    // 노드 식별자(nodeID)로 다이얼한다. IP/DNS 문자열을 직접 다루지 않는다.
    Dial(ctx context.Context, nodeID string) (Conn, error)
    Listen(ctx context.Context) (Listener, error)
    // 노드 주소가 바뀌어도 호출자는 nodeID만 알면 된다.
    Resolve(nodeID string) (Addr, error)
    Close() error
}

type TailscaleTransport struct{ /* tailnet 전용 */ }
type MaruMeshTransport struct{ /* 후기 단계 */ }
설정은 전송 종류만 바꾼다. 예: mesh.transport = "tailscale" | "marumesh".
오케스트레이터·worker 코드, 태스크 JSON, 감사 스키마는 전송 종류에 따라 바뀌지 않는다. 이것이 교체 가능성의 전부이다.
4.2 1단계: TailscaleTransport
이미 David 인프라에서 검증된 tailnet을 쓴다. 공개 포트를 열지 않는다.
worker API는 Tailscale IP(100.x) 및 loopback에서만 수신한다. 공인 IP·LAN 전체 바인딩은 하지 않는다.
노드 식별은 tailnet 이름(window-office, macmini, main, dev, pgx, mbp-home, mbp-office 등)을 nodeID에 매핑하는 것부터 시작한다. 단, tailnet 이름과 marubot nodeID는 같은 값이 아닐 수 있으니 등록 시 명시적으로 매핑 테이블을 만든다.
Tailscale ACL·태그로 marubot 노드 간 통신 범위를 제한하는 것은 운영상 권장이나, 최종 보안은 3.4의 marubot 자체 검증이 담당한다. 네트워크 도달 가능성만으로 신뢰하지 않는다.
주의: Tailscale 계정·컨트롤 플레인 의존, DERP 경유 시 대용량 파일 전송 지연은 1단계의 알려진 제약이다. 대용량 결과는 요약·링크·청크 전송으로 나누는 것을 전제로 한다.
4.3 후기 단계: MaruMeshTransport (marumesh 완성 후)
marumesh 분석 결과 기반이다. userspace WireGuard over ICE, 가상 IP 기본 100.64.0.0/24, MagicDNS .maru 존(<device>.maru), marumesh connect <device> --remote-port 프록시, 로컬 소켓 ~/.marumesh/marumesh.sock이 있다.
교체 시 구현할 것:
MaruMeshTransport가 위 인터페이스를 구현한다. Dial은 marumesh 가상 IP 또는 .maru 이름으로 연결한다.
marubot nodeID ↔ marumesh device_id 매핑 테이블을 만든다. marumesh 기기명은 소유자별 유일하다는 점을 활용한다.
marubot이 피어 목록을 얻는 접점을 확정한다. 후보는 marumesh ctl list/status, 대시보드 API, 로컬 소켓 중 하나이며, 코드 수준 확인이 필요하다. 이 항목은 구현 전 미확인 사항으로 남긴다.
교체 절차 (롤백 포함):
marumesh 배포본을 private 최신과 동기화한다 (2026-09 기준 public v0.11.89 < private v0.11.95 drift 확인됨).
컨트롤 플레인을 직접 띄우거나 기본 https://marumesh.lab.highmaru.com 동작을 검증한다.
노드 1대만 marumesh 전송으로 카나리 전환하고, 동일 태스크를 Tailscale 경로와 비교 검증한다.
노드별로 순차 전환하고, 문제 시 해당 노드만 Tailscale로 되돌린다. 전체 일괄 전환은 하지 않는다.
전 노드 전환 완료 후 TailscaleTransport 코드를 삭제하지 않고 최소 1개 릴리스 동안 유지한다.
marumesh 미완성 전제: 위 후기 단계는 marumesh가 완성·안정화된 뒤에 시작한다. 그 전까지는 marumesh를 전송으로 쓰지 않는다.
5. 능력 등록과 생존 확인
worker는 주기적으로 heartbeat를 오케스트레이터에 보낸다.
NodeInfo (등록 시 + heartbeat 갱신):
{
  "node_id": "main",
  "owner_id": "google-sub-hash",
  "platform": "linux",
  "arch": "amd64",
  "marubot_version": "0.9.13",
  "tools": ["shell", "browser", "filesystem", "docker"],
  "roles": ["build", "docker-host"],
  "location": "server-room",
  "load": 0.2,
  "last_seen": "2026-10-28T11:00:00+09:00"
}
heartbeat가 일정 시간(구현 시 확정, 예: 3회 연속 누락) 없으면 해당 노드는 오프라인으로 표시하고 신규 라우팅에서 제외한다. 진행 중 작업은 타임아웃 후 재라우팅 여부를 판단한다.
버전 호환성: 오케스트레이터와 worker의 태스크 프로토콜 버전이 다르면, 해당 노드는 구버전 작업만 받거나 라우팅에서 제외한다. 무조건 실행하지 않는다.
6. 태스크 모델
6.1 작업 봉투 (Task Envelope)
{
  "task_id": "task-20261028-001",
  "parent_task_id": null,
  "initiator": "person",
  "owner_id": "google-sub-hash",
  "target_node_id": "window-office",
  "scope": "research-and-summarize",
  "instruction": "KBO 2026 포스트시즌 일정 조사 후 요약",
  "inputs": {"urls": ["https://example.com"]},
  "expected_output": {"type": "markdown", "max_chars": 4000},
  "approval_level": "read-only",
  "deadline": "2026-10-28T12:00:00+09:00",
  "idempotency_key": "task-20261028-001-v1",
  "retry_limit": 2
}
initiator는 person | routine | handoff | memory 중 하나이다. 무인 실행(cron/heartbeat)과 사람이 직접 시킨 일을 구분한다. 이는 OpenBot의 initiator 구분 아이디어를 marubot에 맞게 축소한 것이다.
approval_level 예: read-only | local-write | external-send | destructive. destructive는 항목별 확인이 필요하다 (사용자 기존 선호).
멱등성: 같은 idempotency_key로 재전송되면 worker는 이미 수행한 결과를 반환하거나, 수행 중이면 중복 실행하지 않는다. 셸·GPIO 같은 부작용 있는 도구에 특히 중요하다.
6.2 DAG와 상태
오케스트레이터는 작업을 순차·병렬·조건 분기가 있는 작은 DAG로 나눈다. 복잡한 워크플로 엔진이 아니라, 의존성이 있는 subtask 목록 수준으로 시작한다.
태스크 상태: queued | running | need-approval | succeeded | failed | cancelled | rerouted.
결과는 항상 구조화된 봉투로 돌아온다: {task_id, status, output, tool_calls[], provenance, error}.
provenance에는 실행 노드, 사용 도구, 참조 소스가 들어간다. 사용자가 최종 답에서 "어느 노드가 무엇을 했는지"를 확인할 수 있다.
6.3 스트리밍과 진행 상황
worker는 시작·진행·도구 호출·승인 필요·완료·실패 이벤트를 스트리밍한다. 오케스트레이터는 이를 사용자 대화에 요약해 보여준다.
세부 로그는 기본적으로 접어두고, 사용자가 펼칠 때만 보여준다. 모든 노드의 원시 로그를 그대로 쏟아내지 않는다.
7. 오케스트레이터 동작
사용자 요청을 받으면, 먼저 계획을 만든다: subtask 목록, 의존성, 각 subtask에 필요한 능력.
라우팅 점수: 능력 일치 > 온라인 여부 > 부하 > 지역성(파일·하드웨어가 있는 노드 우선) > 사용자 지명 순으로 정한다. 점수 규칙은 설정으로 바꿀 수 있게 한다.
병렬 가능한 subtask는 병렬로, 의존성이 있는 것은 순차로 실행한다.
결과를 모아 최종 답을 합성한다. 합성 시 각 subtask 결과의 출처를 유지한다.
실패 시: 재시도 제한 내에서 재시도 → 동일 능력을 가진 다른 노드로 재라우팅 → 둘 다 안 되면 멈추고 사용자에게 묻는다. 무한 재시도는 하지 않는다.
오프라인 노드가 복귀하면, 대기 중이던 작업을 자동 재개하지 않는다. 오케스트레이터가 상태를 확인한 뒤 사용자가 재개를 승인하거나, 읽기 전용 작업만 자동 재개하도록 정책을 둔다.
8. worker 동작과 로컬 정책 게이트
worker는 오케스트레이터 요청을 받아도, 자기 로컬 정책 게이트를 통과한 도구만 실행한다.
기존 marubot 구조에서 게이트 위치는 pkg/tools의 ToolRegistry.Execute() 앞단이다. 이 지점에서 정책 평가 → 감사 기록 → 실행 → 결과 감사 기록 순서를 넣는다. 이는 OpenBot 게이트웨이 아이디어를 Go 네이티브로 축소한 것이다.
정책 규칙은 처음부터 복잡한 CEL 엔진을 쓰지 않는다. tool.name, 대상 호스트, 명령, 파일 경로, initiator, approval_level 조건의 deny-before-allow 규칙으로 시작한다. 정책이 없거나 깨지면 닫힌 쪽(fail-closed)으로 처리한다.
worker는 로컬 비밀(API 키, 로그인 토큰, 브라우저 프로필 내용)을 결과에 포함하지 않는다. 결과에 비밀이 섞이면 마스킹하거나 해당 필드를 제외한다.
worker는 다른 노드의 도구 출력을 지시로 받지 않는다. 도구 출력은 untrusted 태그를 달고, 그 안에 있는 지시문은 실행하지 않는다.
9. 승인, 감사, UI
9.1 승인
오케스트레이터가 승인을 취합한다. 사용자는 오케스트레이터 대화에서 한 번에 답하고, 오케스트레이터가 해당 subtask의 worker에 승인 결과를 전달한다.
파괴적 작업(삭제·정리·외부 발송·결제 등)은 항목별로 하나씩 확인받는다. 묶음 승인은 하지 않는다. 이는 사용자의 기존 파괴 작업 선호와 일치한다.
승인 대기 중인 작업은 타임아웃을 둔다. 타임아웃 시 자동 승인하지 않고 취소 또는 보류한다.
9.2 감사 로그
기존 marubot의 SQLite(modernc.org/sqlite)를 그대로 쓴다. 별도 DB를 도입하지 않는다.
감사 행 필드: time, task_id, parent_task_id, initiator, owner_id, node_id, tool, target, decision(allow|deny|fail), rule_id, approval_id, duration_ms, error.
감사 기록에는 비밀·토큰·비밀번호를 남기지 않는다. 필요 시 길이만 남긴다.
Web Admin에 오케스트레이션 보기 추가 (후기 단계): 노드 목록·온라인 상태, 진행 중 DAG, 승인 대기 목록, 감사 검색. 1단계는 대화 내 요약 + 감사 로그 조회로 시작한다.
10. 보안 위협 모델 (요약)
노드 탈취: 한 노드의 인증이 새어도, 다른 노드 권한이 자동 확장되지 않는다. 노드별 인증 + 범위 + 로컬 정책이 방어선이다.
네트워크 공격자: Tailscale/marummesh 도달만으로는 작업이 실행되지 않는다. 3.4 검증 + mTLS/토큰이 필요하다.
프롬프트 인젝션: worker 도구 출력, 웹 페이지, 파일 내용은 모두 untrusted로 취급한다. 오케스트레이터는 그 안의 지시를 따르지 않는다.
재생 공격: 작업마다 만료 시간과 nonce(idempotency_key + 타임스탬프)를 두고, 만료된 작업을 거부한다.
오케스트레이터 장애: 오케스트레이터가 멈추면 신규 조율은 멈추지만, 각 worker의 로컬 cron/heartbeat 동작은 독립적으로 유지된다. 오케스트레이터를 단일 장애점으로 만들지 않기 위해, 오케스트레이터 자체도 main 서버에서 서비스로 상시 실행한다.
11. 단계별 로드맵
Phase 0 — 설계 확정 (본 문서)
오케스트레이터 호스트, 인증 형태(mTLS vs 토큰), worker API 포트, heartbeat 간격을 확정한다.
marubot 실행 시 Google SSO 진입점을 코드 수준에서 확인한다 (Web Admin 진입인지, 별도 로그인 단계인지).
Phase 1 — Tailscale 최소 동작
Transport 인터페이스 + TailscaleTransport 구현.
등록·heartbeat·능력 목록 구현.
단일 subtask 위임(오케스트레이터 → worker 1대) + 결과 반환 + 감사 로그.
테스트: 오프라인 노드 제외, 소유자 불일치 거부, 동일 idempotency_key 중복 실행 방지.
Phase 2 — DAG·승인·집계
DAG 분해, 병렬·순차 실행, 재라우팅, 승인 취합, 최종 합성.
로컬 정책 게이트를 ToolRegistry.Execute() 앞단에 적용.
Web Admin 오케스트레이션 보기(최소: 노드 상태, 진행 중 작업, 감사 검색).
Phase 3 — marumesh 전송 교체 (marumesh 완성 후)
MaruMeshTransport 구현, nodeID ↔ device_id 매핑, 피어 목록 접점 확정.
노드 1대 카나리 → 순차 전환 → 롤백 절차 검증.
TailscaleTransport는 최소 1개 릴리스 동안 유지한다.
Phase 4 — 고도화 (필요 시)
결과 캐싱, 대용량 결과 청크 전송, 노드 부하 기반 동적 라우팅, 오케스트레이터 이중화 검토.
12. 구현 전 미확인 사항
marubot 실행 시 Google SSO 로그인의 정확한 진입점과 토큰 저장 위치 (코드 수준 확인 필요).
오케스트레이터 호스트 확정 (권장: main, 사용자 최종 확인 필요).
인증 형태 확정 (mTLS 인증서 vs 노드별 API 토큰).
worker API 포트와 heartbeat 간격, 인증서 유효 기간.
marumesh 피어 목록 접점 (ctl list/status, 대시보드 API, 로컬 소켓 중 확정) — Phase 3에서 확인.
marumesh 완성 기준: 사용자가 "완성"으로 판단하는 릴리스·기능 범위.
13. 참고: 관련 프로젝트 위치
marubot 정본: private dirmich/maru-bot (Go 1.25, 모듈 github.com/dirmich/marubot, v0.9.13). 로컬: window-office d:\work\0.project\0.ai\maru-bot.
marumesh 정본: private dirmich/marumesh (Go 1.26.1). public dirmich/maru-mesh는 배포·문서용.
두 프로젝트는 Go 버전이 다르지만, 네트워크로 통신하므로 같은 바이너리로 묶을 필요가 없다. 각자 빌드·배포한다.
14. 공용 메모리와 위임 명령: LLM wiki 활용 (2026-10-28 제안, 확정 전)
14.1 역할 분리: wiki는 지식, 태스크 큐는 아니다
LLM wiki(knowledge_base, Obsidian vault + git)는 오케스트레이터·worker의 공용 장기 메모리로 쓴다.
wiki에 넣는 것: 위임 명령 정의(명령 카탈로그·런북), 노드 프로필, 재사용 가능한 결정, 시행착오, 환경 이슈, 검증 기준.
wiki에 넣지 않는 것: 실시간 태스크 큐·상태, heartbeat, 승인 토큰, 비밀·토큰·.env 값, 대용량 로그·출력 원본. 이들은 6절 태스크 봉투, 9절 SQLite 감사, 전송 계층에 둔다.
이 구분이 없으면 git 충돌과 동기화 지연이 태스크 실행을 막는다. wiki는 느린 공유 지식, 태스크 버스는 빠른 실행 상태이다.
14.2 wiki 구조 (기존 규칙 준수)
wiki 운영 규칙(LLM wiki 운영 규칙.md)을 그대로 따른다: 세션 시작 시 git status + git pull --rebase, 운영 규칙 + Codex/전역 프로젝트 운영 규칙.md + 관련 프로젝트 wiki를 먼저 읽는다. 세션 종료 시 한글 커밋 메시지 + push, 코드 저장소와 wiki 저장소는 분리 커밋한다. 비밀을 기록하지 않고, 모지바케 파일을 그대로 커밋하지 않으며, 충돌 시 양쪽을 보존하고 강제 푸시를 하지 않는다.
기존 관례(Codex/<프로젝트명>/)에 따라 marubot 공간은 Codex/maru-bot/을 쓴다. 2026-10-13 기준 이 폴더는 아직 없었다.
제안 구조:
Codex/maru-bot/Commands/<command-id>.md — 위임 명령 정의
Codex/maru-bot/Nodes/<node-id>.md — 노드 프로필(플랫폼, 도구, 역할, 제약)
Codex/maru-bot/Decisions/<date>-<topic>.md — 재사용 결정
Codex/maru-bot/Runbooks/<task-type>.md — 반복 작업 절차와 검증 기준
Codex/maru-bot/Memory/<topic>.md — 검증된 공용 사실과 시행착오
vault 경로: Windows D:\work\knowledge_base, macOS /Users/dirmich/work/0.highmaru/knowledge_base, 원격 https://github.com/dirmich/knowledge_base.git.
14.3 위임 명령 정의 형식
오케스트레이터는 자유 문장만으로 worker에 일을 던지지 않고, wiki 명령 정의를 참조한다.
명령 페이지 필수 항목: id, scope, 필요 능력, approval_level, 인자 스키마, 제한 시간, 재시도 횟수, 기대 출력 형식, 안전 주의사항, last_verified(날짜·노드·marubot 버전).
태스크 봉투에는 명령 전체를 복사하지 않고 command_id + wiki_commit_sha + 필요한 발췌를 넣는다. worker는 실행 전 해당 정의를 확인하고, 정의가 없거나 만료됐으면 멈추고 오케스트레이터에 묻는다.
wiki 내용과 최신 사용자 지시·실제 코드 상태가 충돌하면, wiki 운영 규칙대로 최신 지시와 코드 상태를 우선하고 wiki는 갱신 대상으로 본다.
14.4 공용 메모리 읽기·쓰기 흐름
읽기 (세션 시작): 오케스트레이터·worker는 작업 시작 전에 wiki를 pull --rebase로 동기화하고, 관련 명령·노드·런북을 먼저 읽는다. 긴 DAG에서는 단계 경계와 파괴적 실행 전에 다시 동기화하고, 사용한 wiki 커밋 SHA를 태스크 provenance에 남긴다.
쓰기 (단일 작성자 권장): 모든 노드가 직접 wiki에 push하면 충돌이 잦아진다. 따라서 worker는 작업 결과 중 재사용할 만한 내용(결정, 시행착오, 환경 이슈)을 한국어로 초안만 만들어 태스크 결과로 오케스트레이터에 보낸다. 오케스트레이터(또는 vault가 있는 지정 작성 노드)가 pull --rebase 후 한글 커밋으로 push한다.
vault가 없는 노드(RPi 등)는 git에 직접 쓰지 않고, 위 초안 전달 방식만 쓴다.
검색: 1단계는 파일명·frontmatter·전문 검색(키워드)이다. 임베딩·벡터 검색은 후기 단계에서 검토한다.
14.5 범위 주의: maru-agent 결정과는 별개이다
2026-10-13 결정은 maru-agent에 대한 것이다: knowledge_base를 직접 연결·동기화하지 않고 구조만 참고하며, maru-agent는 자체 내장 wiki를 갖는다.
본 절은 marubot 오케스트레이터에 대한 제안이다. maru-agent의 결정을 바꾸는 것이 아니며, maru-agent까지 wiki 공유로 바꾸려면 별도 결정이 필요하다.
15. 와이어 프로토콜: HTTP+JSON+SSE (2026-10-28 확정)
15.1 결정
에이전트 간 와이어 프로토콜은 HTTP+JSON+SSE로 확정한다. gRPC는 쓰지 않는다.
이유: Go 표준 라이브러리만으로 구현 가능하고, 단일 바이너리 철학과 맞으며, 기존 Web Admin·HTTP 구조와 같은 스택을 쓴다. 디버깅도 curl·브라우저로 바로 된다. SSE는 단방향 스트리밍에 충분하고, 양방향 제어가 필요하면 별도 HTTP 요청으로 처리한다.
15.2 전송과 와이어의 구분
4절 Transport는 "어떻게 닿는지"(Tailscale vs marumesh)이고, 본 절은 "닿은 뒤 무슨 말을 주고받는지"이다.
전송을 바꿔도 와이어 형식은 바뀌지 않는다. Tailscale IP로 가든 marumesh 가상 IP·.maru 이름으로 가든, 아래 엔드포인트와 이벤트 형식이 동일하다.
15.3 엔드포인트 (제안, 기존 API가 아님)
POST /v1/tasks — 태스크 제출. 요청 본문은 6.1절 태스크 봉투(JSON). 응답은 202 Accepted + {task_id, status: "queued"}. 멱등 키는 헤더 Idempotency-Key와 본문 idempotency_key를 함께 쓴다.
GET /v1/tasks/{task_id}/events — SSE 스트림. 연결이 끊기면 Last-Event-ID로 재연결하고, 서버는 마지막으로 보낸 이벤트 이후부터 다시 보낸다.
GET /v1/tasks/{task_id} — 현재 상태와 최종 결과 조회 (폴링용 보조 수단, 주 경로는 SSE).
POST /v1/tasks/{task_id}/cancel — 취소 요청. worker는 안전 지점에서 중단하고 cancelled를 반환한다.
POST /v1/tasks/{task_id}/approval — 오케스트레이터가 취합한 승인 결과를 worker에 전달한다.
POST /v1/nodes/heartbeat — worker의 heartbeat + NodeInfo 갱신.
GET /v1/nodes — 오케스트레이터의 노드·능력 목록 조회 (오케스트레이터 전용, worker는 호출하지 않는다).
모든 엔드포인트는 Tailscale/marummesh 사설망에서만 수신한다. 공인 IP 바인딩은 하지 않는다.
15.4 SSE 이벤트 형식
event: task.started
data: {"task_id": "task-20261028-001", "node_id": "window-office", "ts": "2026-10-28T12:00:00+09:00"}

event: task.progress
data: {"task_id": "task-20261028-001", "progress": 0.4, "message": "브라우저 조사 중"}

event: task.tool_call
data: {"task_id": "task-20261028-001", "tool": "browser", "target": "https://example.com"}

event: task.need_approval
data: {"task_id": "task-20261028-001", "approval_level": "destructive", "summary": "파일 삭제 요청"}

event: task.succeeded
data: {"task_id": "task-20261028-001", "output": "...", "provenance": {"node_id": "window-office"}}

event: task.failed
data: {"task_id": "task-20261028-001", "error": {"code": "tool_denied", "message": "..."}}
이벤트마다 id를 붙인다. 재연결 시 Last-Event-ID 이후부터 재전송한다.
진행 이벤트는 너무 자주 보내지 않는다. 도구 호출 단위와 단계 전환 단위로 제한하고, 원시 로그를 그대로 스트리밍하지 않는다.
최종 결과는 SSE task.succeeded/task.failed와 GET /v1/tasks/{task_id} 조회 결과가 동일해야 한다. 둘 중 하나만 믿는 구조로 만들지 않는다.
15.5 보안과 오류 처리 (와이어 수준)
평문 HTTP는 쓰지 않는다. 사설망 안에서도 TLS를 쓴다. 인증은 3.3절에서 확정할 노드 인증(mTLS 또는 노드별 토큰)을 HTTP 계층에 적용한다. 예: mTLS이면 클라이언트 인증서, 토큰이면 Authorization: Bearer <node-token> 헤더. 둘 중 하나를 구현 전 확정한다.
오류 형식은 통일한다: {error: {code, message, retryable}}. retryable=false인 오류(소유자 불일치, 정책 거부, 승인 거부)는 재시도하지 않는다.
요청 본문 크기 제한을 둔다. 대용량 결과는 SSE에 통째로 실지 않고, 요약 + 청크 조회(후기 단계)로 나눈다.
타임아웃: 연결 타임아웃, SSE 무응답 타임아웃, 태스크 deadline을 각각 둔다. SSE가 끊겼다고 태스크가 실패한 것은 아니다. 재연결 또는 상태 조회로 확인한 뒤에만 실패·재라우팅을 판단한다.
15.6 아직 확정되지 않은 것 (와이어와 별개)
worker API 포트 번호, heartbeat 간격, 인증서 유효 기간 (12절 4항과 동일).
인증 형태 최종 선택: mTLS vs 노드별 API 토큰 (3.3절, 12절 3항과 동일).
SSE 재전송 보관 기간 (이벤트 버퍼를 얼마나 오래 둘지).