# 환경 설정

## 구성

Admin backend와 frontend는 `go run ./admin/backend` 또는 빌드된 maru-admin 바이너리 하나로 실행한다. backend가 frontend를 임베드해 제공한다. 기존 각 장비의 로컬 Web Admin은 그대로 유지된다. 신규 전체 시스템 관리는 별도 Admin 주소에서 접속한다.

저장소 루트 `.env.example`을 `.env`로 복사한 뒤 설정한다. Admin은 루트 `.env`를 읽되 이미 설정된 OS 환경변수를 우선한다. worker와 기존 marubot agent는 OS 환경변수를 사용한다. 실제 `.env`, 토큰, 인증서 개인키와 DB는 Git에 넣지 않는다.

## Admin 필수값

| 변수 | 의미 | 안전한 예시 |
|---|---|---|
| MARUBOT_ADMIN_ADDR | 수신 주소 | 127.0.0.1:8000 |
| MARUBOT_ADMIN_PUBLIC_URL | 브라우저 기준 주소, CSRF Origin 및 OAuth callback 기준 | https://admin.example.invalid |
| MARUBOT_ADMIN_DATA | 사용자별 DB와 암호화 키 저장 경로 | admin-data |
| GOOGLE_CLIENT_ID | Google 웹 OAuth client ID | Google 콘솔에서 발급 |
| GOOGLE_CLIENT_SECRET | Google 웹 OAuth secret, backend 전용 | 비밀 관리 환경에서 주입 |

Google redirect URI는 정확히 `<MARUBOT_ADMIN_PUBLIC_URL>/auth/google/callback`으로 등록한다. Google client 값이 없으면 로그인 UI는 설정 필요 상태를 표시하고, 데이터 API는 인증 없이 사용할 수 없다. OAuth state·PKCE·nonce와 ID token의 issuer·audience·sub·만료 및 RS256 서명을 확인한다. 공개키 검증 실패 시 로그인은 실패한다.

외부로 직접 수신할 때 `MARUBOT_ADMIN_CERT`, `MARUBOT_ADMIN_KEY`를 설정한다. 없으면 loopback에만 바인딩할 수 있다. reverse proxy를 쓸 때 backend는 loopback HTTP, 브라우저 기준 공개 URL은 HTTPS다. 기본값의 localhost HTTP는 로컬 개발용이다. 실제 worker는 Admin HTTPS 주소에 연결해야 한다.

`MARUBOT_ADMIN_LLM_CONFIG`는 자연어 계획을 생성할 기존 marubot JSON 설정 파일 경로다. provider와 모델, API 키는 그 파일에서 읽으며 frontend에 보내지 않는다. 설정하지 않으면 직접 단계 작성은 가능하고 자연어 계획은 설정 필요 오류를 반환한다.

## worker 필수값

| 변수 | 의미 |
|---|---|
| MARUBOT_CONTROL_URL | Tailscale에서 접근 가능한 Admin HTTPS 기준 URL, 경로 없이 지정 |
| MARUBOT_OWNER_ID | 로그인한 사용자에게 발급된 소유자 ID |
| MARUBOT_NODE_ID | 등록된 장비 ID |
| MARUBOT_NODE_TOKEN | 발급된 장비 토큰, 원문을 문서에 쓰지 않음 |
| MARUBOT_WORKER_ADDR | 실제 Tailscale IP:8443 또는 loopback IP:8443 |
| MARUBOT_WORKER_CERT / KEY | 해당 장비 HTTPS 이름에 유효한 인증서와 키 |
| MARUBOT_WORKER_TOOLS | 허용할 실제 도구 이름의 쉼표 목록, 예: system_status,read_file,list_dir |
| MARUBOT_CONTROL_CA | 자체 CA 사용 시 신뢰할 PEM 경로, 생략하면 시스템 CA |

장비는 인증 성공 heartbeat 이후에만 원격 도구를 실행하며, 인증을 45초 이상 확인하지 못하면 거부한다. heartbeat는 15초마다 전송한다. 장비 토큰은 90일 만료하며 Admin에서 개별 갱신·폐기한다. 갱신 시 장비 환경도 교체하고 worker를 재시작한다.

원격 파일 도구는 설정된 marubot workspace 내부만 접근한다. 인증 자료 경로와 workspace 밖의 symlink를 거부한다. shell을 허용하면 장비 OS 권한으로 실행하므로 OS 계정의 권한도 제한한다. shell은 읽기 전용으로 선언해도 개별 승인이 필요하다. 재위임·SSH·동적 도구 생성 등은 원격 정책에서 차단한다.

## 대화에서 위임하기

Admin에서 역할을 `orchestrator`로 등록한 장비의 `MARUBOT_CONTROL_URL`, `MARUBOT_NODE_TOKEN`, 선택 CA를 기존 `marubot agent` 또는 서비스 환경에 설정한다. 이때 `orchestration_status`, `orchestration_submit`, `shared_memory` 도구가 추가된다. LLM은 실제 장비와 도구를 조회한 후 DAG를 제출하며 사용자에게 run ID와 결과를 보고한다. 승인은 Admin 로그인 사용자만 수행한다.

## 주소 구분과 검증

- localhost는 현재 프로세스가 실행되는 장비다. 다른 worker에서 Admin의 localhost 주소를 사용하지 않는다.
- 현재 구현은 호스트 바이너리 실행을 기준으로 한다. Docker 내부 localhost는 호스트가 아니다. 사설망 worker 연결과 인증서 이름을 별도로 맞추기 전 컨테이너 배포로 바꾸지 않는다.
- HTTPS 노드 연결은 loopback/Tailscale IPv4·IPv6만 허용한다. DNS 응답의 IP를 확인한 뒤 그 IP로 연결해 주소 재해석에 의한 우회를 방지한다.
- 잘못된 인증서·소유자·노드 ID·폐기된 토큰은 접속 실패로 처리한다. TLS 검증을 끄는 옵션은 없다.

인증 검증 기준: [Google OpenID Connect](https://developers.google.com/identity/openid-connect/openid-connect), [backend ID token 검증](https://developers.google.com/identity/sign-in/web/backend-auth). Google 공개키(JWKS)를 캐시하고 RS256 서명을 backend에서 검증한다.
